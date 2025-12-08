"""
Entry point for the Generals game bridge server (multiprocess version).
"""

import logging
from src.bridge import Bridge


def main():
    """Start bridge server."""
    # Setup logging with proper formatting
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s | %(levelname)8s | %(name)s | %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )

    bridge = Bridge(address="127.0.0.1", port=5000, render_mode=None)
    bridge.start()

    try:
        old_wins = 0
        while True:
            stats = bridge.get_stats()
            if stats["global_wins"] != old_wins:
                old_wins = stats["global_wins"]
                logging.info(f"Stats: {stats}")




    except KeyboardInterrupt:
        logging.info("Shutting down...")
        bridge.stop()
        logging.info("Server stopped")


if __name__ == "__main__":
    main()
