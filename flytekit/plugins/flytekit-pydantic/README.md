# Flytekit Pydantic Plugin

Pydantic is a data validation and settings management library that uses Python type annotations to enforce type hints at runtime and provide user-friendly errors when data is invalid. Pydantic models are classes that inherit from `pydantic.BaseModel` and are used to define the structure and validation of data using Python type annotations.

The plugin adds type support for pydantic models.

To install the plugin, run the following command:

```bash
pip install flytekitplugins-pydantic
```


## Type Example
```python
from pydantic import BaseModel


class TrainConfig(BaseModel):
    lr: float = 1e-3
    batch_size: int = 32
    files: List[FlyteFile]
    directories: List[FlyteDirectory]

@task
def train(cfg: TrainConfig):
    ...
```
