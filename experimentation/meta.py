import pandas as pd
import matplotlib.pyplot as plt
import sys
import os

def plot(file_name):
    # Load CSV data
    df = pd.read_csv(file_name)

    # Convert Start (nanoseconds since epoch) to datetime for easier plotting
    df['StartTime'] = pd.to_datetime(df['Start'], unit='ns')

    # Plot request durations over time
    plt.figure(figsize=(12, 6))
    plt.plot(df['StartTime'], df['Duration'], label='Duration (ns)', marker='o', linestyle='-')

    # Highlight errors with red markers
    errors = df[df['IsError'] == True]
    plt.scatter(errors['StartTime'], errors['Duration'], color='red', label='Errors', zorder=5)

    plt.xlabel("Start Time")
    plt.ylabel("Duration (nanoseconds)")
    plt.title("Request Duration Over Time (Metastability Visualization)")
    plt.legend()
    plt.grid(True)
    plt.tight_layout()

    base_name = os.path.splitext(os.path.basename(file_name))[0]
    output_file = f"{base_name}_plot.png"
    plt.savefig(output_file)
    print(f"Plot saved as {output_file}")


def main():
    if len(sys.argv) != 2:
        sys.exit(1)

    filename = sys.argv[1]
    plot(filename)

if __name__ == "__main__":
    main()
