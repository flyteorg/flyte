#https://github.com/kubeflow/pytorch-operator/blob/b7fef224fef1ef0117f6e74961b557270fcf4b04/examples/mnist/mnist.py
from __future__ import absolute_import
from __future__ import print_function

import argparse
import os

from tensorboardX import SummaryWriter
from torchvision import datasets, transforms
import torch
import torch.distributed as dist
import torch.nn as nn
import torch.nn.functional as F
import torch.optim as optim

from flytekit.sdk.tasks import inputs, outputs
from flytekit.sdk.types import Types
from flytekit.sdk.workflow import workflow_class, Input, Output

from flytekit.sdk.tasks import pytorch_task

WORLD_SIZE = int(os.environ.get('WORLD_SIZE', 1))


class Net(nn.Module):
    def __init__(self):
        super(Net, self).__init__()
        self.conv1 = nn.Conv2d(1, 20, 5, 1)
        self.conv2 = nn.Conv2d(20, 50, 5, 1)
        self.fc1 = nn.Linear(4*4*50, 500)
        self.fc2 = nn.Linear(500, 10)

    def forward(self, x):
        x = F.relu(self.conv1(x))
        x = F.max_pool2d(x, 2, 2)
        x = F.relu(self.conv2(x))
        x = F.max_pool2d(x, 2, 2)
        x = x.view(-1, 4*4*50)
        x = F.relu(self.fc1(x))
        x = self.fc2(x)
        return F.log_softmax(x, dim=1)

def train(model, device, train_loader, optimizer, epoch, writer, log_interval):
    model.train()
    for batch_idx, (data, target) in enumerate(train_loader):
        data, target = data.to(device), target.to(device)
        optimizer.zero_grad()
        output = model(data)
        loss = F.nll_loss(output, target)
        loss.backward()
        optimizer.step()
        if batch_idx % log_interval == 0:
            print('Train Epoch: {} [{}/{} ({:.0f}%)]\tloss={:.4f}'.format(
                epoch, batch_idx * len(data), len(train_loader.dataset),
                       100. * batch_idx / len(train_loader), loss.item()))
            niter = epoch * len(train_loader) + batch_idx
            writer.add_scalar('loss', loss.item(), niter)

def test(model, device, test_loader, writer, epoch):
    model.eval()
    test_loss = 0
    correct = 0
    with torch.no_grad():
        for data, target in test_loader:
            data, target = data.to(device), target.to(device)
            output = model(data)
            test_loss += F.nll_loss(output, target, reduction='sum').item() # sum up batch loss
            pred = output.max(1, keepdim=True)[1] # get the index of the max log-probability
            correct += pred.eq(target.view_as(pred)).sum().item()

    test_loss /= len(test_loader.dataset)
    print('\naccuracy={:.4f}\n'.format(float(correct) / len(test_loader.dataset)))
    accuracy = float(correct) / len(test_loader.dataset)
    writer.add_scalar('accuracy', accuracy, epoch)
    return accuracy

def epoch_step(model, device, train_loader, test_loader, optimizer, epoch, writer, log_interval):
    train(model, device, train_loader, optimizer, epoch, writer, log_interval)
    return test(model, device, test_loader, writer, epoch)

def should_distribute():
    return dist.is_available() and WORLD_SIZE > 1


def is_distributed():
    return dist.is_available() and dist.is_initialized()

@inputs(
    no_cuda=Types.Boolean,
    batch_size=Types.Integer,
    test_batch_size=Types.Integer,
    epochs=Types.Integer,
    learning_rate=Types.Float,
    sgd_momentum=Types.Float,
    seed=Types.Integer,
    log_interval=Types.Integer,
    dir=Types.String)
@outputs(epoch_accuracies=[Types.Float], model_state=Types.Blob)
@pytorch_task(
    workers_count=2,
    per_replica_cpu_request="500m",
    per_replica_memory_request="4Gi",
    per_replica_memory_limit="8Gi",
    per_replica_gpu_limit="1",
)
def mnist_pytorch_job(workflow_params, no_cuda, batch_size, test_batch_size, epochs, learning_rate, sgd_momentum, seed, log_interval, dir, epoch_accuracies, model_state):
    backend_type = dist.Backend.GLOO

    writer = SummaryWriter(dir)

    torch.manual_seed(seed)

    use_cuda = not no_cuda
    device = torch.device('cuda' if use_cuda and torch.cuda.is_available else 'cpu')

    print('Using device: {}, world size: {}'.format(device, WORLD_SIZE))

    if should_distribute():
        print('Using distributed PyTorch with {} backend'.format(backend_type))
        dist.init_process_group(backend=backend_type)

    kwargs = {'num_workers': 1, 'pin_memory': True} if use_cuda else {}
    train_loader = torch.utils.data.DataLoader(
        datasets.MNIST('../data', train=True, download=True,
                       transform=transforms.Compose([
                           transforms.ToTensor(),
                           transforms.Normalize((0.1307,), (0.3081,))
                       ])),
        batch_size=batch_size, shuffle=True, **kwargs)
    test_loader = torch.utils.data.DataLoader(
        datasets.MNIST('../data', train=False, transform=transforms.Compose([
            transforms.ToTensor(),
            transforms.Normalize((0.1307,), (0.3081,))
        ])),
        batch_size=test_batch_size, shuffle=False, **kwargs)

    model = Net().to(device)

    if is_distributed():
        Distributor = nn.parallel.DistributedDataParallel if use_cuda and torch.cuda.is_available \
            else nn.parallel.DistributedDataParallelCPU
        model = Distributor(model)

    optimizer = optim.SGD(model.parameters(), lr=learning_rate, momentum=sgd_momentum)

    accuracies = [epoch_step(model, device, train_loader, test_loader, optimizer, epoch, writer, log_interval) for epoch in range(1, epochs + 1)]

    model_file = "mnist_cnn.pt"
    torch.save(model.state_dict(), model_file)

    model_state.set(model_file)
    epoch_accuracies.set(accuracies)


@workflow_class
class MNISTTest(object):
    no_cuda = Input(Types.Boolean, default=False, help="disables CUDA training")
    batch_size = Input(Types.Integer, default=64, help='input batch size for training (default: 64)')
    test_batch_size = Input(Types.Integer, default=1000, help='input batch size for testing (default: 1000)')
    epochs = Input(Types.Integer, default=1, help='number of epochs to train (default: 10)')
    learning_rate = Input(Types.Float, default=0.01, help='learning rate (default: 0.01)')
    sgd_momentum = Input(Types.Float, default=0.5, help='SGD momentum (default: 0.5)')
    seed = Input(Types.Integer, default=1, help='random seed (default: 1)')
    log_interval = Input(Types.Integer, default=10, help='how many batches to wait before logging training status')
    dir = Input(Types.String, default='logs', help='directory where summary logs are stored')

    mnist_result = mnist_pytorch_job(
        no_cuda=no_cuda,
        batch_size=batch_size,
        test_batch_size=test_batch_size,
        epochs=epochs,
        learning_rate=learning_rate,
        sgd_momentum=sgd_momentum,
        seed=seed,
        log_interval=log_interval,
        dir=dir
    )

    accuracies = Output(mnist_result.outputs.epoch_accuracies, sdk_type=[Types.Float])
    model = Output(mnist_result.outputs.model_state, sdk_type=Types.Blob)
