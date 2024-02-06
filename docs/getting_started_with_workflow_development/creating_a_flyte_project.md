# Creating a Flyte project

## About Flyte projects

A Flyte project is a directory containing task and workflow code, internal Python source code, configuration files, and other artifacts needed to package up your code so that it can be registered to a Flyte cluster.

## Prerequisites

* Follow the steps in {doc}`"Installing development tools" <installing_development_tools>`
* (Optional, but recommended) Install [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)

## Steps

### 1. Activate your Python virtual environment

If you are using conda or another Python virtual environment manager, first, activate the virtual environment you will use to manage dependencies for your Flyte project:

```{prompt} bash $
conda activate flyte-example
```

### 2. Initialize your Flyte project

Next, initialize your Flyte project. The [flytekit-python-template GitHub repository](https://github.com/flyteorg/flytekit-python-template) contains Flyte project templates with sample code that you can run as is or modify to suit your needs.

In this example, we will initialize the [basic-example-imagespec project template](https://github.com/flyteorg/flytekit-python-template/tree/main/basic-example-imagespec).

```{prompt} bash $
pyflyte init my_project
```

:::{note}

To initialize a Flyte project with a different template, use the `--template` parameter:

`pyflyte init --template hello-world hello-world`
:::

### 3. Install additional requirements

After initializing your Flyte project, you will need to install requirements listed in `requirements.txt`:

```{prompt} bash $
cd my_project
pip install -r requirements.txt
```

### 4. (Optional) Version your Flyte project with git

We highly recommend putting your Flyte project code under version control. To do so, initialize a git repository in the Flyte project directory:

```{prompt} bash $
git init
```

```{note}
If you are using a Dockerfile instead of ImageSpec, you will need to initialize a git repository and create at least one commit, since the commit hash is used to tag the image when it is built.
```

### 5. Run your workflow in a local Python environment

To check that your Flyte project was set up correctly, run the workflow in a local Python environment:

```{prompt} bash $
cd workflows
pyflyte run example.py wf
```

## Next steps

To learn about the parts of a Flyte project, including tasks and workflows, see {doc}`"Flyte project components" <flyte_project_components>`.

To run the workflow in your Flyte project in a local Flyte cluster, see the {ref}`"Running a workflow in a local cluster" <getting_started_running_workflow_local_cluster>` section of {doc}`"Running a workflow locally" <running_a_workflow_locally>`.
