"""
.. _hosted_multi_images:

Multiple Container Images in a Single Workflow
----------------------------------------------

When working locally, it is recommended to install all requirements of your project locally (maybe in a single virtual environment). It gets complicated when you want to deploy your code to a remote
environment. This is because most tasks in Flyte (function tasks) are deployed using a Docker Container. 

A Docker container allows you to create an expected environment for your tasks. It is also possible to build a single container image with all your dependencies, but sometimes this is complicated and impractical.

Here are the reasons why it is complicated and not recommended:

#. All the dependencies in one container increase the size of the container image.
#. Some task executions like Spark, SageMaker-based Training, and deep learning use GPUs that need specific runtime configurations. For example,
   
   - Spark needs JavaVirtualMachine installation and Spark entrypoints to be set
   - NVIDIA drivers and corresponding libraries need to be installed to use GPUs for deep learning. However, these are not required for a CPU
   - SageMaker expects the ENTRYPOINT to be designed to accept its parameters

#. Building a single image may increase the build time for the image itself.

.. note::

   Flyte (Service) by default does not require a workflow to be bound to a single container image. Flytekit offers a simple interface to easily alter the images that should be associated with every task, yet keeping the local execution simple for the user.

For every :py:class:`flytekit.PythonFunctionTask` type task or simply a task that is decorated with the ``@task`` decorator, users can supply rules of how the container image should be bound. By default, flytekit binds one container image, i.e., the ``default`` image to all tasks.
To alter the image, use the ``container_image`` parameter available in the :py:func:`flytekit.task` decorator. Any one of the following is an acceptable:

#. Image reference is specified, but the version is derived from the default image version ``container_image="docker.io/redis:{{.image.default.version}},``
#. Both the FQN and the version are derived from the default image ``container_image="{{.image.default.fqn}}:spark-{{.image.default.version}},``

The images themselves are parameterizable in the config in the following format:
 ``{{.image.<name>.<attribute>}}``

- ``name`` refers to the name of the image in the image configuration. The name ``default`` is a reserved keyword and will automatically apply to the default image name for this repository.
- ``fqn`` refers to the fully qualified name of the image. For example, it includes the repository and domain url of the image. Example: ``docker.io/my_repo/xyz``.
- ``version`` refers to the tag of the image. For example: `latest`, or `python-3.8` etc. If the `container_image` is not specified then the default configured image for the project is used.

.. note::

    The default image (name + version) is always ``{{.image.default.fqn}}:{{.image.default.version}}``

.. warning:

   To be able to use the image, push a container image that matches the new name described.

If you wish to build and push your Docker image to GHCR, follow `this <https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry>`_.
If you wish to build and push your Docker image to Dockerhub through your account, follow the below steps:

1. Create an account with `Dockerhub <https://hub.docker.com/signup>`__.
2. Build a Docker image using the Dockerfile:
    
.. code-block::

   docker build . -f ./<dockerfile-folder>/<dockerfile-name> -t <your-name>/<docker-image-name>:<version>
3. Once the Docker image is built, login to your Dockerhub account from the CLI:
    
.. code-block::

   docker login
4. It will prompt you to enter the username and the password.
5. Push the Docker image to Dockerhub:
    
.. code-block::
        
   docker push <your-dockerhub-name>/<docker-image-name>

Example: Suppose your Dockerfile is named `Dockerfile.prediction`, Docker image name is `multi-images-prediction` with the `latest` version, your build and push commands would look like:

.. code-block::

   docker build -f ./path-to-dockerfile/Dockerfile.prediction -t username/multi-images-prediction:latest
   docker login
   docker push dockerhub_name/multi-images-prediction

.. tip::
   
   Sometimes, ``docker login`` may not be successful. In such a case, execute ``docker logout`` and ``docker login``.

Let us understand how multiple images can be used within a single workflow.
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
@task(
    container_image="ghcr.io/flyteorg/flytecookbook:core-with-sklearn-baa17ccf39aa667c5950bd713a4366ce7d5fccaf7f85e6be8c07fe4b522f92c3"
)
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
#     To use your own Docker image, replace the value of `container_image` with the fully qualified name that identifies where the image has been pushed. 
#     One pattern has been specified in the task itself, i.e., specifying the Docker image URI. The recommended usage is:
#
#     ``container_image="{{.image.default.fqn}}:multi-images-preprocess-{{.image.default.version}}"``

# %%
# Define another task that trains the model on the data and computes the accuracy score.
@task(
    container_image="ghcr.io/flyteorg/flytecookbook:multi-image-predict-98b125fd57d20594026941c2ebe7ef662e5acb7d6423660a65f493ca2d9aa267"
)
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
    print(f"Running my_workflow(), accuracy : { my_workflow() }")

# %%
# .. note::
#     Notice that the two task annotators have two different `container_image`s specified.
