from dataclasses import asdict, dataclass
from typing import NamedTuple

import pytest
import torch.nn as nn
import torch.nn.functional as F
import torch.optim as optim
from dataclasses_json import DataClassJsonMixin

from flytekit import task, workflow
from flytekit.core.type_engine import TypeTransformerFailedError
from flytekit.extras.pytorch import PyTorchCheckpoint


@dataclass
class Hyperparameters(DataClassJsonMixin):
    epochs: int
    loss: float


class TupleHyperparameters(NamedTuple):
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
def generate_model_dict(hyperparameters: Hyperparameters) -> PyTorchCheckpoint:
    bn = Net()
    optimizer = optim.SGD(bn.parameters(), lr=0.001, momentum=0.9)
    return PyTorchCheckpoint(module=bn, hyperparameters=asdict(hyperparameters), optimizer=optimizer)


@task
def generate_model_tuple() -> PyTorchCheckpoint:
    bn = Net()
    optimizer = optim.SGD(bn.parameters(), lr=0.001, momentum=0.9)
    return PyTorchCheckpoint(module=bn, hyperparameters=TupleHyperparameters(epochs=5, loss=0.4), optimizer=optimizer)


@task
def generate_model_dataclass(hyperparameters: Hyperparameters) -> PyTorchCheckpoint:
    bn = Net()
    optimizer = optim.SGD(bn.parameters(), lr=0.001, momentum=0.9)
    return PyTorchCheckpoint(module=bn, hyperparameters=hyperparameters, optimizer=optimizer)


@task
def generate_model_only_module() -> PyTorchCheckpoint:
    bn = Net()
    return PyTorchCheckpoint(module=bn)


@task
def empty_checkpoint():
    with pytest.raises(TypeTransformerFailedError):
        return PyTorchCheckpoint()


@task
def t1(checkpoint: PyTorchCheckpoint):
    new_bn = Net()
    new_bn.load_state_dict(checkpoint["module_state_dict"])
    optimizer = optim.SGD(new_bn.parameters(), lr=0.001, momentum=0.9)
    optimizer.load_state_dict(checkpoint["optimizer_state_dict"])

    assert checkpoint["epochs"] == 5
    assert checkpoint["loss"] == 0.4


@workflow
def wf():
    checkpoint_dict = generate_model_dict(hyperparameters=Hyperparameters(epochs=5, loss=0.4))
    checkpoint_tuple = generate_model_tuple()
    checkpoint_dataclass = generate_model_dataclass(hyperparameters=Hyperparameters(epochs=5, loss=0.4))
    t1(checkpoint=checkpoint_dict)
    t1(checkpoint=checkpoint_tuple)
    t1(checkpoint=checkpoint_dataclass)
    generate_model_only_module()
    empty_checkpoint()


@workflow
def test_wf():
    wf()
