import pytest

from flytekit.image_spec.image_spec import ImageSpecBuilder


class MockImageSpecBuilder(ImageSpecBuilder):
    def build_image(self, img):
        print("Building an image...")


@pytest.fixture()
def mock_image_spec_builder():
    return MockImageSpecBuilder()
