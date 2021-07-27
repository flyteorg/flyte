"""
Single GPU Training
-------------------

Training a model on a single node on one GPU is as trivial as writing any Flyte task and simply setting the GPU to ``1``.
As long as the Docker image is built correctly with the right version of the GPU drivers and the Flyte backend is
provisioned to have GPU machines, Flyte will execute the task on a node that has GPU(s).

Currently, Flyte does not provide any specific task type for PyTorch (though it is entirely possible to provide a task-type
that supports *PyTorch-Ignite* or *PyTorch Lightening* support, but this is not critical). One can request for a GPU, simply
by setting GPU="1" resource request and then at runtime, the GPU will be provisioned.

In this example, we'll see how we can create any PyTorch model, train it using Flyte and a specialized container. 
"""

# %%
# First, let's import the libraries.
import json
import os
import typing
from dataclasses import dataclass

import torch
import torch.nn.functional as F
import wandb
from dataclasses_json import dataclass_json
from flytekit import Resources, task, workflow
from flytekit.types.file import PythonPickledFile
from torch import distributed as dist
from torch import nn, optim
from torchvision import datasets, transforms

# %%
# Let's define some variables to be used later.
WORLD_SIZE = int(os.environ.get("WORLD_SIZE", 1))

# %%
# The following variables are specific to ``wandb``:
#
# - ``NUM_BATCHES_TO_LOG``: Number of batches to log from the test data for each test step
# - ``LOG_IMAGES_PER_BATCH``: Number of images to log per test batch
NUM_BATCHES_TO_LOG = 10
LOG_IMAGES_PER_BATCH = 32

# %%
# If running remotely, copy your ``wandb`` API key to the Dockerfile. Next, login to ``wandb``.
# You can disable this if you're already logged in on your local machine.
wandb.login()

# %%
# Next, we initialize the ``wandb`` project.
#
# .. admonition:: MUST DO!
#
#   Replace ``entity`` value with your username.
wandb.init(project="mnist-single-node", entity="your-user-name")

# %%
# Creating the Network
# ====================
#
# We use a simple PyTorch model with :py:class:`pytorch:torch.nn.Conv2d` and :py:class:`pytorch:torch.nn.Linear` layers.
# Let's also use :py:func:`pytorch:torch.nn.functional.relu`, :py:func:`pytorch:torch.nn.functional.max_pool2d`, and
# :py:func:`pytorch:torch.nn.functional.relu` to define the forward pass.
class Net(nn.Module):
    def __init__(self):
        super(Net, self).__init__()
        self.conv1 = nn.Conv2d(1, 20, 5, 1)
        self.conv2 = nn.Conv2d(20, 50, 5, 1)
        self.fc1 = nn.Linear(4 * 4 * 50, 500)
        self.fc2 = nn.Linear(500, 10)

    def forward(self, x):
        x = F.relu(self.conv1(x))
        x = F.max_pool2d(x, 2, 2)
        x = F.relu(self.conv2(x))
        x = F.max_pool2d(x, 2, 2)
        x = x.view(-1, 4 * 4 * 50)
        x = F.relu(self.fc1(x))
        x = self.fc2(x)
        return F.log_softmax(x, dim=1)


# %%
# Training
# ========
#
# We define a ``train`` function to enclose the training loop per epoch, i.e., this gets called for every successive epoch.
# Additionally, we log the loss and epoch progression, which can later be visualized in a ``wandb`` dashboard.
def train(model, device, train_loader, optimizer, epoch, log_interval):
    model.train()

    # hooks into the model to collect gradients and the topology
    wandb.watch(model)

    # loop through the training batches
    for batch_idx, (data, target) in enumerate(train_loader):

        # device conversion
        data, target = data.to(device), target.to(device)

        # clear gradient
        optimizer.zero_grad()

        # forward pass
        output = model(data)

        # compute loss
        loss = F.nll_loss(output, target)

        # propagate the loss backward
        loss.backward()

        # update the model parameters
        optimizer.step()

        if batch_idx % log_interval == 0:
            print(
                "Train Epoch: {} [{}/{} ({:.0f}%)]\tloss={:.4f}".format(
                    epoch,
                    batch_idx * len(data),
                    len(train_loader.dataset),
                    100.0 * batch_idx / len(train_loader),
                    loss.item(),
                )
            )

            # log epoch and loss
            wandb.log({"loss": loss, "epoch": epoch})


# %%
# We define a test logger function which will be called when we run the model on test dataset.
def log_test_predictions(images, labels, outputs, predicted, my_table, log_counter):
    """
    Convenience funtion to log predictions for a batch of test images
    """

    # obtain confidence scores for all classes
    scores = F.softmax(outputs.data, dim=1)
    log_scores = scores.cpu().numpy()
    log_images = images.cpu().numpy()
    log_labels = labels.cpu().numpy()
    log_preds = predicted.cpu().numpy()

    # assign ids based on the order of the images
    _id = 0
    for i, l, p, s in zip(log_images, log_labels, log_preds, log_scores):

        # add required info to data table:
        # id, image pixels, model's guess, true label, scores for all classes
        img_id = str(_id) + "_" + str(log_counter)
        my_table.add_data(img_id, wandb.Image(i), p, l, *s)
        _id += 1
        if _id == LOG_IMAGES_PER_BATCH:
            break


# %%
# Evaluation
# ==========
#
# We define a ``test`` function to test the model on the test dataset.
#
# We log ``accuracy``, ``test_loss``, and a ``wandb`` `table <https://docs.wandb.ai/guides/data-vis/log-tables>`__.
# The ``wandb`` table can help in depicting the model's performance in a structured format.
def test(model, device, test_loader):

    # ``wandb`` tabular columns
    columns = ["id", "image", "guess", "truth"]
    for digit in range(10):
        columns.append("score_" + str(digit))
    my_table = wandb.Table(columns=columns)

    model.eval()

    # hooks into the model to collect gradients and the topology
    wandb.watch(model)

    test_loss = 0
    correct = 0
    log_counter = 0

    # disable gradient
    with torch.no_grad():

        # loop through the test data loader
        for images, targets in test_loader:

            # device conversion
            images, targets = images.to(device), targets.to(device)

            # forward pass -- generate predictions
            outputs = model(images)

            # sum up batch loss
            test_loss += F.nll_loss(outputs, targets, reduction="sum").item()

            # get the index of the max log-probability
            _, predicted = torch.max(outputs.data, 1)

            # compare predictions to true label
            correct += (predicted == targets).sum().item()

            # log predictions to the ``wandb`` table
            if log_counter < NUM_BATCHES_TO_LOG:
                log_test_predictions(
                    images, targets, outputs, predicted, my_table, log_counter
                )
                log_counter += 1

    # compute the average loss
    test_loss /= len(test_loader.dataset)

    print("\naccuracy={:.4f}\n".format(float(correct) / len(test_loader.dataset)))
    accuracy = float(correct) / len(test_loader.dataset)

    # log the average loss, accuracy, and table
    wandb.log(
        {"test_loss": test_loss, "accuracy": accuracy, "mnist_predictions": my_table}
    )

    return accuracy


# %%
# Next, we define a function that runs for every epoch. It calls the ``train`` and ``test`` functions.
def epoch_step(
    model, device, train_loader, test_loader, optimizer, epoch, log_interval
):
    train(model, device, train_loader, optimizer, epoch, log_interval)
    return test(model, device, test_loader)


# %%
# Hyperparameters
# ===============
#
# We define a few hyperparameters for training our model.
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


# %%
# Training and Evaluating
# =======================
#
# The output model using :py:func:`pytorch:torch.save` saves the `state_dict` as described
# `in pytorch docs <https://pytorch.org/tutorials/beginner/saving_loading_models.html#saving-and-loading-models>`_.
# A common convention is to have the ``.pt`` extension for the model file.
#
# .. note::
#    Note the usage of ``requests=Resources(gpu="1")``. This will force Flyte to allocate this task onto a machine with GPU(s).
#    The task will be queued up until a machine with GPU(s) can be procured. Also, for the GPU Training to work, the
#    Dockerfile needs to be built as explained in the :ref:`pytorch-dockerfile` section.
TrainingOutputs = typing.NamedTuple(
    "TrainingOutputs",
    epoch_accuracies=typing.List[float],
    model_state=PythonPickledFile,
)


@task(retries=2, cache=True, cache_version="1.0", requests=Resources(gpu="1"))
def train_mnist(hp: Hyperparameters) -> TrainingOutputs:

    # store the hyperparameters' config in ``wandb``
    cfg = wandb.config
    cfg.update(json.loads(hp.to_json()))
    print(wandb.config)

    # set random seed
    torch.manual_seed(hp.seed)

    # ideally, if GPU training is required, and if cuda is not available, we can raise an exception
    # however, as we want this algorithm to work locally as well (and most users don't have a GPU locally), we will fallback to using a CPU
    use_cuda = torch.cuda.is_available()
    print(f"Use cuda {use_cuda}")
    device = torch.device("cuda" if use_cuda else "cpu")

    print("Using device: {}, world size: {}".format(device, WORLD_SIZE))

    # load Data
    kwargs = {"num_workers": 1, "pin_memory": True} if use_cuda else {}
    training_data_loader = torch.utils.data.DataLoader(
        datasets.MNIST(
            "../data",
            train=True,
            download=True,
            transform=transforms.Compose(
                [transforms.ToTensor(), transforms.Normalize((0.1307,), (0.3081,))]
            ),
        ),
        batch_size=hp.batch_size,
        shuffle=True,
        **kwargs,
    )
    test_data_loader = torch.utils.data.DataLoader(
        datasets.MNIST(
            "../data",
            train=False,
            transform=transforms.Compose(
                [transforms.ToTensor(), transforms.Normalize((0.1307,), (0.3081,))]
            ),
        ),
        batch_size=hp.test_batch_size,
        shuffle=False,
        **kwargs,
    )

    # train the model
    model = Net().to(device)

    optimizer = optim.SGD(
        model.parameters(), lr=hp.learning_rate, momentum=hp.sgd_momentum
    )

    # run multiple epochs and capture the accuracies for each epoch
    accuracies = [
        epoch_step(
            model,
            device,
            train_loader=training_data_loader,
            test_loader=test_data_loader,
            optimizer=optimizer,
            epoch=epoch,
            log_interval=hp.log_interval,
        )
        for epoch in range(1, hp.epochs + 1)
    ]

    # after training the model, we can simply save it to disk and return it from the Flyte task as a :py:class:`flytekit.types.file.FlyteFile`
    # type, which is the ``PythonPickledFile``. ``PythonPickledFile`` is simply a decorator on the ``FlyteFile`` that records the format
    # of the serialized model as ``pickled``
    model_file = "mnist_cnn.pt"
    torch.save(model.state_dict(), model_file)

    return TrainingOutputs(
        epoch_accuracies=accuracies, model_state=PythonPickledFile(model_file)
    )


# %%
# Finally, we define a workflow to run the training algorithm. We return the model and accuracies.
@workflow
def pytorch_training_wf(
    hp: Hyperparameters,
) -> (PythonPickledFile, typing.List[float]):
    accuracies, model = train_mnist(hp=hp)
    return model, accuracies


# %%
# Running the Model Locally
# =========================
#
# It is possible to run the model locally with almost no modifications (as long as the code takes care of resolving
# if the code is distributed or not). This is how we can do it:
if __name__ == "__main__":
    model, accuracies = pytorch_training_wf(
        hp=Hyperparameters(epochs=10, batch_size=128)
    )
    print(f"Model: {model}, Accuracies: {accuracies}")

# %%
# Weights & Biases Report
# =======================
#
# Lastly, let's look at the reports that are generated by the model.
#
# .. figure:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flytesnacks/pytorch/single-node/wandb_graphs.png
#   :alt: Wandb Graphs
#   :class: with-shadow
#
#   Wandb Graphs
#
# .. figure:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flytesnacks/pytorch/single-node/wandb_table.png
#   :alt: Wandb Table
#   :class: with-shadow
#
#   Wandb Table
#
# You can refer to the complete ``wandb`` report `here <https://wandb.ai/samhita-alla/pytorch-single-node/reports/PyTorch-Single-Node-Training-Report--Vmlldzo4NzUwNjA>`__.
#
# .. tip::
#   A lot more customizations can be done to the report according to your requirement!
