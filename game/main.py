"""
Entry point for the Generals game bridge server.
"""

import threading
from src.bridge import Bridge


def main():
    """Start the bridge server."""
    bridge = Bridge(
        address="0.0.0.0",  # Listen on all interfaces
        port=5000,
        render_mode=None  # Set to "human" for visual rendering
    )

    bridge.start()

    try:
        while True:
            threading.Event().wait(5)
            stats = bridge.get_stats()
            print(f"\nStatus: {stats['active_clients']} clients, {
                  stats['active_games']} games")

    except KeyboardInterrupt:
        print("\nShutting down...")
        bridge.stop()
        print("Server stopped")


if __name__ == "__main__":
    main()
