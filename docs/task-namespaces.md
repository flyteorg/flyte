# Select a task's Kubernetes namespace

A task can declare the existing namespace where its Kubernetes workload should
run. For example, with the companion Flyte SDK change:

```python
import flyte

env = flyte.TaskEnvironment(name="training", image="my-training-image:latest")

@env.task(namespace="gpu-user-123")
async def train() -> str:
    return "running in gpu-user-123"
```

The SDK also accepts an environment default (`TaskEnvironment(namespace=...)`),
a task call override (`train.override(namespace=...)`), and `namespace=...` on a
`ContainerTask`. Namespace names must be valid Kubernetes DNS labels of at most
63 characters. Omitting the namespace keeps the existing placement.

An operator must provision the namespace and allow it in the executor:

```yaml
executor:
  allowedTaskNamespaces:
    - gpu-user-123
```

With either Helm chart, this can be supplied as an inline configuration:

```yaml
configuration:
  inline:
    executor:
      allowedTaskNamespaces:
        - gpu-user-123
```

The default empty allowlist prevents tasks from leaving their TaskAction's
namespace. `"*"` permits any namespace the executor's Kubernetes credentials
can access. This is a deployment-wide allowlist, not per-user authorization.

TaskActions remain in the control namespace. Kubernetes workload objects use
`TaskTemplate.config["kubernetes.namespace"]`. Children can declare a different
namespace from the parent task; unset children use the normal default.
Cross-namespace pods carry annotations identifying their TaskAction, which keep
watch events routed to the controller. Cross-namespace owner references are
omitted; the TaskAction finalizer deletes the workload when the action is deleted,
including retained terminal workloads. Ordinary same-namespace owner references
and cleanup behavior remain unchanged.

The executor's permissions must cover the chosen namespaces. Resource quotas,
network policy, service accounts, registry credentials and namespace-local
secrets apply there. The feature does not create namespaces or copy credentials.
A missing or terminating namespace fails the task explicitly. Embedded
cross-namespace env-var secrets retain the existing requirement for
`webhook.embeddedSecretManagerConfig.envVarsFromValues`; enabling it puts secret
values in Pod specifications. File injection follows its existing behavior.

The Kubernetes API integration tests verify workload placement, namespace
validation and allowlisting, terminal workload cleanup, missing namespaces, and
a one-GPU quota admitting the first Pod and rejecting the second. Envtest does
not start containers or require GPU hardware.
