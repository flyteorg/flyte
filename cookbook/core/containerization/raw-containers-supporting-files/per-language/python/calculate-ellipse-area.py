import math
import sys


def read_input(input_dir, v):
    with open(f"{input_dir}/{v}", "r") as f:
        return float(f.read())


def write_output(output_dir, output_file, v):
    with open(f"{output_dir}/{output_file}", "w") as f:
        f.write(str(v))


def calculate_area(a, b):
    return math.pi * a * b


def main(input_dir, output_dir):
    a = read_input(input_dir, "a")
    b = read_input(input_dir, "b")

    area = calculate_area(a, b)

    write_output(output_dir, "area", area)
    write_output(output_dir, "metadata", "[from python rawcontainer]")


if __name__ == "__main__":
    input_dir = sys.argv[1]
    output_dir = sys.argv[2]

    main(input_dir, output_dir)
