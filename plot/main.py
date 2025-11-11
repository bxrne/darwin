import os
import argparse 
from pathlib import Path

import pandas as pd
import numpy as np 
import seaborn as sns
import matplotlib.pyplot as plt
import matplotlib.ticker as mticker

def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Evolution Metrics Plotting Script")
    parser.add_argument(
        "--csv",
        type=str,
        default="../default_metrics.csv",
        help="Path to the CSV file containing metrics data."
    )
    parser.add_argument(
        "--output-dir",
        type=str,
        default="./output",
        help="Directory to save generated plots."
    )
    return parser.parse_args()

def setup_plotting_style() -> None:
    """Configure matplotlib and seaborn styling."""
    plt.style.use('seaborn-v0_8-darkgrid')
    sns.set_palette("tab10")
    plt.rcParams.update({
        'figure.figsize': (12, 8),
        'font.size': 10,
        'axes.titlesize': 14,
        'axes.labelsize': 12,
        'xtick.labelsize': 10,
        'ytick.labelsize': 10,
        'legend.fontsize': 10,
        'figure.dpi': 100,
        'savefig.dpi': 300,
        'savefig.bbox': 'tight'
    })

def plot_fitness_evolution(data: pd.DataFrame, output_dir: Path) -> None:
    """Create fitness progression plot showing best/avg/min/max fitness over generations."""
    fig, ax = plt.subplots(figsize=(12, 8))
    
    # Plot fitness metrics
    ax.plot(data['generation'], data['best_fitness'], label='Best Fitness', linewidth=2, color='green')
    ax.plot(data['generation'], data['avg_fitness'], label='Average Fitness', linewidth=2, color='blue')
    ax.plot(data['generation'], data['min_fitness'], label='Min Fitness', linewidth=1, alpha=0.7, color='red')
    ax.plot(data['generation'], data['max_fitness'], label='Max Fitness', linewidth=1, alpha=0.7, color='orange')
    
    ax.set_xlabel('Generation')
    ax.set_ylabel('Fitness')
    ax.set_title('Fitness Evolution Over Generations')
    ax.legend()
    ax.grid(True, alpha=0.3)
    
    # Format y-axis to show scientific notation for better readability
    ax.yaxis.set_major_formatter(mticker.ScalarFormatter(useMathText=True))
    ax.ticklabel_format(style='scientific', axis='y', scilimits=(0,0))
    
    plt.tight_layout()
    plt.savefig(output_dir / 'fitness_evolution.png')
    plt.close()

def plot_complexity_evolution(data: pd.DataFrame, output_dir: Path) -> None:
    """Create tree complexity evolution plot showing depth statistics over generations."""
    fig, ax = plt.subplots(figsize=(12, 8))
    
    # Plot depth metrics
    ax.plot(data['generation'], data['min_depth'], label='Min Depth', linewidth=2, color='lightblue')
    ax.plot(data['generation'], data['avg_depth'], label='Average Depth', linewidth=2, color='blue')
    ax.plot(data['generation'], data['max_depth'], label='Max Depth', linewidth=2, color='darkblue')
    
    # Fill area between min and max depth
    ax.fill_between(data['generation'], data['min_depth'], data['max_depth'], 
                   alpha=0.2, color='blue', label='Depth Range')
    
    ax.set_xlabel('Generation')
    ax.set_ylabel('Tree Depth')
    ax.set_title('Tree Complexity Evolution')
    ax.legend()
    ax.grid(True, alpha=0.3)
    ax.set_ylim(bottom=0)
    
    plt.tight_layout()
    plt.savefig(output_dir / 'complexity_evolution.png')
    plt.close()

def plot_performance_metrics(data: pd.DataFrame, output_dir: Path) -> None:
    """Create performance metrics plot showing generation duration and efficiency trends."""
    fig, (ax1, ax2) = plt.subplots(2, 1, figsize=(12, 10))
    
    # Convert duration to milliseconds for better readability
    duration_ms = data['duration_ns'] / 1_000_000
    
    # Plot 1: Generation Duration
    ax1.plot(data['generation'], duration_ms, linewidth=2, color='purple')
    ax1.set_ylabel('Duration (ms)')
    ax1.set_title('Generation Performance Over Time')
    ax1.grid(True, alpha=0.3)
    
    # Add moving average for trend
    window_size = min(5, len(duration_ms))
    if window_size > 1:
        moving_avg = duration_ms.rolling(window=window_size, center=True).mean()
        ax1.plot(data['generation'], moving_avg, linewidth=2, alpha=0.7, 
                color='red', label=f'{window_size}-gen Moving Avg')
        ax1.legend()
    
    # Plot 2: Performance Distribution
    ax2.hist(duration_ms, bins=20, alpha=0.7, color='purple', edgecolor='black')
    ax2.set_xlabel('Duration (ms)')
    ax2.set_ylabel('Frequency')
    ax2.set_title('Generation Duration Distribution')
    ax2.grid(True, alpha=0.3)
    
    # Add statistics text
    stats_text = f'Mean: {duration_ms.mean():.2f} ms\nStd: {duration_ms.std():.2f} ms'
    ax2.text(0.02, 0.98, stats_text, transform=ax2.transAxes, 
             verticalalignment='top', bbox=dict(boxstyle='round', facecolor='white', alpha=0.8))
    
    plt.tight_layout()
    plt.savefig(output_dir / 'performance_metrics.png')
    plt.close()

def plot_convergence_analysis(data: pd.DataFrame, output_dir: Path) -> None:
    """Create convergence analysis plot with fitness improvement rate."""
    fig, (ax1, ax2) = plt.subplots(2, 1, figsize=(12, 10))
    
    # Calculate fitness improvement rate
    fitness_improvement = data['avg_fitness'].diff().fillna(0)
    
    # Plot 1: Fitness Improvement Rate
    ax1.bar(data['generation'], fitness_improvement, alpha=0.7, color='teal')
    ax1.axhline(y=0, color='black', linestyle='-', alpha=0.3)
    ax1.set_ylabel('Fitness Improvement')
    ax1.set_title('Fitness Improvement Rate per Generation')
    ax1.grid(True, alpha=0.3)
    
    # Plot 2: Convergence Detection (rolling standard deviation)
    window_size = min(10, len(data))
    if window_size > 1:
        rolling_std = data['avg_fitness'].rolling(window=window_size).std()
        ax2.plot(data['generation'], rolling_std, linewidth=2, color='orange')
        ax2.set_ylabel('Rolling Std Dev')
        ax2.set_title(f'Convergence Stability ({window_size}-gen rolling std dev)')
        ax2.grid(True, alpha=0.3)
        
        # Add convergence threshold line
        threshold = rolling_std.quantile(0.25)  # 25th percentile as threshold
        ax2.axhline(y=threshold, color='red', linestyle='--', alpha=0.7, 
                   label=f'Convergence Threshold: {threshold:.3f}')
        ax2.legend()
    
    plt.tight_layout()
    plt.savefig(output_dir / 'convergence_analysis.png')
    plt.close()

def create_dashboard(data: pd.DataFrame, output_dir: Path) -> None:
    """Create comprehensive dashboard with subplots for all key metrics."""
    fig, ((ax1, ax2), (ax3, ax4)) = plt.subplots(2, 2, figsize=(16, 12))
    fig.suptitle('Evolution Algorithm Dashboard', fontsize=16, fontweight='bold')
    
    # Convert duration to milliseconds
    duration_ms = data['duration_ns'] / 1_000_000
    
    # Subplot 1: Fitness Evolution
    ax1.plot(data['generation'], data['best_fitness'], label='Best', linewidth=2)
    ax1.plot(data['generation'], data['avg_fitness'], label='Average', linewidth=2)
    ax1.plot(data['generation'], data['min_fitness'], label='Min', linewidth=1, alpha=0.7)
    ax1.set_title('Fitness Evolution')
    ax1.set_xlabel('Generation')
    ax1.set_ylabel('Fitness')
    ax1.legend()
    ax1.grid(True, alpha=0.3)
    
    # Subplot 2: Tree Complexity
    ax2.plot(data['generation'], data['min_depth'], label='Min', linewidth=2)
    ax2.plot(data['generation'], data['avg_depth'], label='Average', linewidth=2)
    ax2.plot(data['generation'], data['max_depth'], label='Max', linewidth=2)
    ax2.fill_between(data['generation'], data['min_depth'], data['max_depth'], alpha=0.2)
    ax2.set_title('Tree Complexity')
    ax2.set_xlabel('Generation')
    ax2.set_ylabel('Depth')
    ax2.legend()
    ax2.grid(True, alpha=0.3)
    
    # Subplot 3: Performance
    ax3.plot(data['generation'], duration_ms, color='purple', linewidth=2)
    ax3.set_title('Generation Duration')
    ax3.set_xlabel('Generation')
    ax3.set_ylabel('Duration (ms)')
    ax3.grid(True, alpha=0.3)
    
    # Subplot 4: Fitness Distribution
    ax4.hist(data['avg_fitness'], bins=15, alpha=0.7, color='skyblue', edgecolor='black')
    ax4.set_title('Average Fitness Distribution')
    ax4.set_xlabel('Average Fitness')
    ax4.set_ylabel('Frequency')
    ax4.grid(True, alpha=0.3)
    
    plt.tight_layout()
    plt.savefig(output_dir / 'dashboard.png')
    plt.close()

def validate_data(data: pd.DataFrame) -> None:
    """Validate that required columns exist in the dataframe."""
    required_columns = [
        'generation', 'duration_ns', 'best_fitness', 'avg_fitness', 
        'min_fitness', 'max_fitness', 'min_depth', 'max_depth', 'avg_depth'
    ]
    
    missing_columns = [col for col in required_columns if col not in data.columns]
    if missing_columns:
        raise ValueError(f"Missing required columns in CSV: {missing_columns}")

    # drop bad values
    data.replace([np.inf, -np.inf], np.nan, inplace=True)
    data.dropna(subset=required_columns, inplace=True)

def main(args: argparse.Namespace) -> None:
    """Main function to generate all plots."""
    if not os.path.exists(args.csv):
        raise FileNotFoundError(f"CSV file not found: {args.csv}")
    
    # Create output directory
    output_dir = Path(args.output_dir)
    output_dir.mkdir(exist_ok=True)
    
    # Load and validate data
    data = pd.read_csv(args.csv)
    validate_data(data)
    
    # Setup plotting style
    setup_plotting_style()
    
    print(f"Loaded {len(data)} generations of data from {args.csv}")
    print(f"Generating plots in {output_dir}")
    
    # Generate all plots
    print("Creating fitness evolution plot...")
    plot_fitness_evolution(data, output_dir)
    
    print("Creating complexity evolution plot...")
    plot_complexity_evolution(data, output_dir)
    
    print("Creating performance metrics plot...")
    plot_performance_metrics(data, output_dir)
    
    print("Creating convergence analysis plot...")
    plot_convergence_analysis(data, output_dir)
    
    print("Creating comprehensive dashboard...")
    create_dashboard(data, output_dir)
    
    print(f"Plots saved successfully to {output_dir}")

if __name__ == "__main__":
    main(parse_args())
