import torch

from flytekit import task, workflow


@task
def generate_tensor_1d() -> torch.Tensor:
    return torch.zeros(5, dtype=torch.int32)


@task
def generate_tensor_2d() -> torch.Tensor:
    return torch.tensor([[1.0, -1.0, 2], [1.0, -1.0, 9], [0, 7.0, 3]])


@task
def generate_module() -> torch.nn.Module:
    bn = torch.nn.BatchNorm1d(3, track_running_stats=True)
    return bn


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
def generate_model() -> torch.nn.Module:
    return MyModel()


@task
def t1(tensor: torch.Tensor) -> torch.Tensor:
    assert tensor.dtype == torch.int32
    tensor[0] = 1
    return tensor


@task
def t2(tensor: torch.Tensor) -> torch.Tensor:
    # convert 2D to 3D
    tensor.unsqueeze_(-1)
    return tensor.expand(3, 3, 2)


@task
def t3(model: torch.nn.Module) -> torch.Tensor:
    return model.weight


@task
def t4(model: torch.nn.Module) -> torch.nn.Module:
    return model.l1


@workflow
def wf():
    t1(tensor=generate_tensor_1d())
    t2(tensor=generate_tensor_2d())
    t3(model=generate_module())
    t4(model=MyModel())


@workflow
def test_wf():
    wf()
