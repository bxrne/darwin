"""
Entry point for the Generals game bridge server (multiprocess version).
"""

import time
import logging
from src.bridge import Bridge


def main():
    """Start the bridge server."""
    logging.basicConfig(
        level=logging.WARNING, format="%(asctime)s - %(levelname)s - %(message)s"
    )
    bridge = Bridge(address="127.0.0.1", port=5000, render_mode=None)
    bridge.start()

    try:
        while True:
            time.sleep(10)  # Just keep main thread alive

    except KeyboardInterrupt:
        logging.warning("Shutting down...")
        bridge.stop()
        logging.warning("Server stopped")


if __name__ == "__main__":
    main()
