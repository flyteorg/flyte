import math
import sys


def write_output(output_dir, output_file, v):
    with open(f"{output_dir}/{output_file}", "w") as f:
        f.write(str(v))


def calculate_area(a, b):
    return math.pi * a * b


def main(a, b, output_dir):
    a = float(a)
    b = float(b)

    area = calculate_area(a, b)

    write_output(output_dir, "area", area)
    write_output(output_dir, "metadata", "[from python rawcontainer]")


if __name__ == "__main__":
    a = sys.argv[1]
    b = sys.argv[2]
    output_dir = sys.argv[3]

    main(a, b, output_dir)
