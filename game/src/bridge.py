"""
Socket server bridge that manages client connections and game instances.
"""

import socket
import threading
import json
import logging
from typing import Dict, Optional
from src.game import Game
from src.payloads import (
    MessageType, ConnectRequest, ConnectedResponse, ActionRequest,
    ObservationResponse, ResetRequest, ErrorResponse, GameOverResponse,
    to_dict, NumpyEncoder
)

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)


class Bridge:
    """Socket server that manages game instances for connected clients."""

    def __init__(self, address="localhost", port=5000, backlog=5, render_mode=None):
        """
        Initialize the bridge server.

        Args:
            address: Server bind address
            port: Server bind port
            backlog: Maximum number of queued connections
            render_mode: Optional render mode for games
        """
        self.address = address
        self.port = port
        self.backlog = backlog
        self.render_mode = render_mode

        self.server_socket: Optional[socket.socket] = None
        self.clients: Dict[str, socket.socket] = {}  # client_id -> socket
        self.games: Dict[str, Game] = {}  # client_id -> Game
        self.running = False
        self.lock = threading.Lock()
        self.logger = logging.getLogger(self.__class__.__name__)
        self.client_counter = 0

    def start(self):
        """Start the bridge server."""
        try:
            self.server_socket = socket.socket(
                socket.AF_INET, socket.SOCK_STREAM)
            self.server_socket.setsockopt(
                socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            self.server_socket.bind((self.address, self.port))
            self.server_socket.listen(self.backlog)
            self.running = True

            self.logger.info(f"Bridge started on {self.address}:{self.port}")

            # Accept connections
            accept_thread = threading.Thread(
                target=self._accept_connections, daemon=True)
            accept_thread.start()

        except Exception as e:
            self.logger.error(f"Failed to start bridge: {e}")
            raise

    def _accept_connections(self):
        """Accept incoming client connections."""
        while self.running:
            try:
                client_socket, client_address = self.server_socket.accept()
                self.logger.info(f"New connection from {client_address}")

                # Generate client ID
                with self.lock:
                    self.client_counter += 1
                    client_id = f"client_{self.client_counter}"

                # Spawn handler thread
                client_thread = threading.Thread(
                    target=self._handle_client,
                    args=(client_socket, client_address, client_id),
                    daemon=True
                )
                client_thread.start()

            except OSError:
                if not self.running:
                    break
            except Exception as e:
                self.logger.error(f"Error accepting connection: {e}")

    def _handle_client(self, client_socket: socket.socket, client_address, client_id: str):
        """Handle a single client connection."""
        buffer = ""
        game: Optional[Game] = None

        try:
            while self.running:
                data = client_socket.recv(4096).decode('utf-8')

                if not data:
                    self.logger.info(f"Client {client_id} disconnected")
                    break

                buffer += data

                # Process complete messages
                while '\n' in buffer:
                    message, buffer = buffer.split('\n', 1)

                    if message.strip():
                        try:
                            json_data = json.loads(message)
                            response = self._process_message(
                                client_id, client_socket, json_data, game
                            )

                            # Update game reference if created
                            if response and "game" in response:
                                game = response["game"]

                        except json.JSONDecodeError as e:
                            self.logger.error(
                                f"Invalid JSON from {client_id}: {e}")
                            self._send_error(
                                client_socket, "Invalid JSON format", str(e))

        except ConnectionResetError:
            self.logger.info(f"Connection reset by {client_id}")
        except Exception as e:
            self.logger.error(f"Error handling client {client_id}: {e}")
        finally:
            self._cleanup_client(client_id, client_socket, game)

    def _process_message(
        self,
        client_id: str,
        client_socket: socket.socket,
        json_data: Dict,
        game: Optional[Game]
    ) -> Optional[Dict]:
        """Process a message from client."""
        msg_type = json_data.get("type")

        try:
            if msg_type == MessageType.CONNECT:
                return self._handle_connect(client_id, client_socket, json_data)

            elif msg_type == MessageType.ACTION:
                if not game:
                    self._send_error(
                        client_socket, "No active game", "Send CONNECT first")
                    return None
                return self._handle_action(client_socket, json_data, game)

            elif msg_type == MessageType.RESET:
                if not game:
                    self._send_error(
                        client_socket, "No active game", "Send CONNECT first")
                    return None
                return self._handle_reset(client_socket, game)

            else:
                self._send_error(
                    client_socket, "Unknown message type", f"Type: {msg_type}")
                return None

        except Exception as e:
            print(json_data)
            self.logger.error(f"Error processing message: {e}")
            self._send_error(client_socket, "Processing error", str(e))
            return None

    def _handle_connect(self, client_id: str, client_socket: socket.socket, json_data: Dict) -> Dict:
        """Handle client connection request."""
        opponent_type = json_data.get("opponent_type", "random")
        print(json_data)
        # Create game instance
        game = Game(client_id, opponent_type=opponent_type,
                    render_mode=self.render_mode)

        with self.lock:
            self.clients[client_id] = client_socket
            self.games[client_id] = game

        # Initialize game
        observation, info = game.reset()

        # Send connected response
        response = ConnectedResponse(
            agent_id=client_id,
            opponent_id=game.opponent.id,
            message=f"Connected! Playing as {client_id} vs {game.opponent.id}"
        )
        self._send_json(client_socket, to_dict(response))

        # Send initial observation
        obs_response = ObservationResponse(
            observation=observation,
            reward=0.0,
            terminated=False,
            truncated=False,
            info=info
        )
        self._send_json(client_socket, to_dict(obs_response))

        self.logger.info(f"Game created for {client_id}")
        return {"game": game}

    def _handle_action(self, client_socket: socket.socket, json_data: Dict, game: Game) -> None:
        """Handle action from client."""
        action = json_data.get("action")

        if action is None:
            self._send_error(client_socket, "Missing action",
                             "Action field is required")
            return

        # Execute game step
        result = game.step(action)

        # Send observation response
        response = ObservationResponse(
            observation=result["observation"],
            reward=result["reward"],
            terminated=result["terminated"],
            truncated=result["truncated"],
            info=result["info"]
        )
        self._send_json(client_socket, to_dict(response))

        # If game over, send game over message
        if result["terminated"] or result["truncated"]:
            game_over = GameOverResponse(
                winner=game.get_winner(),
                final_rewards={game.client_id: result["reward"]},
                reason="Game completed"
            )
            self._send_json(client_socket, to_dict(game_over))

    def _handle_reset(self, client_socket: socket.socket, game: Game) -> None:
        """Handle reset request."""
        observation, info = game.reset()

        response = ObservationResponse(
            observation=observation,
            reward=0.0,
            terminated=False,
            truncated=False,
            info=info
        )
        self._send_json(client_socket, to_dict(response))

    def _send_json(self, client_socket: socket.socket, data: Dict) -> bool:
        """Send JSON message to client."""
        try:
            message = json.dumps(data, cls=NumpyEncoder) + "\n"
            client_socket.sendall(message.encode('utf-8'))
            return True
        except Exception as e:
            self.logger.error(f"Failed to send JSON: {e}")
            return False

    def _send_error(self, client_socket: socket.socket, message: str, details: str = ""):
        """Send error message to client."""
        error = ErrorResponse(message=message, details=details)
        self._send_json(client_socket, to_dict(error))

    def _cleanup_client(self, client_id: str, client_socket: socket.socket, game: Optional[Game]):
        """Clean up client resources."""
        try:
            client_socket.close()

            with self.lock:
                if client_id in self.clients:
                    del self.clients[client_id]

                if client_id in self.games:
                    if game:
                        game.close()
                    del self.games[client_id]

            self.logger.info(f"Cleaned up client {client_id}")
        except Exception as e:
            self.logger.error(f"Error during cleanup: {e}")

    def stop(self):
        """Stop the bridge server."""
        self.logger.info("Stopping bridge...")
        self.running = False

        # Close all games and clients
        with self.lock:
            for game in self.games.values():
                game.close()
            self.games.clear()

            for client_socket in self.clients.values():
                try:
                    client_socket.close()
                except:
                    pass
            self.clients.clear()

        # Close server socket
        if self.server_socket:
            try:
                self.server_socket.close()
            except:
                pass

        self.logger.info("Bridge stopped")

    def get_stats(self) -> Dict[str, int]:
        """Get server statistics."""
        with self.lock:
            return {
                "active_clients": len(self.clients),
                "active_games": len(self.games)
            }
