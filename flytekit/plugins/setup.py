import os

import pip
from setuptools import setup
from setuptools.command.develop import develop
from setuptools.command.install import install

PACKAGE_NAME = "flytekitplugins-parent"

__version__ = "0.0.0+develop"

# Please maintain an alphabetical order in the following list
SOURCES = {
    "flytekitplugins-async-fsspec": "flytekit-async-fsspec",
    "flytekitplugins-athena": "flytekit-aws-athena",
    "flytekitplugins-awsbatch": "flytekit-aws-batch",
    "flytekitplugins-awssagemaker": "flytekit-aws-sagemaker",
    "flytekitplugins-bigquery": "flytekit-bigquery",
    "flytekitplugins-dask": "flytekit-dask",
    "flytekitplugins-dbt": "flytekit-dbt",
    "flytekitplugins-deck-standard": "flytekit-deck-standard",
    "flytekitplugins-dolt": "flytekit-dolt",
    "flytekitplugins-duckdb": "flytekit-duckdb",
    "flytekitplugins-data-fsspec": "flytekit-data-fsspec",
    "flytekitplugins-envd": "flytekit-envd",
    "flytekitplugins-great_expectations": "flytekit-greatexpectations",
    "flytekitplugins-hive": "flytekit-hive",
    "flytekitplugins-huggingface": "flytekit-huggingface",
    "flytekitplugins-pod": "flytekit-k8s-pod",
    "flytekitplugins-kfmpi": "flytekit-kf-mpi",
    "flytekitplugins-kfpytorch": "flytekit-kf-pytorch",
    "flytekitplugins-kftensorflow": "flytekit-kf-tensorflow",
    "flytekitplugins-mlflow": "flytekit-mlflow",
    "flytekitplugins-modin": "flytekit-modin",
    "flytekitplugins-onnxscikitlearn": "flytekit-onnx-scikitlearn",
    "flytekitplugins-onnxtensorflow": "flytekit-onnx-tensorflow",
    "flytekitplugins-onnxpytorch": "flytekit-onnx-pytorch",
    "flytekitplugins-pandera": "flytekit-pandera",
    "flytekitplugins-papermill": "flytekit-papermill",
    "flytekitplugins-polars": "flytekit-polars",
    "flytekitplugins-ray": "flytekit-ray",
    "flytekitplugins-snowflake": "flytekit-snowflake",
    "flytekitplugins-spark": "flytekit-spark",
    "flytekitplugins-sqlalchemy": "flytekit-sqlalchemy",
    "flytekitplugins-vaex": "flytekit-vaex",
    "flytekitplugins-whylogs": "flytekit-whylogs",
    "flytekitplugins-flyteinteractive": "flytekit-flyteinteractive",
}


def install_all_plugins(sources, develop=False):
    """
    Use pip to install all plugins
    """
    print("Installing all Flyte plugins in {} mode".format("development" if develop else "normal"))
    wd = os.getcwd()
    for k, v in sources.items():
        try:
            os.chdir(os.path.join(wd, v))
            if develop:
                pip.main(["install", "-e", "."])
            else:
                pip.main(["install", "."])
        except Exception as e:
            print("Oops, something went wrong installing", k)
            print(e)
        finally:
            os.chdir(wd)


class DevelopCmd(develop):
    """Add custom steps for the develop command"""

    def run(self):
        install_all_plugins(SOURCES, develop=True)
        develop.run(self)


class InstallCmd(install):
    """Add custom steps for the install command"""

    def run(self):
        install_all_plugins(SOURCES, develop=False)
        install.run(self)


setup(
    name=PACKAGE_NAME,
    version=__version__,
    author="flyteorg",
    author_email="admin@flyte.org",
    description="This is a microlib package to help install all the plugins",
    license="apache2",
    classifiers=["Private :: Do Not Upload to pypi server"],
    install_requires=[],
    cmdclass={"install": InstallCmd, "develop": DevelopCmd},
)
