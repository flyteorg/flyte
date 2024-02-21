# Flytekit Dask Plugin

Flyte can execute `dask` jobs natively on a Kubernetes Cluster, which manages the virtual `dask` cluster's lifecycle
(spin-up and tear down). It leverages the open-source Kubernetes Dask Operator and can be enabled without signing up
for any service. This is like running a transient (ephemeral) `dask` cluster - a type of cluster spun up for a specific
task and torn down after completion. This helps in making sure that the Python environment is the same on the job-runner
(driver), scheduler and the workers.

To install the plugin, run the following command:

```bash
pip install flytekitplugins-dask
```

To configure Dask in the Flyte deployment's backed, follow
[step 1](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/kubernetes/k8s_dask/index.html#step-1-deploy-the-dask-plugin-in-the-flyte-backend)
and
[step 2](https://docs.flyte.org/projects/cookbook/en/latest/auto/auto/integrations/kubernetes/k8s_dask/index.html#step-2-environment-setup)

An [example](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/kubernetes/k8s_dask/index.html)
can be found in the documentation.
