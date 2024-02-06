# Running a workflow locally

You can run a workflow locally in two ways:

* **{ref}`In a local Python environment <getting_started_running_workflow_local_python_environment>`:** To develop and test your code quickly without the overhead of setting up a local Flyte cluster, you can run your workflow in your local Python environment.
* **{ref}`In a local Flyte cluster <getting_started_running_workflow_local_cluster>`:** To test your code in a more production-like setting, you can run your workflow in a local cluster, such as the demo Flyte cluster.

(getting_started_running_workflow_local_python_environment)=

## Running a workflow in a local Python environment

### Prerequisites

* {doc}`Install development tools <installing_development_tools>`
* {doc}`Create a Flyte project <creating_a_flyte_project>`

### Steps

1. On the command line, navigate to the workflows directory of your Flyte project:
```{prompt} bash $
cd my_project/workflows
```
2. Run the workflow with `pyflyte run`:
```{prompt} bash $
pyflyte run example.py wf
```

:::{note}
While you can run the example file like a Python script with `python example.py`, we recommend using `pyflyte run` instead. To run the file like a Python script, you would have to add a `main` module conditional at the end of the script:
```python
if __name__ == "__main__":
    print(wf())
```

Your code would become even more verbose if you wanted to pass arguments to the workflow:
```python
if __name__ == "__main__":
    from argparse import ArgumentParser

    parser = ArgumentParser()
    parser.add_argument("--name", type=str)

    args = parser.parse_args()
    print(wf(name=args.name))
```
:::

(getting_started_running_workflow_local_cluster)=

## Running a workflow in a local Flyte cluster

### Prerequisites

#### 1. Install development tools

If you have not already done so, follow the steps in {doc}`"Installing development tools" <installing_development_tools>` to install Python, Flytekit, and optionally, conda.

#### 2. Create a Flyte project

If you have not already done so, follow the steps in {doc}`"Creating a Flyte project" <creating_a_flyte_project>` to create a Flyte project.

#### 3. Install Docker

Follow the steps in the [Docker installation guide](https://docs.docker.com/get-docker/) to install Docker.

Flyte supports any [OCI-compatible](https://opencontainers.org/) container technology (like [Podman](https://podman.io/), [LXD](https://linuxcontainers.org/lxd/introduction/), and [Containerd](https://containerd.io/)), but for the purpose of this documentation, `flytectl` uses Docker to start a local Kubernetes cluster that you can interact with on your machine.

#### 4. Install `flytectl`

You must install `flytectl` to start and configure a local Flyte cluster, as well as register workflows to a local or remote Flyte cluster.

````{tabbed} macOS
To use Homebrew, on the command line, run the following:

```{prompt} bash $
brew install flyteorg/homebrew-tap/flytectl
```

To use `curl`, on the command line, run the following:

```{prompt} bash $
curl -sL https://ctl.flyte.org/install | sudo bash -s -- -b /usr/local/bin
```

To download manually, see the [flytectl releases](https://github.com/flyteorg/flytectl/releases).
````

````{tabbed} Linux
To use `curl`, on the command line, run the following:

```{prompt} bash $
curl -sL https://ctl.flyte.org/install | sudo bash -s -- -b /usr/local/bin
```

To download manually, see the [flytectl releases](https://github.com/flyteorg/flytectl/releases).
````

````{tabbed} Windows
To use `curl`, in a Linux shell (such as [WSL](https://learn.microsoft.com/en-us/windows/wsl/install)), on the command line, run the following:

```{prompt} bash $
curl -sL https://ctl.flyte.org/install | sudo bash -s -- -b /usr/local/bin
```

To download manually, see the [flytectl releases](https://github.com/flyteorg/flytectl/releases).
````

### Steps

1. Export the `FLYTECTL_CONFIG` environment variable in your shell:

```{prompt} bash $
export FLYTECTL_CONFIG=~/.flyte/config-sandbox.yaml
```
2. Start the Docker daemon.
3. Start the demo cluster:

```{prompt} bash $
flytectl demo start
```
4. Create a project on the demo cluster to correspond to your local Flyte project:
```{prompt} bash $
flytectl create project \
      --id "my-project" \
      --labels "my-label=my-project" \
      --description "My Flyte project" \
      --name "My project"
```
3. On the command line, navigate to the workflows directory of your Flyte project:
```{prompt} bash $
cd my_project/workflows
```
4. Run the workflow on the Flyte cluster with `pyflyte run` using the `--remote` flag and additional parameters for the project name and domain. In this example, you can also optionally pass a `name` parameter to the workflow:
```{prompt} bash $
pyflyte run --remote -p my-project -d development example.py wf --name Ada
```

You should see a URL to the workflow execution on your demo Flyte cluster, where `<execution_name>` is a unique identifier for the workflow execution:

```{prompt} bash $
Go to http://localhost:30080/console/projects/flytesnacks/domains/development/executions/<execution_name> to see execution in the console.
```

### Inspecting a workflow run in the FlyteConsole web interface

You can inspect the results of a workflow run by navigating to the URL produced by `pyflyte run` for the workflow execution. You should see FlyteConsole, the web interface used to manage Flyte entities such as tasks, workflows, and executions. The default execution view shows the list of tasks executing in sequential order.

![Landing page of Flyte UI showing two successful tasks run for one workflow execution, along with Nodes, Graph, and Timeline view switcher links](https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/getting_started/flyteconsole_default.png)

#### Task panel

Clicking on a single task will open a panel that shows task logs, inputs, outputs, and metadata.

![Single task panel showing task logs, rerun task button, and executions, inputs, outputs, and task metadata sections](https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/getting_started/flyteconsole_task_panel.png)

#### Graph view

The **Graph** view shows the execution graph of the workflow, providing visual information about the topology of the graph and the state of each node as the workflow progresses.

![Graph view of single workflow execution showing directed acyclic graph of start node, say_hello_node, greeting_length node, and end node](https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/getting_started/flyteconsole_graph_view.png)
