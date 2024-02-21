# Some standard imports
from pathlib import Path

import numpy as np
import onnxruntime
import requests
import torch.nn.init as init
import torch.onnx
import torch.utils.model_zoo as model_zoo
import torchvision.transforms as transforms
from flytekitplugins.onnxpytorch import PyTorch2ONNX, PyTorch2ONNXConfig
from PIL import Image
from torch import nn
from typing_extensions import Annotated

import flytekit
from flytekit import task, workflow
from flytekit.types.file import JPEGImageFile, ONNXFile


class SuperResolutionNet(nn.Module):
    def __init__(self, upscale_factor, inplace=False):
        super(SuperResolutionNet, self).__init__()

        self.relu = nn.ReLU(inplace=inplace)
        self.conv1 = nn.Conv2d(1, 64, (5, 5), (1, 1), (2, 2))
        self.conv2 = nn.Conv2d(64, 64, (3, 3), (1, 1), (1, 1))
        self.conv3 = nn.Conv2d(64, 32, (3, 3), (1, 1), (1, 1))
        self.conv4 = nn.Conv2d(32, upscale_factor**2, (3, 3), (1, 1), (1, 1))
        self.pixel_shuffle = nn.PixelShuffle(upscale_factor)

        self._initialize_weights()

    def forward(self, x):
        x = self.relu(self.conv1(x))
        x = self.relu(self.conv2(x))
        x = self.relu(self.conv3(x))
        x = self.pixel_shuffle(self.conv4(x))
        return x

    def _initialize_weights(self):
        init.orthogonal_(self.conv1.weight, init.calculate_gain("relu"))
        init.orthogonal_(self.conv2.weight, init.calculate_gain("relu"))
        init.orthogonal_(self.conv3.weight, init.calculate_gain("relu"))
        init.orthogonal_(self.conv4.weight)


def test_onnx_pytorch():
    @task
    def train() -> (
        Annotated[
            PyTorch2ONNX,
            PyTorch2ONNXConfig(
                args=torch.randn(1, 1, 224, 224, requires_grad=True),
                export_params=True,  # store the trained parameter weights inside
                opset_version=10,  # the ONNX version to export the model to
                do_constant_folding=True,  # whether to execute constant folding for optimization
                input_names=["input"],  # the model's input names
                output_names=["output"],  # the model's output names
                dynamic_axes={"input": {0: "batch_size"}, "output": {0: "batch_size"}},  # variable length axes
            ),
        ]
    ):
        # Create the super-resolution model by using the above model definition.
        torch_model = SuperResolutionNet(upscale_factor=3)

        # Load pretrained model weights
        model_url = "https://s3.amazonaws.com/pytorch/test_data/export/superres_epoch100-44c6958e.pth"

        # Initialize model with the pretrained weights
        map_location = lambda storage, loc: storage  # noqa: E731
        if torch.cuda.is_available():
            map_location = None
        torch_model.load_state_dict(model_zoo.load_url(model_url, map_location=map_location))

        return PyTorch2ONNX(model=torch_model)

    @task
    def onnx_predict(model_file: ONNXFile) -> JPEGImageFile:
        ort_session = onnxruntime.InferenceSession(model_file.download())

        img = Image.open(
            requests.get(
                "https://raw.githubusercontent.com/flyteorg/static-resources/main/flytekit/onnx/cat.jpg", stream=True
            ).raw
        )

        resize = transforms.Resize([224, 224])
        img = resize(img)

        img_ycbcr = img.convert("YCbCr")
        img_y, img_cb, img_cr = img_ycbcr.split()

        to_tensor = transforms.ToTensor()
        img_y = to_tensor(img_y)
        img_y.unsqueeze_(0)

        # compute ONNX Runtime output prediction
        ort_inputs = {
            ort_session.get_inputs()[0].name: img_y.detach().cpu().numpy()
            if img_y.requires_grad
            else img_y.cpu().numpy()
        }
        ort_outs = ort_session.run(None, ort_inputs)
        img_out_y = ort_outs[0]

        img_out_y = Image.fromarray(np.uint8((img_out_y[0] * 255.0).clip(0, 255)[0]), mode="L")

        # get the output image follow post-processing step from PyTorch implementation
        final_img = Image.merge(
            "YCbCr",
            [
                img_out_y,
                img_cb.resize(img_out_y.size, Image.BICUBIC),
                img_cr.resize(img_out_y.size, Image.BICUBIC),
            ],
        ).convert("RGB")

        img_path = Path(flytekit.current_context().working_directory) / "cat_superres_with_ort.jpg"
        final_img.save(img_path)

        # Save the image, we will compare this with the output image from mobile device
        return JPEGImageFile(path=str(img_path))

    @workflow
    def wf() -> JPEGImageFile:
        model = train()
        return onnx_predict(model_file=model)

    print(wf())
