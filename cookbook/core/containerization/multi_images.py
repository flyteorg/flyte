"""
.. _hosted_multi_images:

Multiple Container Images in a Single Workflow
----------------------------------------------

.. tags:: Containerization, Intermediate

When working locally, it is recommended to install all requirements of your project locally (maybe in a single virtual environment). It gets complicated when you want to deploy your code to a remote
environment since most tasks in Flyte (function tasks) are deployed using a Docker Container.

For every :py:class:`flytekit.PythonFunctionTask` type task or simply a task decorated with the ``@task`` decorator, users can supply rules of how the container image should be bound. By default, flytekit binds one container image, i.e., the ``default`` image to all tasks.
To alter the image, use the ``container_image`` parameter available in the :py:func:`flytekit.task` decorator. Any one of the following is an acceptable:

#. Image reference is specified, but the version is derived from the default image version ``container_image="docker.io/redis:{{.image.default.version}}``
#. Both the FQN and the version are derived from the default image ``container_image="{{.image.default.fqn}}:spark-{{.image.default.version}}``

.. warning:

   To use the image, push a container image that matches the new name described.

If you wish to build and push your Docker image to GHCR, follow `this <https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry>`_.
If you wish to build and push your Docker image to Dockerhub through your account, follow the below steps:

1. Create an account with `Dockerhub <https://hub.docker.com/signup>`__.
2. Build a Docker image using the Dockerfile:

    .. code-block::

        docker build . -f ./<dockerfile-folder>/<dockerfile-name> -t <your-name>/<docker-image-name>:<version>
3. Once the Docker image is built, login to your Dockerhub account from the CLI:

    .. code-block::

        docker login
4. It prompts you to enter the username and the password.
5. Push the Docker image to Dockerhub:

.. code-block::

   docker push <your-dockerhub-name>/<docker-image-name>

Example: Suppose your Dockerfile is named `Dockerfile.prediction`, Docker image name is `multi-images-prediction` with the `latest` version, your build and push commands would look like:

.. code-block::

    docker build -f ./path-to-dockerfile/Dockerfile.prediction -t username/multi-images-prediction:latest
    docker login
    docker push dockerhub_name/multi-images-prediction

.. tip::

    Sometimes, ``docker login`` may not be successful. In such cases, execute ``docker logout`` and ``docker login``.

Let's dive into the example.
"""
# %%
# Import the necessary dependencies.
from typing import NamedTuple

import pandas as pd
from flytekit import task, workflow
from sklearn.model_selection import train_test_split
from sklearn.svm import SVC

split_data = NamedTuple(
    "split_data",
    train_features=pd.DataFrame,
    test_features=pd.DataFrame,
    train_labels=pd.DataFrame,
    test_labels=pd.DataFrame,
)

dataset_url = "https://raw.githubusercontent.com/harika-bonthu/SupportVectorClassifier/main/datasets_229906_491820_Fish.csv"

# %%
# Define a task that fetches data and splits the data into train and test sets.
@task(container_image="{{.image.trainer.fqn }}:{{.image.trainer.version}}")
def svm_trainer() -> split_data:
    fish_data = pd.read_csv(dataset_url)
    X = fish_data.drop(["Species"], axis="columns")
    y = fish_data.Species
    X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.31)
    y_train = y_train.to_frame()
    y_test = y_test.to_frame()

    return split_data(
        train_features=X_train,
        test_features=X_test,
        train_labels=y_train,
        test_labels=y_test,
    )


# %%
# .. note ::
#
#     To use your own Docker image, replace the value of `container_image` with the fully qualified name that identifies where the image has been pushed.
#     The recommended usage (specified in the example) is:
#
#     ``container_image= "{{.image.default.fqn}}:{{.image.default.version}}"``
#
#     #. ``image`` refers to the name of the image in the image configuration. The name ``default`` is a reserved keyword and will automatically apply to the default image name for this repository.
#     #. ``fqn`` refers to the fully qualified name of the image. For example, it includes the repository and domain url of the image. Example: ``docker.io/my_repo/xyz``.
#     #. ``version`` refers to the tag of the image. For example: `latest`, or `python-3.8` etc. If the `container_image` is not specified then the default configured image for the project is used.
#
#     The images themselves are parameterizable in the config file in the following format:
#
#     ``{{.image.<name>.<attribute>}}``


# %%
# Define another task that trains the model on the data and computes the accuracy score.
@task(container_image="{{.image.predictor.fqn }}:{{.image.predictor.version}}")
def svm_predictor(
    X_train: pd.DataFrame,
    X_test: pd.DataFrame,
    y_train: pd.DataFrame,
    y_test: pd.DataFrame,
) -> float:
    model = SVC(kernel="linear", C=1)
    model.fit(X_train, y_train.values.ravel())
    svm_pred = model.predict(X_test)
    accuracy_score = float(model.score(X_test, y_test.values.ravel()))
    return accuracy_score


# %%
# Define a workflow.
@workflow
def my_workflow() -> float:
    train_model = svm_trainer()
    svm_accuracy = svm_predictor(
        X_train=train_model.train_features,
        X_test=train_model.test_features,
        y_train=train_model.train_labels,
        y_test=train_model.test_labels,
    )
    return svm_accuracy


if __name__ == "__main__":
    print(f"Running my_workflow(), accuracy: {my_workflow()}")

# %%
# Configuring sandbox.config
# ==========================
#
# The container image referenced in the tasks above is specified in the sandbox.config file. Provided a name to every Docker image, and reference that in ``container_image``. In this example, we have used the ``core`` image for both the tasks for illustration purposes.
#
# sandbox.config
# ^^^^^^^^^^^^^^
# .. literalinclude::  ../../../../core/sandbox.config
