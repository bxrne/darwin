import pandas as pd
import matplotlib.pyplot as plt
# Load the CSV
df = pd.read_csv("../test_small_argmax.csv")

# Split into two dataframes:
# Trees → depth > 0
trees = df[df["avg_depth"] != 0]

# Weights → depth == 0
weights = df[df["avg_depth"] == 0]


trees = trees.reset_index(drop=True)
weights = weights.reset_index(drop=True)

# trees = your dataframe created earlier
# trees = df[df["avg_depth"] != 0]
# Simple moving average smoother


def smooth(series, window=5):
    return series.rolling(window=window, center=True, min_periods=1).mean()

# ----------------------------------------------------------------------
# Plot avg/min/max for a metric (with shading under average)
# ----------------------------------------------------------------------


def plot_metric_group(df, base_name, title, smooth_window=5):
    plt.figure(figsize=(10, 5))
    x = df.index

    # Smooth curves
    min_y = smooth(df[f"min_{base_name}"], window=smooth_window)
    avg_y = smooth(df[f"avg_{base_name}"], window=smooth_window)
    max_y = smooth(df[f"max_{base_name}"], window=smooth_window)

    # Plot lines
    line_min = plt.plot(x, min_y, label=f"min_{base_name}")[0]
    line_avg = plt.plot(x, avg_y, label=f"avg_{base_name}")[0]
    line_max = plt.plot(x, max_y, label=f"max_{base_name}")[0]

    min_color = line_min.get_color()
    avg_color = line_avg.get_color()
    max_color = line_max.get_color()

    # --- Layered shading “following the curves” ---
    plt.fill_between(x, 0, min_y, color=min_color,
                     alpha=0.22)      # min: from 0 to min
    plt.fill_between(x, min_y, avg_y, color=avg_color,
                     alpha=0.22)  # avg: min to avg
    plt.fill_between(x, avg_y, max_y, color=max_color,
                     alpha=0.22)  # max: avg to max

    plt.xlabel("Index")
    plt.ylabel(base_name)
    plt.title(title)
    plt.legend()
    plt.tight_layout()
    plt.show()


# ----------------------------------------------------------------------
# 1. DEPTH
# ----------------------------------------------------------------------
plot_metric_group(trees, "depth", "Tree Depth (smoothed)")

# ----------------------------------------------------------------------
# 2. NODES
# ----------------------------------------------------------------------
plot_metric_group(trees, "nodes", "Node Count (smoothed)")

# ----------------------------------------------------------------------
# 3. FITNESS
# ----------------------------------------------------------------------
plot_metric_group(trees, "fit", "Fitness (smoothed)")

# ----------------------------------------------------------------------
# 4. ALL W0–W7 VALUES ON ONE PLOT
# ----------------------------------------------------------------------


def plot_all_w_values(df, smooth_window=5):
    plt.figure(figsize=(12, 6))
    x = df.index

    for w in range(0, 8):
        base = f"w{w}"
        for prefix in ["avg_"]:
            col = prefix + base
            if col in df.columns:
                y_smoothed = smooth(df[col], window=smooth_window)
                plt.plot(x, y_smoothed, label=col)

                # Shade under avg only
                if prefix == "avg_":
                    plt.fill_between(x, y_smoothed, alpha=0.12)

    plt.xlabel("Index")
    plt.ylabel("Weight values")
    plt.title("All w-values (avg/min/max for w0–w7) — smoothed")
    plt.legend(ncol=4, fontsize=8)
    plt.tight_layout()
    plt.show()


plot_all_w_values(trees)

# ----------------------------------------------------------------------
# 5. REMAINING METRICS
# ----------------------------------------------------------------------


def plot_metric_group_list(df, base_names, title, smooth_window=5):
    plt.figure(figsize=(14, 6))
    x = df.index

    for base_name in base_names:
        # Skip if the metric doesn't exist
        if f"min_{base_name}" not in df.columns:
            continue

        avg_y = smooth(df[f"avg_{base_name}"], window=smooth_window)

        # Plot lines
        line_avg = plt.plot(x, avg_y, label=f"avg_{base_name}")[0]

        # Get colors
        avg_color = line_avg.get_color()

        # Layered shading: min → avg → max
        plt.fill_between(x, 0, avg_y, color=avg_color, alpha=0.18)

    plt.xlabel("Index")
    plt.ylabel("Value")
    plt.title(title)
    plt.legend(ncol=3, fontsize=8)
    plt.tight_layout()
    plt.show()


# ----------------------------------------------------------------------
# Variable metrics (everything in variable_set)
# ----------------------------------------------------------------------
variable_set = [
    "army_diff",
    "land_diff",
    "distance_to_enemy_general",
    "enemy_general_x",
    "max_owned_army_x",
    "max_owned_army_y",
    "enemy_general_y",
    "min_city_x",
    "min_city_y",
    "visible_cities_count",
    "visible_mountains_count",
    "army_ratio",
    "land_ratio",
    "border_pressure",
]

plot_metric_group_list(trees, variable_set, "All Variables (min/avg/max)")

# ----------------------------------------------------------------------
# Operand metrics (everything in operand_set)
# ----------------------------------------------------------------------
operand_set = ["+", "-", "*", "/", "^"]

plot_metric_group_list(trees, operand_set, "All Operands (min/avg/max)")
