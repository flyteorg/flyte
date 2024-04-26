---
jupytext:
  cell_metadata_filter: all
  formats: md:myst
  notebook_metadata_filter: all
  text_representation:
    extension: .md
    format_name: myst
    format_version: 0.13
    jupytext_version: 1.16.1
kernelspec:
  display_name: Python 3
  language: python
  name: python3
language_info:
  codemirror_mode:
    name: ipython
    version: 3
  file_extension: .py
  mimetype: text/x-python
  name: python
  nbconvert_exporter: python
  pygments_lexer: ipython3
  version: 3.9.9
vscode:
  interpreter:
    hash: 93d1c4f33f306e18e1c08a771c972fe86afbedaedb2338666e30a98a5179caac
---

# How to Trigger the Feast Workflow using FlyteRemote

The goal of this notebook is to train a simple [Gaussian Naive Bayes model using sklearn](https://scikit-learn.org/stable/modules/generated/sklearn.naive_bayes.GaussianNB.html) on a modified [Horse-Colic dataset from UCI](https://archive.ics.uci.edu/ml/datasets/Horse+Colic).

The model aims to classify if the lesion of the horse is surgical or not.

Let's get started!

+++

Set the AWS environment variables before importing Flytekit.

```{code-cell} ipython3
import os

os.environ["FLYTE_AWS_ENDPOINT"] = os.environ["FEAST_S3_ENDPOINT_URL"] = "http://localhost:30084/"
os.environ["FLYTE_AWS_ACCESS_KEY_ID"] = os.environ["AWS_ACCESS_KEY_ID"] = "minio"
os.environ["FLYTE_AWS_SECRET_ACCESS_KEY"] = os.environ["AWS_SECRET_ACCESS_KEY"] = "miniostorage"
```

## 01. Register the code

The actual workflow code is auto-documented and rendered using sphinx [here](https://docs.flyte.org/projects/cookbook/en/latest/auto/case_studies/feature_engineering/feast_integration/index.html). We've used [Flytekit](https://docs.flyte.org/projects/flytekit/en/latest/) to express the pipeline in pure Python.

You can use [FlyteConsole](https://github.com/flyteorg/flyteconsole) to launch, monitor, and introspect Flyte executions. However here, let's use [flytekit.remote](https://docs.flyte.org/projects/flytekit/en/latest/design/control_plane.html) to interact with the Flyte backend.

```{code-cell} ipython3
:scrolled: true

from flytekit.remote import FlyteRemote
from flytekit.configuration import Config

# The `for_sandbox` method instantiates a connection to the demo cluster.
remote = FlyteRemote(
    config=Config.for_sandbox(),
    default_project="flytesnacks",
    default_domain="development"
)
```

The ``register_script`` method can be used to register the workflow.

```{code-cell} ipython3
from flytekit.configuration import ImageConfig

from feast_workflow import feast_workflow

wf = remote.register_script(
    feast_workflow,
    image_config=ImageConfig.from_images(
        "ghcr.io/flyteorg/flytecookbook:feast_integration-latest"
    ),
    version="v2",
    source_path="../",
    module_name="feast_workflow",
)
```

## 02: Launch an execution

+++

FlyteRemote provides convenient methods to retrieve version of the pipeline from the remote server.

**NOTE**: It is possible to get a specific version of the workflow and trigger a launch for that, but let's just get the latest.

```{code-cell} ipython3
lp = remote.fetch_launch_plan(name="feast_integration.feast_workflow.feast_workflow")
lp.id.version
```

The ``execute`` method can be used to execute a Flyte entity — a launch plan in our case.

```{code-cell} ipython3
execution = remote.execute(
    lp,
    inputs={"num_features_univariate": 5},
    wait=True
)
```

## 03. Sync an execution

You can sync an execution to retrieve the workflow's outputs. ``sync_nodes`` is set to True to retrieve the intermediary nodes' outputs as well.

**NOTE**: It is possible to fetch an existing execution or simply retrieve an already commenced execution. Also, if you launch an execution with the same name, Flyte will respect that and not restart a new execution!

```{code-cell} ipython3
from flytekit.models.core.execution import WorkflowExecutionPhase

synced_execution = remote.sync(execution, sync_nodes=True)
print(f"Execution {synced_execution.id.name} is in {WorkflowExecutionPhase.enum_to_string(synced_execution.closure.phase)} phase")
```

## 04. Retrieve the output

Fetch the model and the model prediction.

```{code-cell} ipython3
model = synced_execution.outputs["o0"]
prediction = synced_execution.outputs["o1"]
prediction
```

**NOTE**: The output model is available locally as a JobLibSerialized file, which can be downloaded and loaded.

```{code-cell} ipython3
:scrolled: true

model
```

Fetch the ``repo_config``.

```{code-cell} ipython3
repo_config = synced_execution.node_executions["n0"].outputs["o0"]
```

## 05. Generate predictions

Reuse the `predict` function from the workflow to generate predictions — Flytekit will automatically manage the IO for you!

+++

### Load features from the online feature store

```{code-cell} ipython3
import os

from feast_workflow import predict, FEAST_FEATURES, retrieve_online

inference_point = retrieve_online(
    repo_config=repo_config,
    online_store=synced_execution.node_executions["n4"].outputs["o0"],
    data_point=533738,
)
inference_point
```

### Generate a prediction

```{code-cell} ipython3
predict(model_ser=model, features=inference_point)
```
