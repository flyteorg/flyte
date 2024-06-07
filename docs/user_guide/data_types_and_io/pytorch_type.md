
(pytorch_type)=

# PyTorch type

```{eval-rst}
.. tags:: MachineLearning, Basic
```

Flyte advocates for the use of strongly-typed data to simplify the development of robust and testable pipelines. In addition to its application in data engineering, Flyte is primarily used for machine learning.
To streamline the communication between Flyte tasks, particularly when dealing with tensors and models, we have introduced support for PyTorch types.

## Tensors and modules

At times, you may find the need to pass tensors and modules (models) within your workflow. Without native support for PyTorch tensors and modules, Flytekit relies on {std:ref}`pickle <pickle_type>` for serializing and deserializing these entities, as well as any unknown types. However, this approach isn't the most efficient. As a result, we've integrated PyTorch's serialization and deserialization support into the Flyte type system.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/pytorch_type.py
:caption: data_types_and_io/pytorch_type.py
:lines: 5-50
```

Passing around tensors and modules is no more a hassle!

## Checkpoint

`PyTorchCheckpoint` is a specialized checkpoint used for serializing and deserializing PyTorch models.
It checkpoints `torch.nn.Module`'s state, hyperparameters and optimizer state.

This module checkpoint differs from the standard checkpoint as it specifically captures the module's `state_dict`.
Therefore, when restoring the module, the module's `state_dict` must be used in conjunction with the actual module.
According to the PyTorch [docs](https://pytorch.org/tutorials/beginner/saving_loading_models.html#save-load-entire-model),
it's recommended to store the module's `state_dict` rather than the module itself,
although the serialization should work in either case.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/pytorch_type.py
:caption: data_types_and_io/pytorch_type.py
:lines: 63-117
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

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/data_types_and_io/
