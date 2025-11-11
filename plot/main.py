import os
import argparse 

import pandas as pd
import numpy as np 
import seaborn as sns

def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Plotting Script")
    parser.add_argument(
        "--csv",
        type=str,
        default="../default_metrics.csv",
        help="Path to the CSV file containing metrics data."
    )
    return parser.parse_args()

def main(args: argparse.Namespace) -> None:
    if not os.path.exists(args.csv):
        raise FileNotFoundError(f"CSV file not found: {args.csv}")

    data = pd.read_csv(args.csv)


if __name__ == "__main__":
    main(parse_args())
