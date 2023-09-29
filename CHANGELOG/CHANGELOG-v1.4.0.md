# Flyte 1.4 release

The main features of the 1.4 release are:
- Support for `PodTemplate` at the task-level
- Revamped auth system in flytekit

As python 3.7 [reached](https://endoflife.date/python) EOL support in December of 2022, we dropped support for that version on this release.

## Platform

### Support for `PodTemplate` at the task-level. 
Users can now define [PodTemplate](https://docs.flyte.org/en/latest/deployment/configuration/general.html#using-default-k8s-podtemplates) as part of the definition of a task. For example, note how we have access a full [V1PodSpec](https://github.com/kubernetes-client/python/blob/master/kubernetes/docs/V1PodSpec.md) as part of the task definition:

```python=
@task(
    pod_template=PodTemplate(
        primary_container_name="primary",
        labels={"lKeyA": "lValA", "lKeyB": "lValB"},
        annotations={"aKeyA": "aValA", "aKeyB": "aValB"},
        pod_spec=V1PodSpec(
            containers=[
                V1Container(
                    name="primary",
                    image="repo/placeholderImage:0.0.0",
                    command="echo",
                    args=["wow"],
                    resources=V1ResourceRequirements(limits={"cpu": "999", "gpu": "999"}),
                    env=[V1EnvVar(name="eKeyC", value="eValC"), V1EnvVar(name="eKeyD", value="eValD")],
                ),
            ],
            volumes=[V1Volume(name="volume")],
            tolerations=[
                V1Toleration(
                    key="num-gpus",
                    operator="Equal",
                    value=1,
                    effect="NoSchedule",
                ),
            ],
        )
    )
)
def t1(i: str):
    ...
```

We are working on more examples in our documentation. Stay tuned!

## Flytekit

As promised in https://github.com/flyteorg/flytekit/releases/tag/v1.3.0, we're backporting important changes to the 1.2.x release branch. In the past month we had 2 releases: https://github.com/flyteorg/flytekit/releases/tag/v1.2.8 and https://github.com/flyteorg/flytekit/releases/tag/v1.2.9.

Here's some of the highlights of this release. For a full changelog please visit https://github.com/flyteorg/flytekit/releases/tag/v1.4.0.

### Revamped auth system
In https://github.com/flyteorg/flytekit/pull/1458 we introduced a new OAuth2 handling system based on [client-side grpc interceptors](https://grpc.github.io/grpc/python/grpc.html#client-side-interceptor). 

## New sandbox features
In this new release `flytectl demo` brings the following new features:
- Support for specifying extra configuration for Flyte
- Support for specifying extra cluster resource templates for bootstrapping new namespaces
- Sandbox state (DB, buckets) is now persistent across restarts and upgrades

## Flyteconsole
- [Added domain settings for project dashboard](https://github.com/flyteorg/flyteconsole/pull/689)
- [Added support for ApprovedCondition for GateNodes](https://github.com/flyteorg/flyteconsole/pull/688)
- [Performance refactor for viewing dynamic nodes](https://github.com/flyteorg/flyteconsole/pull/680)
