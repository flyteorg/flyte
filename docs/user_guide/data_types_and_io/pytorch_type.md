---
jupytext:
  cell_metadata_filter: all
  formats: md:myst
  main_language: python
  notebook_metadata_filter: all
  text_representation:
    extension: .md
    format_name: myst
    format_version: 0.13
    jupytext_version: 1.16.1
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

+++ {"lines_to_next_cell": 0}

(pytorch_type)=

# PyTorch type

```{eval-rst}
.. tags:: MachineLearning, Basic
```

Flyte advocates for the use of strongly-typed data to simplify the development of robust and testable pipelines.
In addition to its application in data engineering, Flyte is primarily used for machine learning.
To streamline the communication between Flyte tasks, particularly when dealing with tensors and models,
we have introduced support for PyTorch types.

## Tensors and modules

At times, you may find the need to pass tensors and modules (models) within your workflow.
Without native support for PyTorch tensors and modules, Flytekit relies on {std:ref}`pickle <pickle_type>` for serializing
and deserializing these entities, as well as any unknown types.
However, this approach isn't the most efficient. As a result, we've integrated PyTorch's
serialization and deserialization support into the Flyte type system.

```{code-cell}
import torch
from flytekit import task, workflow


@task
def generate_tensor_2d() -> torch.Tensor:
    return torch.tensor([[1.0, -1.0, 2], [1.0, -1.0, 9], [0, 7.0, 3]])


@task
def reshape_tensor(tensor: torch.Tensor) -> torch.Tensor:
    # convert 2D to 3D
    tensor.unsqueeze_(-1)
    return tensor.expand(3, 3, 2)


@task
def generate_module() -> torch.nn.Module:
    bn = torch.nn.BatchNorm1d(3, track_running_stats=True)
    return bn


@task
def get_model_weight(model: torch.nn.Module) -> torch.Tensor:
    return model.weight


class MyModel(torch.nn.Module):
    def __init__(self):
        super(MyModel, self).__init__()
        self.l0 = torch.nn.Linear(4, 2)
        self.l1 = torch.nn.Linear(2, 1)

    def forward(self, input):
        out0 = self.l0(input)
        out0_relu = torch.nn.functional.relu(out0)
        return self.l1(out0_relu)


@task
def get_l1() -> torch.nn.Module:
    model = MyModel()
    return model.l1


@workflow
def pytorch_native_wf():
    reshape_tensor(tensor=generate_tensor_2d())
    get_model_weight(model=generate_module())
    get_l1()
```

+++ {"lines_to_next_cell": 0}

Passing around tensors and modules is no more a hassle!

## Checkpoint

`PyTorchCheckpoint` is a specialized checkpoint used for serializing and deserializing PyTorch models.
It checkpoints `torch.nn.Module`'s state, hyperparameters and optimizer state.

This module checkpoint differs from the standard checkpoint as it specifically captures the module's `state_dict`.
Therefore, when restoring the module, the module's `state_dict` must be used in conjunction with the actual module.
According to the PyTorch [docs](https://pytorch.org/tutorials/beginner/saving_loading_models.html#save-load-entire-model),
it's recommended to store the module's `state_dict` rather than the module itself,
although the serialization should work in either case.

```{code-cell}
:lines_to_next_cell: 2

from dataclasses import dataclass

import torch.nn as nn
import torch.nn.functional as F
import torch.optim as optim
from dataclasses_json import dataclass_json
from flytekit.extras.pytorch import PyTorchCheckpoint


@dataclass_json
@dataclass
class Hyperparameters:
    epochs: int
    loss: float


class Net(nn.Module):
    def __init__(self):
        super(Net, self).__init__()
        self.conv1 = nn.Conv2d(3, 6, 5)
        self.pool = nn.MaxPool2d(2, 2)
        self.conv2 = nn.Conv2d(6, 16, 5)
        self.fc1 = nn.Linear(16 * 5 * 5, 120)
        self.fc2 = nn.Linear(120, 84)
        self.fc3 = nn.Linear(84, 10)

    def forward(self, x):
        x = self.pool(F.relu(self.conv1(x)))
        x = self.pool(F.relu(self.conv2(x)))
        x = x.view(-1, 16 * 5 * 5)
        x = F.relu(self.fc1(x))
        x = F.relu(self.fc2(x))
        x = self.fc3(x)
        return x


@task
def generate_model(hyperparameters: Hyperparameters) -> PyTorchCheckpoint:
    bn = Net()
    optimizer = optim.SGD(bn.parameters(), lr=0.001, momentum=0.9)
    return PyTorchCheckpoint(module=bn, hyperparameters=hyperparameters, optimizer=optimizer)


@task
def load(checkpoint: PyTorchCheckpoint):
    new_bn = Net()
    new_bn.load_state_dict(checkpoint["module_state_dict"])
    optimizer = optim.SGD(new_bn.parameters(), lr=0.001, momentum=0.9)
    optimizer.load_state_dict(checkpoint["optimizer_state_dict"])


@workflow
def pytorch_checkpoint_wf():
    checkpoint = generate_model(hyperparameters=Hyperparameters(epochs=10, loss=0.1))
    load(checkpoint=checkpoint)
```

:::{note}
`PyTorchCheckpoint` supports serializing hyperparameters of types `dict`, `NamedTuple` and `dataclass`.
:::

## Auto GPU to CPU and CPU to GPU conversion

Not all PyTorch computations require a GPU. In some cases, it can be advantageous to transfer the
computation to a CPU, especially after training the model on a GPU.
To utilize the power of a GPU, the typical construct to use is: `to(torch.device("cuda"))`.

When working with GPU variables on a CPU, variables need to be transferred to the CPU using the `to(torch.device("cpu"))` construct.
However, this manual conversion recommended by PyTorch may not be very user-friendly.
To address this, we added support for automatic GPU to CPU conversion (and vice versa) for PyTorch types.

```python
from flytekit import Resources
from typing import Tuple


@task(requests=Resources(gpu="1"))
def train() -> Tuple[PyTorchCheckpoint, torch.Tensor, torch.Tensor, torch.Tensor]:
    ...
    device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
    model = Model(X_train.shape[1])
    model.to(device)
    ...
    X_train, X_test = X_train.to(device), X_test.to(device)
    y_train, y_test = y_train.to(device), y_test.to(device)
    ...
    return PyTorchCheckpoint(module=model), X_train, X_test, y_test

@task
def predict(
    checkpoint: PyTorchCheckpoint,
    X_train: torch.Tensor,
    X_test: torch.Tensor,
    y_test: torch.Tensor,
):
    new_bn = Model(X_train.shape[1])
    new_bn.load_state_dict(checkpoint["module_state_dict"])

    accuracy_list = np.zeros((5,))

    with torch.no_grad():
        y_pred = new_bn(X_test)
        correct = (torch.argmax(y_pred, dim=1) == y_test).type(torch.FloatTensor)
        accuracy_list = correct.mean()
```

The `predict` task will run on a CPU, and
the device conversion from GPU to CPU will be automatically handled by Flytekit.
