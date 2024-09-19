# Flytekit Kubeflow TensorFlow Plugin

This plugin uses the Kubeflow TensorFlow Operator and provides an extremely simplified interface for executing distributed training using various TensorFlow backends.

To install the plugin, run the following command:

```bash
pip install flytekitplugins-kftensorflow
```

## Code Example
To build a TFJob with:
10 workers with restart policy as failed and 2 CPU and 2Gi Memory
1 ps replica with resources the same as task defined resources
1 chief replica with resources the same as task defined resources and restart policy as always
run policy as clean up all pods after job is finished.

You code:
```python
from flytekitplugins.kftensorflow import PS, Chief, CleanPodPolicy, RestartPolicy, RunPolicy, TfJob, Worker

@task(
    task_config=TfJob(
        worker=Worker(
            replicas=5,
            requests=Resources(cpu="2", mem="2Gi"),
            limits=Resources(cpu="2", mem="2Gi"),
            restart_policy=RestartPolicy.FAILURE,
        ),
        ps=PS(replicas=1),
        chief=Chief(replicas=1, restart_policy=RestartPolicy.ALWAYS),
        run_policy=RunPolicy(clean_pod_policy=CleanPodPolicy.RUNNING),
    ),
    image="test_image",
    resources=Resources(cpu="1", mem="1Gi"),
)
def tf_job():
    ...
```


## Upgrade TensorFlow Plugin from V0 to V1
Tensorflow plugin is now updated from v0 to v1 to enable more configuration options.
To migrate from v0 to v1, change the following:
1. Update flytepropeller to v1.6.0
2. Update flytekit version to v1.6.2
3. Update your code from:
    ```
    task_config=TfJob(num_workers=10, num_ps_replicas=1, num_chief_replicas=1),
    ```
    to:
    ```
    task_config=TfJob(worker=Worker(replicas=10), ps=PS(replicas=1), chief=Chief(replicas=1)),
    ```
