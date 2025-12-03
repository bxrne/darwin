"""
Entry point for the Generals game bridge server (multiprocess version).
"""

from src.bridge import Bridge


def main():
    """Start the bridge server."""
    bridge = Bridge(address="127.0.0.1", port=5000, render_mode=None)
    bridge.start()

    try:
        while True:
            stats = bridge.get_stats()
            print(f"\nStatus: {stats['active_clients']} clients, {stats['active_workers']} active workers (games)")

    except KeyboardInterrupt:
        print("\nShutting down...")
        bridge.stop()
        print("Server stopped")


if __name__ == "__main__":
    main()
