# Flytekit Python Plugins

All the Flytekit plugins maintained by the core team are added here. It is not necessary to add plugins here, but this is a good starting place.

## Currently Available Plugins 🔌

| Plugin                       | Installation                                              | Description                                                                                                                | Version                                                                                                                                                      | Type          |
|------------------------------|-----------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------|
| AWS Sagemaker Training       | ```bash pip install flytekitplugins-awssagemaker ```      | Installs SDK to author Sagemaker built-in and custom training jobs in python                                               | [![PyPI version fury.io](https://badge.fury.io/py/flytekitplugins-awssagemaker.svg)](https://pypi.python.org/pypi/flytekitplugins-awssagemaker/)             | Backend       |
| dask                         | ```bash pip install flytekitplugins-dask ```              | Installs SDK to author dask jobs that can be executed natively on Kubernetes using the Flyte backend plugin                | [![PyPI version fury.io](https://badge.fury.io/py/flytekitplugins-awssagemaker.svg)](https://pypi.python.org/pypi/flytekitplugins-dask/)             | Backend       |
| Hive Queries                 | ```bash pip install flytekitplugins-hive ```              | Installs SDK to author Hive Queries that can be executed on a configured hive backend using Flyte backend plugin           | [![PyPI version fury.io](https://badge.fury.io/py/flytekitplugins-hive.svg)](https://pypi.python.org/pypi/flytekitplugins-hive/)                             | Backend       |
| K8s distributed PyTorch Jobs | ```bash pip install flytekitplugins-kfpytorch ```         | Installs SDK to author Distributed pyTorch Jobs in python using Kubeflow PyTorch Operator                                  | [![PyPI version fury.io](https://badge.fury.io/py/flytekitplugins-kfpytorch.svg)](https://pypi.python.org/pypi/flytekitplugins-kfpytorch/)                   | Backend       |
| K8s native tensorflow Jobs   | ```bash pip install flytekitplugins-kftensorflow ```      | Installs SDK to author Distributed tensorflow Jobs in python using Kubeflow Tensorflow Operator                            | [![PyPI version fury.io](https://badge.fury.io/py/flytekitplugins-kftensorflow.svg)](https://pypi.python.org/pypi/flytekitplugins-kftensorflow/)             | Backend       |
| K8s native MPI Jobs          | ```bash pip install flytekitplugins-kfmpi ```             | Installs SDK to author Distributed MPI Jobs in python using Kubeflow MPI Operator                                          | [![PyPI version fury.io](https://badge.fury.io/py/flytekitplugins-kfmpi.svg)](https://pypi.python.org/pypi/flytekitplugins-kfmpi/)                           | Backend       |
| Papermill based Tasks        | ```bash pip install flytekitplugins-papermill ```         | Execute entire notebooks as Flyte Tasks and pass inputs and outputs between them and python tasks                          | [![PyPI version fury.io](https://badge.fury.io/py/flytekitplugins-papermill.svg)](https://pypi.python.org/pypi/flytekitplugins-papermill/)                   | Flytekit-only |
| Pod Tasks                    | ```bash pip install flytekitplugins-pod ```               | Installs SDK to author Pods in python. These pods can have multiple containers, use volumes and have non exiting side-cars | [![PyPI version fury.io](https://badge.fury.io/py/flytekitplugins-pod.svg)](https://pypi.python.org/pypi/flytekitplugins-pod/)                               | Flytekit-only |
| spark                        | ```bash pip install flytekitplugins-spark ```             | Installs SDK to author Spark jobs that can be executed natively on Kubernetes with a supported backend Flyte plugin        | [![PyPI version fury.io](https://badge.fury.io/py/flytekitplugins-spark.svg)](https://pypi.python.org/pypi/flytekitplugins-spark/)                           | Backend       |
| AWS Athena Queries           | ```bash pip install flytekitplugins-athena ```            | Installs SDK to author queries executed on AWS Athena                                                                      | [![PyPI version fury.io](https://badge.fury.io/py/flytekitplugins-athena.svg)](https://pypi.python.org/pypi/flytekitplugins-athena/)                         | Backend       |
| DOLT                         | ```bash pip install flytekitplugins-dolt ```              | Read & write dolt data sets and use dolt tables as native types                                                            | [![PyPI version fury.io](https://badge.fury.io/py/flytekitplugins-dolt.svg)](https://pypi.python.org/pypi/flytekitplugins-dolt/)                             | Flytekit-only |
| Pandera                      | ```bash pip install flytekitplugins-pandera ```           | Use Pandera schemas as native Flyte types, which enable data quality checks.                                               | [![PyPI version fury.io](https://badge.fury.io/py/flytekitplugins-pandera.svg)](https://pypi.python.org/pypi/flytekitplugins-pandera/)                       | Flytekit-only |
| SQLAlchemy                   | ```bash pip install flytekitplugins-sqlalchemy ```        | Write queries for any database that supports SQLAlchemy                                                                    | [![PyPI version fury.io](https://badge.fury.io/py/flytekitplugins-sqlalchemy.svg)](https://pypi.python.org/pypi/flytekitplugins-sqlalchemy/)                 | Flytekit-only |
| Great Expectations           | ```bash pip install flytekitplugins-great-expectations``` | Enforce data quality for various data types within Flyte                                                                   | [![PyPI version fury.io](https://badge.fury.io/py/flytekitplugins-great-expectations.svg)](https://pypi.python.org/pypi/flytekitplugins-great-expectations/) | Flytekit-only |
| Snowflake                    | ```bash pip install flytekitplugins-snowflake```          | Use Snowflake as a 'data warehouse-as-a-service' within Flyte                                                              | [![PyPI version fury.io](https://badge.fury.io/py/flytekitplugins-snowflake.svg)](https://pypi.python.org/pypi/flytekitplugins-snowflake/)                   | Backend       |
| dbt                          | ```bash pip install flytekitplugins-dbt```                | Run dbt within Flyte                                                                                                       | [![PyPI version fury.io](https://badge.fury.io/py/flytekitplugins-dbt.svg)](https://pypi.python.org/pypi/flytekitplugins-dbt/)                               | Flytekit-only |
| Huggingface                  | ```bash pip install flytekitplugins-huggingface```        | Read & write Hugginface Datasets as Flyte StructuredDatasets                                                               | [![PyPI version fury.io](https://badge.fury.io/py/flytekitplugins-huggingface.svg)](https://pypi.python.org/pypi/flytekitplugins-huggingface/)               | Flytekit-only |
| DuckDB                  | ```bash pip install flytekitplugins-duckdb```        | Run analytical workloads with ease using DuckDB | [![PyPI version fury.io](https://badge.fury.io/py/flytekitplugins-duckdb.svg)](https://pypi.python.org/pypi/flytekitplugins-duckdb/)               | Flytekit-only |

## Have a Plugin Idea? 💡
Please [file an issue](https://github.com/flyteorg/flyte/issues/new?assignees=&labels=untriaged%2Cplugins&template=backend-plugin-request.md&title=%5BPlugin%5D).

## Development 💻
Flytekit plugins are structured as micro-libs and can be authored in an independent repository.

> Refer to the [Python microlibs](https://medium.com/@jherreras/python-microlibs-5be9461ad979) blog to understand the idea of microlibs.

The plugins maintained by the core team can be found in this repository and provide a simple way of discovery.

## Unit tests 🧪
Plugins should have their own unit tests.

## Guidelines 📜
Some guidelines to help you write the Flytekit plugins better.

1. The folder name has to be `flytekit-*`, e.g., `flytekit-hive`. In case you want to group for a specific service, then use `flytekit-aws-athena`.
2. Flytekit plugins use a concept called [Namespace packages](https://packaging.python.org/guides/creating-and-discovering-plugins/#using-namespace-packages), and thus, the package structure is essential.

   Please use the following Python package structure:
   ```
   flytekit-myplugin/
      - README.md
      - setup.py
      - flytekitplugins/
          - myplugin/
             - __init__.py
      - tests
          - __init__.py
   ```
   *NOTE:* the inner package `flytekitplugins` DOES NOT have an `__init__.py` file.

3. The published packages have to be named `flytekitplugins-{package-name}`, where `{package-name}` is a unique identifier for the plugin.

4. The setup.py file has to have the following template. You can use it as is by editing the TODO sections.

```python
from setuptools import setup

# TODO put the plugin name here
PLUGIN_NAME = "<plugin-name e.g. pandera>"

# TODO decide if the plugin is regular or `data`
# for regular plugins
microlib_name = f"flytekitplugins-{PLUGIN_NAME}"
# For data/persistence plugins
# microlib_name = f"flytekitplugins-data-{PLUGIN_NAME}"

# TODO add additional requirements
plugin_requires = ["flytekit>=1.1.0b0,<2.0.0, "<other requirements>"]

__version__ = "0.0.0+develop"

setup(
    name=microlib_name,
    version=__version__,
    author="flyteorg",
    author_email="admin@flyte.org",
    # TODO Edit the description
    description="My awesome plugin.....",
    # TODO alter the last part of the following URL
    url="https://github.com/flyteorg/flytekit/tree/master/plugins/flytekit-...",
    long_description=open("README.md").read(),
    long_description_content_type="text/markdown",
    namespace_packages=["flytekitplugins"],
    packages=[f"flytekitplugins.{PLUGIN_NAME}"],
    install_requires=plugin_requires,
    license="apache2",
    python_requires=">=3.8",
    classifiers=[
        "Intended Audience :: Science/Research",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: Apache Software License",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Topic :: Scientific/Engineering",
        "Topic :: Scientific/Engineering :: Artificial Intelligence",
        "Topic :: Software Development",
        "Topic :: Software Development :: Libraries",
        "Topic :: Software Development :: Libraries :: Python Modules",
    ],
    # TODO OPTIONAL
    # FOR Plugins where auto-loading on installation is desirable, please uncomment this line and ensure that the
    # __init__.py has the right modules available to be loaded, or point to the right module
    # entry_points={"flytekit.plugins": [f"{PLUGIN_NAME}=flytekitplugins.{PLUGIN_NAME}"]},
)
```
5. Each plugin should have a README.md, which describes how to install it with a simple example. For example, refer to flytekit-greatexpectations' [README](./flytekit-greatexpectations/README.md).

6. Each plugin should have its own tests' package. *NOTE:* `tests` folder should have an `__init__.py` file.

7. There may be some cases where you might want to auto-load some of your modules when the plugin is installed. This is especially true for `data-plugins` and `type-plugins`.
In such a case, you can add a special directive in the `setup.py` which will instruct Flytekit to automatically load the prescribed modules.

   Following shows an excerpt from the `flytekit-data-fsspec` plugin's setup.py file.

    ```python
    setup(
        entry_points={"flytekit.plugins": [f"{PLUGIN_NAME}=flytekitplugins.{PLUGIN_NAME}"]},
    )

    ```

### Flytekit Version Pinning
Currently we advocate pinning to minor releases of flytekit. To bump the pins across the board, `cd plugins/` and then
update the command below with the appropriate range and run

```bash
for f in $(ls **/setup.py); do sed -i "s/flytekit>.*,<1.1/flytekit>=1.1.0b0,<1.2/" $f; done
```

Try using `gsed` instead of `sed` if you are on a Mac. Also this only works of course for setup files that start with the version in your sed command. There may be plugins that have different pins to start out with.

## References 📚
- Example of a simple Python task that allows adding only Python side functionality: [flytekit-greatexpectations](./flytekit-greatexpectations/)
- Example of a TypeTransformer or a Type Plugin: [flytekit-pandera](./flytekit-pandera/). These plugins add new types to Flyte and tell Flyte how to transform them and add additional features through types. Flyte is a multi-lang system, and type transformers allow marshaling between Flytekit and backend and other languages.
- Example of TaskTemplate plugin which also allows plugin writers to supply a prebuilt container for runtime: [flytekit-sqlalchemy](./flytekit-sqlalchemy/)
- Example of a SQL backend plugin where the actual query invocation is done by a backend plugin: [flytekit-snowflake](./flytekit-snowflake/)
- Example of a Meta plugin that can wrap other tasks: [flytekit-papermill](./flytekit-papermill/)
- Example of a plugin that modifies the execution command: [flytekit-spark](./flytekit-spark/) OR [flytekit-aws-sagemaker](./flytekit-aws-sagemaker/)
- Example that allows executing the user container with some other context modifications: [flytekit-kf-tensorflow](./flytekit-kf-tensorflow/)
- Example of a Persistence Plugin that allows data to be stored to different persistence layers: [flytekit-data-fsspec](./flytekit-data-fsspec/)
