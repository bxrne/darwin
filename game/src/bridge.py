from src.payloads import (
    MessageType,
    ConnectedResponse,
    ObservationResponse,
    ErrorResponse,
    GameOverResponse,
    to_dict,
    NumpyEncoder,
)
from src.game import Game
from typing import Dict, Optional
import logging
from logging.handlers import QueueHandler, QueueListener
import json
import threading
import multiprocessing
import socket

# ---------------------------------------------------
# Worker process
# ---------------------------------------------------


def worker_process(
    client_id: str,
    client_socket_fd: int,
    log_queue: multiprocessing.Queue,
    disconnect_queue: multiprocessing.Queue,
    render_mode=None,
):
    """
    Handles a single client connection in a separate process.
    """
    import socket

    client_socket = socket.socket(fileno=client_socket_fd)
    client_socket.setblocking(True)

    # Setup logger
    logger = logging.getLogger(f"Game-{client_id}")
    qh = QueueHandler(log_queue)
    logger.addHandler(qh)
    logger.setLevel(logging.INFO)
    logger.propagate = False

    buffer = ""
    game: Optional[Game] = None
    try:
        while True:
            data = client_socket.recv(4096).decode("utf-8")
            if not data:
                logger.info(f"Client {client_id} disconnected (reader)")
                break
            buffer += data
            while "\n" in buffer:
                message, buffer = buffer.split("\n", 1)
                if not message.strip():
                    continue
                try:
                    json_data = json.loads(message)
                    msg_type = json_data.get("type")

                    if msg_type == MessageType.CONNECT:

                        opponent_type = json_data.get(
                            "opponent_type", "random")

                        game = Game(
                            client_id,
                            opponent_type=opponent_type,
                            render_mode=render_mode,
                        )
                        observation, info = game.reset()

                        print("GOT HERE")
                        # Connected response
                        response = ConnectedResponse(
                            agent_id=client_id,
                            opponent_id=game.opponent.id,
                            message=f"Connected! Playing as {client_id} vs {
                                game.opponent.id
                            }",
                        )
                        client_socket.sendall(
                            (
                                json.dumps(to_dict(response),
                                           cls=NumpyEncoder) + "\n"
                            ).encode("utf-8")
                        )

                        # Initial observation
                        obs_response = ObservationResponse(
                            observation=observation,
                            reward=0.0,
                            terminated=False,
                            truncated=False,
                            info=info,
                        )
                        client_socket.sendall(
                            (
                                json.dumps(to_dict(obs_response),
                                           cls=NumpyEncoder)
                                + "\n"
                            ).encode("utf-8")
                        )

                        logger.info(f"Game created: {client_id} vs {
                                    game.opponent.id}")
                        logger.info(f"Game reset")

                    elif msg_type == MessageType.ACTION:
                        if not game:
                            err = ErrorResponse(
                                "No active game", "Send CONNECT first")
                            client_socket.sendall(
                                (
                                    json.dumps(
                                        to_dict(err), cls=NumpyEncoder) + "\n"
                                ).encode("utf-8")
                            )
                            continue
                        action = json_data.get("action")
                        result = game.step(action)
                        reward = result["reward"]

                        # Log reward before sending
                        if reward != 0:
                            logger.info(f"Sending non-zero reward: {reward}")

                        # Observation response
                        response = ObservationResponse(
                            observation=result["observation"],
                            reward=reward,
                            terminated=result["terminated"],
                            truncated=result["truncated"],
                            info=result["info"],
                        )
                        client_socket.sendall(
                            (
                                json.dumps(to_dict(response),
                                           cls=NumpyEncoder) + "\n"
                            ).encode("utf-8")
                        )

                        # Game over
                        if result["terminated"] or result["truncated"]:
                            game_over = GameOverResponse(
                                winner=game.get_winner(),
                                final_rewards={
                                    game.client_id: result["reward"]},
                                reason="Game completed",
                            )
                            client_socket.sendall(
                                (
                                    json.dumps(to_dict(game_over),
                                               cls=NumpyEncoder)
                                    + "\n"
                                ).encode("utf-8")
                            )

                    elif msg_type == MessageType.RESET:
                        if not game:
                            err = ErrorResponse(
                                "No active game", "Send CONNECT first")
                            client_socket.sendall(
                                (
                                    json.dumps(
                                        to_dict(err), cls=NumpyEncoder) + "\n"
                                ).encode("utf-8")
                            )
                            continue
                        observation, info = game.reset()
                        response = ObservationResponse(
                            observation=observation,
                            reward=0.0,
                            terminated=False,
                            truncated=False,
                            info=info,
                        )
                        client_socket.sendall(
                            (
                                json.dumps(to_dict(response),
                                           cls=NumpyEncoder) + "\n"
                            ).encode("utf-8")
                        )

                    else:
                        err = ErrorResponse(
                            "Unknown message type", f"Type: {msg_type}")
                        client_socket.sendall(
                            (json.dumps(to_dict(err), cls=NumpyEncoder) + "\n").encode(
                                "utf-8"
                            )
                        )

                except Exception as e:
                    err = ErrorResponse("Processing error", str(e))
                    client_socket.sendall(
                        (json.dumps(to_dict(err), cls=NumpyEncoder) + "\n").encode(
                            "utf-8"
                        )
                    )
                    logger.error(f"Error processing message: {e}")

    except ConnectionResetError:
        logger.info(f"Connection reset by {client_id}")
    except Exception as e:
        logger.error(f"Worker error: {e}")
    finally:
        try:
            if game:
                game.close()
            client_socket.close()
        except Exception:
            pass
        # Notify main process about disconnect
        disconnect_queue.put({"client_id": client_id})
        logger.info(f"Worker for {client_id} exited")


# ---------------------------------------------------
# Main Bridge
# ---------------------------------------------------
class Bridge:
    def __init__(self, address="0.0.0.0", port=5000, backlog=50, render_mode=None):
        self.address = address
        self.port = port
        self.backlog = backlog
        self.render_mode = render_mode

        self.server_socket: Optional[socket.socket] = None
        self.clients: Dict[str, socket.socket] = {}
        self.workers: Dict[str, multiprocessing.Process] = {}

        self.log_queue: multiprocessing.Queue = multiprocessing.Queue()
        self.disconnect_queue: multiprocessing.Queue = multiprocessing.Queue()

        self.running = False
        self.lock = threading.Lock()
        self.client_counter = 0

    def start(self):
        # Setup logging listener
        listener = QueueListener(self.log_queue, logging.StreamHandler())
        listener.start()

        # Start server socket
        self.server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.server_socket.setsockopt(
            socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        # Add timeout to allow clean shutdown
        self.server_socket.settimeout(1.0)
        self.server_socket.bind((self.address, self.port))
        self.server_socket.listen(self.backlog)
        self.running = True

        # Start disconnect handler
        threading.Thread(target=self._handle_disconnects, daemon=True).start()

        # Accept connections
        threading.Thread(target=self._accept_connections, daemon=True).start()
        logging.info(f"Bridge started on {self.address}:{self.port}")

    def _accept_connections(self):
        while self.running:
            try:
                if self.server_socket is None:
                    break
                client_socket, client_address = self.server_socket.accept()
                with self.lock:
                    self.client_counter += 1
                    client_id = f"client_{self.client_counter}"
                    self.clients[client_id] = client_socket

                # Spawn worker process
                p = multiprocessing.Process(
                    target=worker_process,
                    args=(
                        client_id,
                        client_socket.fileno(),
                        self.log_queue,
                        self.disconnect_queue,
                        self.render_mode,
                    ),
                    daemon=False,
                )
                p.start()
                self.workers[client_id] = p
                logging.info(
                    f"Accepted connection from {client_address} as {client_id}"
                )

            except socket.timeout:
                # Timeout is expected, allows checking self.running
                continue
            except Exception as e:
                if self.running:  # Only log errors if we're still running
                    logging.error(f"Error accepting connection: {e}")

    def _handle_disconnects(self):
        """Clean up workers when they notify disconnection."""
        while self.running:
            try:
                msg = self.disconnect_queue.get(timeout=1)
                client_id = msg["client_id"]
                with self.lock:
                    # remove worker
                    p = self.workers.pop(client_id, None)
                    if p and p.is_alive():
                        p.terminate()
                        p.join(timeout=5)  # Wait for graceful shutdown
                    # close socket
                    sock = self.clients.pop(client_id, None)
                    if sock:
                        try:
                            sock.close()
                        except:
                            pass
                logging.info(f"Cleaned up client {client_id} (disconnect)")
            except Exception:
                continue

    def stop(self):
        logging.info("Stopping bridge...")
        self.running = False
        # Close all client sockets
        with self.lock:
            for sock in self.clients.values():
                try:
                    sock.close()
                except:
                    pass
            self.clients.clear()
            # Terminate workers and wait for them to exit
            for p in self.workers.values():
                if p is not None and p.is_alive():
                    p.terminate()
                    # Wait up to 5 seconds for graceful shutdown
                    p.join(timeout=5)
                    if p.is_alive():
                        # Force kill if still alive
                        p.kill()
                        p.join(timeout=2)
            self.workers.clear()
        if self.server_socket:
            try:
                self.server_socket.close()
            except:
                pass
        logging.info("Bridge stopped")

    def get_stats(self):
        """Return current number of active clients and workers."""
        with self.lock:
            # Remove dead workers automatically
            dead_workers = [cid for cid,
                            p in self.workers.items() if not p.is_alive()]
            for cid in dead_workers:
                self.workers.pop(cid, None)
                self.clients.pop(cid, None)
            return {
                "active_clients": len(self.clients),
                "active_workers": len(self.workers),
            }
