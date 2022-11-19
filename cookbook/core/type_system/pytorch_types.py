"""
.. _pytorch_types:

PyTorch Types
=============

.. tags:: MachineLearning, Basic

Flyte promotes the use of strongly-typed data to make it easier to write pipelines that are more robust and easier to test.
Flyte is primarily used for machine learning besides data engineering. To simplify the communication between Flyte tasks, especially when passing
around tensors and models, we added support for the PyTorch types.
"""

# %%
# Tensors & Modules
# -----------------
#
# Many a times, you may need to pass around tensors and modules (aka models). In the absence of native type support for PyTorch tensors and modules,
# Flytekit resorts to using pickle to serialize and deserialize the entities; in fact, pickle is used for any unknown type.
# This is not very efficient, and hence, we added PyTorch's serialization and deserialization support to the Flyte type system.
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


# %%
# Passing around tensors and modules is no more a hassle!

# %%
# Checkpoint
# ----------
#
# ``PyTorchCheckpoint`` is a special type of checkpoint to serialize and deserialize PyTorch models.
# It checkpoints ``torch.nn.Module``'s state, hyperparameters, and optimizer state.
# The module checkpoint differs from the standard checkpoint in that it checkpoints the module's ``state_dict``.
# Hence, when restoring the module, the module's ``state_dict`` needs to be used in conjunction with the actual module.
# As per PyTorch `docs <https://pytorch.org/tutorials/beginner/saving_loading_models.html#save-load-entire-model>`__, it is recommended to
# store the module's ``state_dict`` rather than the module itself. However, the serialization should work either way.
import torch.nn as nn
import torch.nn.functional as F
import torch.optim as optim
from dataclasses import dataclass

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
    return PyTorchCheckpoint(
        module=bn, hyperparameters=hyperparameters, optimizer=optimizer
    )


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


# %%
# .. note::
#   ``PyTorchCheckpoint`` supports serializing hyperparameters of types ``dict``, ``NamedTuple``, and ``dataclass``.
#
# Auto GPU to CPU & CPU to GPU Conversion
# ---------------------------------------
#
# Not all PyTorch computations require a GPU to run. There are some cases where it is beneficial to move the computation to the CPU after, say,
# the model is trained on a GPU. To avail the GPU power, we do ``to(torch.device("cuda"))``.
# To use the GPU-variables on a CPU, we need to move the variables to a CPU using ``to(torch.device("cpu"))`` construct.
# This manual conversion proposed by PyTorch is not very user friendly, and hence,
# we added support for automatic GPU to CPU conversion (and vice versa) for the PyTorch types.
#
# .. code-block:: python
#
#   from flytekit import Resources
#   from typing import Tuple
#
#
#   @task(requests=Resources(gpu="1"))
#   def train() -> Tuple[PyTorchCheckpoint, torch.Tensor, torch.Tensor, torch.Tensor]:
#       ...
#       device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
#       model = Model(X_train.shape[1])
#       model.to(device)
#       ...
#       X_train, X_test = X_train.to(device), X_test.to(device)
#       y_train, y_test = y_train.to(device), y_test.to(device)
#       ...
#       return PyTorchCheckpoint(module=model), X_train, X_test, y_test
#
#   @task
#   def predict(
#     checkpoint: PyTorchCheckpoint,
#     X_train: torch.Tensor,
#     X_test: torch.Tensor,
#     y_test: torch.Tensor,
#   ):
#       new_bn = Model(X_train.shape[1])
#       new_bn.load_state_dict(checkpoint["module_state_dict"])
#
#       accuracy_list = np.zeros((5,))
#
#       with torch.no_grad():
#           y_pred = new_bn(X_test)
#           correct = (torch.argmax(y_pred, dim=1) == y_test).type(torch.FloatTensor)
#           accuracy_list = correct.mean()
#
# The ``predict`` task here runs on a CPU.
# As can be seen, you need not do the device conversion from GPU to CPU in the ``predict`` task as that's handled automatically by Flytekit!
