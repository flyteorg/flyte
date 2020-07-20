import cv2
import sys

def filter_edges(input_image_path: str, output_image_path: str):
    print("Reading {}".format(input_image_path))
    img = cv2.imread(input_image_path, 0)
    if img is None:
        raise Exception("Failed to read image")
    edges = cv2.Canny(img, 50, 200)  # hysteresis thresholds
    cv2.imwrite(output_image_path, edges)
    return output_image_path

if __name__ == "__main__":
    inp = sys.argv[1]
    out = sys.argv[2]
    out = "{}.png".format(out)
    print("filtering only edges from {}\n".format(inp))
    filter_edges(inp, out)
    print("Done, created {}".format(out))
