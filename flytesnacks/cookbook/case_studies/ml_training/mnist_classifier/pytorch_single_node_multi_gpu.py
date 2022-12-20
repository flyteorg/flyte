"""
Single Node, Multi GPU Training
-------------------------------

When you need to scale up model training in pytorch, you can use the :py:class:`~torch:torch.nn.DataParallel` for
single node, multi-gpu/cpu training or :py:class:`~torch:torch.nn.parallel.DistributedDataParallel` for multi-node,
multi-gpu training.

This tutorial will cover how to write a simple training script on the MNIST dataset that uses
``DistributedDataParallel`` since its functionality is a superset of ``DataParallel``, supporting both single- and
multi-node training, and this is the `recommended way <https://pytorch.org/docs/stable/notes/cuda.html#cuda-nn-ddp-instead>`__
of distributing your training workload. Note, however, that this tutorial will only work for single-node, multi-gpu
settings.

For training on a single node and gpu see
:ref:`this tutorial <pytorch_single_node_and_gpu>`, and for more
information on distributed training, check out the
`pytorch documentation <https://pytorch.org/tutorials/intermediate/ddp_tutorial.html>`__.

The following video has further explanation:

..  youtube:: nTMFb7TmArI
"""

# %%
# Import the required libraries.
import json
import os
import typing

import torch
import torch.nn.functional as F
import wandb
from flytekit import Resources, task, workflow
from flytekit.types.file import PythonPickledFile

# %%
# We'll re-use certain classes and functions from the
# :ref:`single node and gpu tutorial <pytorch_single_node_and_gpu>`
# such as the ``Net`` model architecture, ``Hyperparameters``, and ``log_test_predictions``\.
from mnist_classifier.pytorch_single_node_and_gpu import Hyperparameters, Net, log_test_predictions
from torch import distributed as dist
from torch import multiprocessing as mp
from torch import nn, optim
from torchvision import datasets, transforms

# %%
# Let's define some variables to be used later.
#
# ``WORLD_SIZE`` defines the total number of GPUs we want to use to distribute our training job and ``DATA_DIR``
# specifies where the downloaded data should be written to.
WORLD_SIZE = 2
DATA_DIR = "./data"

# %%
# The following variables are specific to ``wandb``\:
#
# - ``NUM_BATCHES_TO_LOG``\: Number of batches to log from the test data for each test step
# - ``LOG_IMAGES_PER_BATCH``\: Number of images to log per test batch
NUM_BATCHES_TO_LOG = 10
LOG_IMAGES_PER_BATCH = 32


# %%
# If running remotely, copy your ``wandb`` API key to the Dockerfile under the environment variable ``WANDB_API_KEY``\.
# This function logs into ``wandb`` and initializes the project. If you built your Docker image with the
# ``WANDB_USERNAME``\, this will work. Otherwise, replace ``my-user-name`` with your ``wandb`` user name.
#
# We'll call this function in the ``pytorch_mnist_task`` defined below.
def wandb_setup():
    wandb.login()
    wandb.init(
        project="mnist-single-node-multi-gpu",
        entity=os.environ.get("WANDB_USERNAME", "my-user-name"),
    )


# %%
# Re-Using the Network From the Single GPU Example
# ================================================
#
# We'll use the same neural network architecture as the one we define in the
# :ref:`single node and gpu tutorial <sphx_glr_auto_case_studies_ml_training_mnist_classifier_pytorch_single_node_and_gpu.py>`.


# %%
# Data Downloader
# ===============
#
# We'll use this helper function to download the training and test sets before-hand to avoid race conditions when
# initializing the train and test dataloaders during training.
def download_mnist(data_dir):
    for train in [True, False]:
        datasets.MNIST(data_dir, train=train, download=True)


# %%
# The Data Loader
# ===============
#
# This function will be called in the training function to be distributed across all available GPUs. Note that
# we set ``download=False`` here to avoid race conditions as mentioned above.
def mnist_dataloader(
    data_dir,
    batch_size,
    train=True,
    distributed=False,
    rank=None,
    world_size=None,
    **kwargs,
):
    dataset = datasets.MNIST(
        data_dir,
        train=train,
        download=False,
        transform=transforms.Compose(
            [transforms.ToTensor(), transforms.Normalize((0.1307), (0.3081))]
        ),
    )
    if distributed:
        assert (
            rank is not None
        ), "rank needs to be specified when doing distributed training."
        sampler = torch.utils.data.distributed.DistributedSampler(
            dataset,
            rank=rank,
            num_replicas=1 if world_size is None else world_size,
            shuffle=True,
        )
    else:
        sampler = None
    return torch.utils.data.DataLoader(
        dataset,
        batch_size=batch_size,
        shuffle=False,
        sampler=sampler,
        **kwargs,
    )


# %%
# Training
# ========
#
# We define a ``train`` function to enclose the training loop per epoch, and  we log the loss and epoch progression,
# which can later be visualized in a ``wandb`` dashboard.
def train(model, rank, train_loader, optimizer, epoch, log_interval):
    model.train()

    # hooks into the model to collect gradients and the topology
    if rank == 0:
        wandb.watch(model)

    # loop through the training batches
    for batch_idx, (data, target) in enumerate(train_loader):
        data, target = data.to(rank), target.to(rank)  # device conversion
        optimizer.zero_grad()  # clear gradient
        output = model(data)  # forward pass
        loss = F.nll_loss(output, target)  # compute loss
        loss.backward()  # propagate the loss backward
        optimizer.step()  # update the model parameters

        if rank == 0 and batch_idx % log_interval == 0:
            # log epoch and loss
            print(
                "Train Epoch: {} [{}/{} ({:.0f}%)]\tloss={:.4f}".format(
                    epoch,
                    batch_idx * len(data),
                    len(train_loader.dataset),
                    100.0 * batch_idx / len(train_loader),
                    loss.item(),
                )
            )
            wandb.log({"loss": loss, "epoch": epoch})


# %%
# Evaluation
# ==========
#
# We define a ``test`` function to test the model on the test dataset, logging ``accuracy``\, and ``test_loss`` to a
# ``wandb`` `table <https://docs.wandb.ai/guides/data-vis/log-tables>`__, which helps us visualize the model's
# performance in a structured format.
def test(model, rank, test_loader):
    model.eval()

    # define ``wandb`` tabular columns and hooks into the model to collect gradients and the topology
    columns = ["id", "image", "guess", "truth", *[f"score_{i}" for i in range(10)]]
    if rank == 0:
        my_table = wandb.Table(columns=columns)
        wandb.watch(model)

    test_loss = 0
    correct = 0
    log_counter = 0

    # disable gradient
    with torch.no_grad():

        # loop through the test data loader
        total = 0.0
        for images, targets in test_loader:
            total += len(targets)
            images, targets = images.to(rank), targets.to(rank)  # device conversion
            outputs = model(images)  # forward pass -- generate predictions
            test_loss += F.nll_loss(
                outputs, targets, reduction="sum"
            ).item()  # sum up batch loss
            _, predicted = torch.max(
                outputs.data, 1
            )  # get the index of the max log-probability
            correct += (
                (predicted == targets).sum().item()
            )  # compare predictions to true label

            # log predictions to the ``wandb`` table
            if log_counter < NUM_BATCHES_TO_LOG:
                if rank == 0:
                    log_test_predictions(
                        images, targets, outputs, predicted, my_table, log_counter
                    )
                log_counter += 1

    # compute the average loss
    test_loss /= total
    accuracy = float(correct) / total

    if rank == 0:
        print("\ntest_loss={:.4f}\naccuracy={:.4f}\n".format(test_loss, accuracy))
        # log the average loss, accuracy, and table
        wandb.log(
            {
                "test_loss": test_loss,
                "accuracy": accuracy,
                "mnist_predictions": my_table,
            }
        )

    return accuracy


# %%
# Training and Evaluating
# =======================

TrainingOutputs = typing.NamedTuple(
    "TrainingOutputs",
    epoch_accuracies=typing.List[float],
    model_state=PythonPickledFile,
)


# %%
# Setting up Distributed Training
# ===============================
#
# ``dist_setup`` is a helper function that instantiates a distributed environment. We're pointing all of the
# processes across all available GPUs to the address of the main process.


def dist_setup(rank, world_size, backend):
    os.environ["MASTER_ADDR"] = "localhost"
    os.environ["MASTER_PORT"] = "8888"
    dist.init_process_group(backend, rank=rank, world_size=world_size)


# %%
# These global variables point to the location of where to save the model and validation accuracies.
MODEL_FILE = "./mnist_cnn.pt"
ACCURACIES_FILE = "./mnist_cnn_accuracies.json"


# %%
# Then we define the ``train_mnist`` function. Note the conditionals that check for ``rank == 0``\. These parts of the
# functions are only called in the main process, which is the ``0``\th rank. The reason for this is that we only want the
# main process to perform certain actions such as:
#
# - log metrics via ``wandb``
# - save the trained model to disk
# - keep track of validation metrics


def train_mnist(rank: int, world_size: int, hp: Hyperparameters):
    # store the hyperparameters' config in ``wandb``
    if rank == 0:
        wandb_setup()
        wandb.config.update(json.loads(hp.to_json()))

    # set random seed
    torch.manual_seed(hp.seed)

    use_cuda = torch.cuda.is_available()
    print(f"Using distributed PyTorch with {hp.backend} backend")
    print(f"Running MNIST training on rank {rank}, world size: {world_size}")
    print(f"Use cuda: {use_cuda}")
    dist_setup(rank, world_size, hp.backend)
    print(f"Rank {rank + 1}/{world_size} process initialized.\n")

    # load data
    kwargs = {"num_workers": 0, "pin_memory": True} if use_cuda else {}
    print("Getting data loaders")
    training_data_loader = mnist_dataloader(
        DATA_DIR,
        hp.batch_size,
        train=True,
        distributed=use_cuda,
        rank=rank,
        world_size=world_size,
        **kwargs,
    )
    test_data_loader = mnist_dataloader(
        DATA_DIR, hp.test_batch_size, train=False, **kwargs
    )

    # define the distributed model and optimizer
    print("Defining model")
    model = Net().cuda(rank)
    model = nn.parallel.DistributedDataParallel(model, device_ids=[rank])

    optimizer = optim.SGD(
        model.parameters(), lr=hp.learning_rate, momentum=hp.sgd_momentum
    )

    # train the model: run multiple epochs and capture the accuracies for each epoch
    print(f"Training for {hp.epochs} epochs")
    accuracies = []
    for epoch in range(1, hp.epochs + 1):
        train(model, rank, training_data_loader, optimizer, epoch, hp.log_interval)

        # only compute validation metrics in the main process
        if rank == 0:
            accuracies.append(test(model, rank, test_data_loader))

        # wait for the main process to complete validation before continuing the training process
        dist.barrier()

    if rank == 0:
        # tell wandb that we're done logging metrics
        wandb.finish()

        # after training the model, we can simply save it to disk and return it from the Flyte
        # task as a `flytekit.types.file.FlyteFile` type, which is the `PythonPickledFile`.
        # `PythonPickledFile` is simply a decorator on the `FlyteFile` that records the format
        # of the serialized model as `pickled`
        print("Saving model")
        torch.save(model.state_dict(), MODEL_FILE)

        # save epoch accuracies
        print("Saving accuracies")
        with open(ACCURACIES_FILE, "w") as fp:
            json.dump(accuracies, fp)

    print(f"Rank {rank + 1}/{world_size} process complete.\n")
    dist.destroy_process_group()  # clean up


# %%
# The output model using :py:func:`torch:torch.save` saves the `state_dict` as described
# `in pytorch docs <https://pytorch.org/tutorials/beginner/saving_loading_models.html#saving-and-loading-models>`_.
# A common convention is to have the ``.pt`` extension for the model file.
#
# .. note::
#    Note the usage of ``requests=Resources(gpu=WORLD_SIZE)``\. This will force Flyte to allocate this task onto a
#    machine with GPU(s), which in our case is 2 gpus. The task will be queued up until a machine with GPU(s) can be
#    procured. Also, for the GPU Training to work, the Dockerfile needs to be built as explained in the
#    :ref:`pytorch-dockerfile` section.

# %%
# Defining the ``task``
# =====================
#
# Next we define the flyte task that kicks off the distributed training process. Here we call the
# pytorch :py:func:`multiprocessing <torch:torch.multiprocessing.spawn>` function to initiate a process on each
# available GPU. Since we're parallelizing the data, each process will contain a copy of the model and pytorch
# will handle syncing the weights across all processes on ``optimizer.step()`` calls.
#
# Read more about pytorch distributed training `here <https://pytorch.org/tutorials/beginner/dist_overview.html>`__.


# %%
# Set memory, gpu and storage depending on whether we are trying to register against sandbox or not:
if os.getenv("SANDBOX") != "":
    mem = "100Mi"
    gpu = "0"
    storage = "500Mi"
    ephemeral_storage = "500Mi"
else:
    mem = "30Gi"
    gpu = str(WORLD_SIZE)
    ephemeral_storage = "500Mi"
    storage = "20Gi"


@task(
    retries=2,
    cache=True,
    cache_version="1.2",
    requests=Resources(
        gpu=gpu, mem=mem, storage=storage, ephemeral_storage=ephemeral_storage
    ),
    limits=Resources(
        gpu=gpu, mem=mem, storage=storage, ephemeral_storage=ephemeral_storage
    ),
)
def pytorch_mnist_task(hp: Hyperparameters) -> TrainingOutputs:
    print("Start MNIST training:")

    world_size = torch.cuda.device_count()
    print(f"Device count: {world_size}")
    download_mnist(DATA_DIR)
    mp.spawn(
        train_mnist,
        args=(world_size, hp),
        nprocs=world_size,
        join=True,
    )
    print("Training Complete")
    with open(ACCURACIES_FILE) as fp:
        accuracies = json.load(fp)
    return TrainingOutputs(
        epoch_accuracies=accuracies, model_state=PythonPickledFile(MODEL_FILE)
    )


# %%
# Finally, we define a workflow to run the training algorithm. We return the model and accuracies.
@workflow
def pytorch_training_wf(
    hp: Hyperparameters = Hyperparameters(epochs=10, batch_size=128)
) -> TrainingOutputs:
    return pytorch_mnist_task(hp=hp)


# %%
# Running the Model Locally
# =========================
#
# It is possible to run the model locally with almost no modifications (as long as the code takes care of resolving
# if the code is distributed or not). This is how to do it:
if __name__ == "__main__":
    model, accuracies = pytorch_training_wf(
        hp=Hyperparameters(epochs=10, batch_size=128)
    )
    print(f"Model: {model}, Accuracies: {accuracies}")

# %%
# Weights & Biases Report
# =======================
#
# You can refer to the complete ``wandb`` report `here <https://wandb.ai/niels-bantilan/mnist-single-node-multi-gpu/reports/Pytorch-Single-node-Multi-GPU-Report--Vmlldzo5Mjk4Nzk>`__.
#
# .. tip::
#   Many more customizations can be done to the report according to your requirements!
