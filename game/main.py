"""
Entry point for the Generals game bridge server (multiprocess version).
"""

import time
import logging
from src.bridge import Bridge


def main():
    """Start the bridge server."""
    logging.basicConfig(
        level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s"
    )
    bridge = Bridge(address="127.0.0.1", port=5000, render_mode=None)
    bridge.start()

    try:
        counter = 0
        while True:
            stats = bridge.get_stats()
            # Only log if there's activity or every 30 seconds
            if (
                stats["active_clients"] > 0
                or stats["active_workers"] > 0
                or counter % 6 == 0
            ):
                logging.info(
                    "Status: %d clients, %d active workers (games)",
                    stats["active_clients"],
                    stats["active_workers"],
                )
            time.sleep(5)  # Check every 5 seconds, log every 30 seconds when idle
            counter += 1

    except KeyboardInterrupt:
        logging.info("Shutting down...")
        bridge.stop()
        logging.info("Server stopped")


if __name__ == "__main__":
    main()
