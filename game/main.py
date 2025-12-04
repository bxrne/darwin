"""
Entry point for the Generals game bridge server (multiprocess version).
"""

import time
from src.bridge import Bridge


def main():
    """Start the bridge server."""
    bridge = Bridge(address="127.0.0.1", port=5000, render_mode=None)
    bridge.start()

    try:
        counter = 0
        while True:
            stats = bridge.get_stats()
            # Only print if there's activity or every 30 seconds
            if (
                stats["active_clients"] > 0
                or stats["active_workers"] > 0
                or counter % 6 == 0
            ):
                print(
                    f"\nStatus: {stats['active_clients']} clients, {stats['active_workers']} active workers (games)"
                )
            time.sleep(5)  # Check every 5 seconds, print every 30 seconds when idle
            counter += 1

    except KeyboardInterrupt:
        print("\nShutting down...")
        bridge.stop()
        print("Server stopped")


if __name__ == "__main__":
    main()
