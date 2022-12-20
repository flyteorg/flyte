"""
Distributed Pytorch on Sagemaker
################################

This example is adapted from
`this sagemaker example <https://github.com/aws/amazon-sagemaker-examples/blob/89831fcf99ea3110f52794db0f6433a4013a5bca/sagemaker-python-sdk/pytorch_mnist/mnist.py>`__:

It shows how distributed training can be completely performed on the user side with minimal changes using Flyte.

.. NOTE::

    Flytekit will be adding further simplifications to make writing a distributed training algorithm even simpler, but
    this example basically provides the full details.
"""
import logging
import os
import typing
from dataclasses import dataclass

import flytekit
import torch
import torch.distributed as dist
import torch.multiprocessing as mp
import torch.nn as nn
import torch.nn.functional as functional
import torch.optim as optim
from dataclasses_json import dataclass_json
from flytekit import task, workflow
from flytekit.types.file import PythonPickledFile
from flytekitplugins.awssagemaker import (
    AlgorithmName,
    AlgorithmSpecification,
    InputContentType,
    InputMode,
    SagemakerTrainingJobConfig,
    TrainingJobResourceConfig,
)
from torchvision import datasets, transforms


@dataclass_json
@dataclass
class Hyperparameters(object):
    """
    Args:
        batch_size: input batch size for training (default: 64)
        test_batch_size: input batch size for testing (default: 1000)
        epochs: number of epochs to train (default: 10)
        learning_rate: learning rate (default: 0.01)
        sgd_momentum: SGD momentum (default: 0.5)
        seed: random seed (default: 1)
        log_interval: how many batches to wait before logging training status
        dir: directory where summary logs are stored
    """

    backend: str = dist.Backend.GLOO
    sgd_momentum: float = 0.5
    seed: int = 1
    log_interval: int = 10
    batch_size: int = 64
    test_batch_size: int = 1000
    epochs: int = 10
    learning_rate: float = 0.01


@dataclass
class TrainingArgs(Hyperparameters):
    """
    These are training arguments that contain additional metadata beyond the hyper parameters useful especially in
    distributed training:
    """

    hosts: typing.List[int] = None
    current_host: int = 0
    num_gpus: int = 0
    data_dir: str = "/tmp"
    model_dir: str = "/tmp"

    def is_distributed(self) -> bool:
        return len(self.hosts) > 1 and self.backend is not None


# Based on https://github.com/pytorch/examples/blob/master/mnist/main.py
class Net(nn.Module):
    def __init__(self):
        super(Net, self).__init__()
        self.conv1 = nn.Conv2d(1, 10, kernel_size=5)
        self.conv2 = nn.Conv2d(10, 20, kernel_size=5)
        self.conv2_drop = nn.Dropout2d()
        self.fc1 = nn.Linear(320, 50)
        self.fc2 = nn.Linear(50, 10)

    def forward(self, x):
        x = functional.relu(functional.max_pool2d(self.conv1(x), 2))
        x = functional.relu(functional.max_pool2d(self.conv2_drop(self.conv2(x)), 2))
        x = x.view(-1, 320)
        x = functional.relu(self.fc1(x))
        x = functional.dropout(x, training=self.training)
        x = self.fc2(x)
        return functional.log_softmax(x, dim=1)


def _get_train_data_loader(batch_size, training_dir, is_distributed, **kwargs):
    logging.info("Get train data loader")
    dataset = datasets.MNIST(
        training_dir,
        train=True,
        download=False,
        transform=transforms.Compose(
            [transforms.ToTensor(), transforms.Normalize((0.1307,), (0.3081,))]
        ),
    )
    logging.info("Dataset is downloaded. Creating a train_sampler")
    train_sampler = (
        torch.utils.data.distributed.DistributedSampler(dataset)
        if is_distributed
        else None
    )
    logging.info("Train_sampler is successfully created. Creating a DataLoader")
    return torch.utils.data.DataLoader(
        dataset,
        batch_size=batch_size,
        shuffle=train_sampler is None,
        sampler=train_sampler,
        **kwargs,
    )


def _get_test_data_loader(test_batch_size, training_dir, **kwargs):
    logging.info("Get test data loader")
    return torch.utils.data.DataLoader(
        datasets.MNIST(
            training_dir,
            train=False,
            download=False,
            transform=transforms.Compose(
                [transforms.ToTensor(), transforms.Normalize((0.1307,), (0.3081,))]
            ),
        ),
        batch_size=test_batch_size,
        shuffle=True,
        **kwargs,
    )


def _average_gradients(model):
    # Gradient averaging.
    size = float(dist.get_world_size())
    for param in model.parameters():
        dist.all_reduce(param.grad.data, op=dist.reduce_op.SUM)
        param.grad.data /= size


def configure_model(model, is_distributed, gpu):
    if is_distributed:
        # multi-machine multi-gpu case
        model = torch.nn.parallel.DistributedDataParallel(
            model, device_ids=[gpu], output_device=gpu
        )
    else:
        # single-machine multi-gpu case or single-machine or multi-machine cpu case
        model = torch.nn.DataParallel(model)
    return model


# %%
# The Actual Trainer
def train(gpu: int, args: TrainingArgs):
    logging.basicConfig(level="INFO")
    is_distributed = args.is_distributed()
    logging.warning("Distributed training - {}".format(is_distributed))
    use_cuda = args.num_gpus > 0
    logging.warning("Number of gpus available - {}".format(args.num_gpus))
    kwargs = {"num_workers": 1, "pin_memory": True} if use_cuda else {}
    device = torch.device("cuda" if use_cuda else "cpu")

    rank = 0

    if is_distributed:
        # Initialize the distributed environment
        world_size = len(args.hosts) * args.num_gpus
        os.environ["WORLD_SIZE"] = str(world_size)
        rank = args.hosts.index(args.current_host) * args.num_gpus + gpu
        os.environ["RANK"] = str(rank)
        dist.init_process_group(
            backend=args.backend, init_method="env://", rank=rank, world_size=world_size
        )
        logging.info(
            "Initialized the distributed environment: '{}' backend on {} nodes. ".format(
                args.backend, dist.get_world_size()
            )
            + "Current host rank is {}. Number of gpus: {}".format(
                dist.get_rank(), args.num_gpus
            )
        )
        torch.cuda.set_device(gpu)

    # set the seed for generating random numbers
    torch.manual_seed(args.seed)
    if use_cuda:
        torch.cuda.manual_seed(args.seed)

    train_loader = _get_train_data_loader(
        args.batch_size, args.data_dir, is_distributed, **kwargs
    )
    test_loader = _get_test_data_loader(args.test_batch_size, args.data_dir, **kwargs)

    logging.info(
        "Processes {}/{} ({:.0f}%) of train data".format(
            len(train_loader.sampler),
            len(train_loader.dataset),
            100.0 * len(train_loader.sampler) / len(train_loader.dataset),
        )
    )

    logging.info(
        "Processes {}/{} ({:.0f}%) of test data".format(
            len(test_loader.sampler),
            len(test_loader.dataset),
            100.0 * len(test_loader.sampler) / len(test_loader.dataset),
        )
    )

    model = Net().to(device)

    model = configure_model(model, is_distributed, gpu)

    optimizer = optim.SGD(
        model.parameters(), lr=args.learning_rate, momentum=args.sgd_momentum
    )

    logging.info(
        "[rank {}|local-rank {}] Totally {} epochs".format(rank, gpu, args.epochs + 1)
    )
    for epoch in range(1, args.epochs + 1):
        model.train()
        for batch_idx, (data, target) in enumerate(train_loader, 1):
            if use_cuda:
                data, target = (
                    data.cuda(non_blocking=True),
                    target.cuda(non_blocking=True),
                )
            optimizer.zero_grad()
            output = model(data)
            loss = functional.nll_loss(output, target)
            loss.backward()
            if is_distributed and not use_cuda:
                # average gradients manually for multi-machine cpu case only
                _average_gradients(model)
            optimizer.step()
            if batch_idx % args.log_interval == 0:
                if not is_distributed or (is_distributed and rank == 0):
                    logging.info(
                        "[rank {}|local-rank {}] Train Epoch: {} [{}/{} ({:.0f}%)] Loss: {:.6f}".format(
                            rank,
                            gpu,
                            epoch,
                            batch_idx * len(data),
                            len(train_loader.sampler),
                            100.0 * batch_idx / len(train_loader),
                            loss.item(),
                        )
                    )
        test(model, test_loader, device)

    if not is_distributed or (is_distributed and rank == 0):
        save_model(model, args.model_dir)


# %%
# Lets test the trained model
def test(model, test_loader, device):
    model.eval()
    test_loss = 0
    correct = 0
    with torch.no_grad():
        for data, target in test_loader:
            if device.type == "cuda":
                data, target = (
                    data.cuda(non_blocking=True),
                    target.cuda(non_blocking=True),
                )
            output = model(data)
            test_loss += functional.nll_loss(
                output, target, size_average=False
            ).item()  # sum up batch loss
            pred = output.max(1, keepdim=True)[
                1
            ]  # get the index of the max log-probability
            correct += pred.eq(target.view_as(pred)).sum().item()

    test_loss /= len(test_loader.dataset)
    logging.info(
        "Test set: Average loss: {:.4f}, Accuracy: {}/{} ({:.0f}%)\n".format(
            test_loss,
            correct,
            len(test_loader.dataset),
            100.0 * correct / len(test_loader.dataset),
        )
    )


def model_fn(model_dir):
    device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
    model = torch.nn.DataParallel(Net())
    with open(os.path.join(model_dir, "model.pth"), "rb") as f:
        model.load_state_dict(torch.load(f))
    return model.to(device)


# %%
# Save the model to a local path
def save_model(model, model_dir) -> PythonPickledFile:
    logging.info("Saving the model.")
    path = os.path.join(model_dir, "model.pth")
    # recommended way from http://pytorch.org/docs/master/notes/serialization.html
    torch.save(model.cpu().state_dict(), path)
    print(f"Model saved to {path}")
    return path


def download_training_data(training_dir):
    logging.info("Downloading train data")
    datasets.MNIST(
        training_dir,
        train=True,
        download=True,
        transform=transforms.Compose(
            [transforms.ToTensor(), transforms.Normalize((0.1307,), (0.3081,))]
        ),
    )


def download_test_data(training_dir):
    logging.info("Downloading test data")
    datasets.MNIST(
        training_dir,
        train=False,
        download=True,
        transform=transforms.Compose(
            [transforms.ToTensor(), transforms.Normalize((0.1307,), (0.3081,))]
        ),
    )


# https://github.com/aws/amazon-sagemaker-examples/blob/89831fcf99ea3110f52794db0f6433a4013a5bca/sagemaker-python-sdk/pytorch_mnist/mnist.py
@task(
    task_config=SagemakerTrainingJobConfig(
        algorithm_specification=AlgorithmSpecification(
            input_mode=InputMode.FILE,
            algorithm_name=AlgorithmName.CUSTOM,
            algorithm_version="",
            input_content_type=InputContentType.TEXT_CSV,
        ),
        training_job_resource_config=TrainingJobResourceConfig(
            instance_type="ml.p3.8xlarge",
            instance_count=2,
            volume_size_in_gb=25,
        ),
    ),
    cache_version="1.0",
    cache=True,
)
def mnist_pytorch_job(hp: Hyperparameters) -> PythonPickledFile:
    # pytorch's save() function does not create a path if the path specified does not exist
    # therefore we must pass an existing path

    ctx = flytekit.current_context()
    data_dir = os.path.join(ctx.working_directory, "data")
    model_dir = os.path.join(ctx.working_directory, "model")
    os.makedirs(data_dir, exist_ok=True)
    os.makedirs(model_dir, exist_ok=True)
    args = TrainingArgs(
        hosts=ctx.distributed_training_context.hosts,
        current_host=ctx.distributed_training_context.current_host,
        num_gpus=torch.cuda.device_count(),
        batch_size=hp.batch_size,
        test_batch_size=hp.test_batch_size,
        epochs=hp.epochs,
        learning_rate=hp.learning_rate,
        sgd_momentum=hp.sgd_momentum,
        seed=hp.seed,
        log_interval=hp.log_interval,
        backend=hp.backend,
        data_dir=data_dir,
        model_dir=model_dir,
    )

    # Data shouldn't be downloaded by the functions called in mp.spawn due to race conditions
    # These can be replaced by Flyte's blob type inputs. Note that the data here are assumed
    # to be accessible via a local path:
    download_training_data(args.data_dir)
    download_test_data(args.data_dir)

    if len(args.hosts) > 1:
        # Config MASTER_ADDR and MASTER_PORT for PyTorch Distributed Training
        os.environ["MASTER_ADDR"] = args.hosts[0]
        os.environ["MASTER_PORT"] = "29500"
        os.environ[
            "NCCL_SOCKET_IFNAME"
        ] = ctx.distributed_training_context.network_interface_name
        os.environ["NCCL_DEBUG"] = "INFO"
        # The function is called as fn(i, *args), where i is the process index and args is the passed
        # through tuple of arguments.
        # https://pytorch.org/docs/stable/multiprocessing.html#torch.multiprocessing.spawn
        mp.spawn(train, nprocs=args.num_gpus, args=(args,))
    else:
        # Config for Multi GPU with a single instance training
        if args.num_gpus > 1:
            gpu_devices = ",".join([str(gpu_id) for gpu_id in range(args.num_gpus)])
            os.environ["CUDA_VISIBLE_DEVICES"] = gpu_devices
        train(-1, args)

    pth = os.path.join(model_dir, "model.pth")
    print(f"Returning model @ {pth}")
    return pth


# %%
# Create a Pipeline
# ------------------
# Now the training and the plotting can be put together into a pipeline, where the training is performed first
# followed by the plotting of the accuracy. Data is passed between them and the workflow itself outputs the image and
# the serialize model:
@workflow
def pytorch_training_wf(hp: Hyperparameters) -> PythonPickledFile:
    return mnist_pytorch_job(hp=hp)


if __name__ == "__main__":
    model = pytorch_training_wf(hp=Hyperparameters(epochs=2, batch_size=128))
    print(f"Model: {model}")
