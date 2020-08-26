[Back to Plugins Menu](..)
# How to create a multi-container task - a.k.a K8s pod

If you're interested in running multiple containers in a task, if for example you need to run a sidecar to fetch
or orchestrate data, handle additional logging or monitoring, or if you want to run a pod with it's complete set of
possibilities- mounting volumes etc then use sidecar ``sidecar_task``
can be leveraged to fully customize the pod spec used to run your task.


## Sidecar Example Workflows
Flyte supports both `sidecar_task` and `dynamic_sidecar_task` tasks:

1. [Sidecar Workflow](sidecar.py)
2. [Dynamic Sidecar Workflow](dynamic_sidecar.py)