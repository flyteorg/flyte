from native_typing import rotate


def test_rotate_mandelbrot():
    x = rotate.rotate(
        image_location="https://upload.wikimedia.org/wikipedia/commons/thumb/2/21/Mandel_zoom_00_mandelbrot_set.jpg/640px-Mandel_"
        "zoom_00_mandelbrot_set.jpg"
    )

    # Just test that the file is mostly there I guess. Not sure how to really test it.
    assert x.path.startswith("/tmp")
