
import sys
from generals import Replay


def main():
    if len(sys.argv) < 2:
        print("Usage: python play_replay.py <replay_file.pkl>")
        sys.exit(1)

    filepath = sys.argv[1]
    replay = Replay.load(filepath)
    replay.play()


if __name__ == "__main__":
    main()
