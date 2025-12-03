
"""
Entry point for the Generals game bridge server (multiprocess version).
"""

import threading
# <-- import updated bridge
from src.bridge import Bridge


def main():
    """Start the bridge server."""
    bridge = Bridge(
        address="0.0.0.0",
        port=5000,
        render_mode=None
    )

    bridge.start()

    try:
        while True:
            threading.Event().wait(5)
            stats = bridge.get_stats()
            print(f"\nStatus: {stats['active_clients']} clients, "
                  f"{stats['active_workers']} active workers (games)")
    except KeyboardInterrupt:
        print("\nShutting down...")
        bridge.stop()
        print("Server stopped")


if __name__ == "__main__":
    main()
