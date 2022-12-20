"""

.. _secrets:

=======================
Using Secrets in a Task
=======================

.. tags:: Kubernetes, Intermediate

Flyte supports running a variety of tasks, from containers to SQL queries and service calls. For Flyte-run
containers to request and access secrets, Flyte provides a native Secret construct.

For a simple task that launches a Pod, the flow would look similar to this:

.. image:: https://mermaid.ink/img/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4gICAgUHJvcGVsbGVyLT4-K1BsdWdpbnM6IENyZWF0ZSBLOHMgUmVzb3VyY2VcbiAgICBQbHVnaW5zLT4-LVByb3BlbGxlcjogUmVzb3VyY2UgT2JqZWN0XG4gICAgUHJvcGVsbGVyLT4-K1Byb3BlbGxlcjogU2V0IExhYmVscyAmIEFubm90YXRpb25zXG4gICAgUHJvcGVsbGVyLT4-K0FwaVNlcnZlcjogQ3JlYXRlIE9iamVjdCAoZS5nLiBQb2QpXG4gICAgQXBpU2VydmVyLT4-K1BvZCBXZWJob29rOiAvbXV0YXRlXG4gICAgUG9kIFdlYmhvb2stPj4rUG9kIFdlYmhvb2s6IExvb2t1cCBnbG9iYWxzXG4gICAgUG9kIFdlYmhvb2stPj4rUG9kIFdlYmhvb2s6IEluamVjdCBTZWNyZXQgQW5ub3RhdGlvbnMgKGUuZy4gSzhzLCBWYXVsdC4uLiBldGMuKVxuICAgIFBvZCBXZWJob29rLT4-LUFwaVNlcnZlcjogTXV0YXRlZCBQb2RcbiAgICBcbiAgICAgICAgICAgICIsIm1lcm1haWQiOnt9LCJ1cGRhdGVFZGl0b3IiOmZhbHNlfQ
   :target: https://mermaid.ink/img/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4gICAgUHJvcGVsbGVyLT4-K1BsdWdpbnM6IENyZWF0ZSBLOHMgUmVzb3VyY2VcbiAgICBQbHVnaW5zLT4-LVByb3BlbGxlcjogUmVzb3VyY2UgT2JqZWN0XG4gICAgUHJvcGVsbGVyLT4-K1Byb3BlbGxlcjogU2V0IExhYmVscyAmIEFubm90YXRpb25zXG4gICAgUHJvcGVsbGVyLT4-K0FwaVNlcnZlcjogQ3JlYXRlIE9iamVjdCAoZS5nLiBQb2QpXG4gICAgQXBpU2VydmVyLT4-K1BvZCBXZWJob29rOiAvbXV0YXRlXG4gICAgUG9kIFdlYmhvb2stPj4rUG9kIFdlYmhvb2s6IExvb2t1cCBnbG9iYWxzXG4gICAgUG9kIFdlYmhvb2stPj4rUG9kIFdlYmhvb2s6IEluamVjdCBTZWNyZXQgQW5ub3RhdGlvbnMgKGUuZy4gSzhzLCBWYXVsdC4uLiBldGMuKVxuICAgIFBvZCBXZWJob29rLT4-LUFwaVNlcnZlcjogTXV0YXRlZCBQb2RcbiAgICBcbiAgICAgICAgICAgICIsIm1lcm1haWQiOnt9LCJ1cGRhdGVFZGl0b3IiOmZhbHNlfQ

Where:

1. Flyte invokes a plugin to create the K8s object. This can be a Pod or a more complex CRD (e.g. Spark, PyTorch, etc.)

   .. tip:: The plugin ensures that the labels and annotations are passed to any Pod that is spawned due to the creation of the CRD.

2. Flyte applies labels and annotations that are referenced to all secrets the task is requesting access to. Note that secrets are not case sensitive. 
3. Flyte sends a POST request to ApiServer to create the object.
4. Before persisting the Pod, ApiServer invokes all the registered Pod Webhooks. Flyte's Pod Webhook is called.
5. Using the labels and annotiations attached in step 2, Flyte Pod Webhook looks up globally mounted secrets for each of the requested secrets.
6. If found, the Pod Webhook mounts them directly in the Pod. If not found, the Pod Webhook injects the appropriate annotations to load the secrets for K8s (or Vault or Confidant or any secret management system plugin configured) into the task pod.

Once the secret is injected into the task pod, Flytekit can read it using the secret manager (see examples below).

The webhook is included in all overlays in the Flytekit repo. The deployment file creates (mainly) two things; a Job and a Deployment.

1) ``flyte-pod-webhook-secrets`` Job: This job runs ``flytepropeller webhook init-certs`` command that issues self-signed CA Certificate as well as a derived TLS certificate and its private key. Ensure that the private key is in lower case, that is, ``my_token`` in contrast to ``MY_TOKEN``. It stores them into a new secret ``flyte-pod-webhook-secret``.
2) ``flyte-pod-webhook`` Deployment: This deployment creates the Webhook pod which creates a MutatingWebhookConfiguration on startup. This serves as the registration contract with the ApiServer to know about the Webhook before it starts serving traffic.

Secret Discovery
================

Flyte identifies secrets using a secret group and a secret key.
In a task decorator you request a secret like this: ``@task(secret_requests=[Secret(group=SECRET_GROUP, key=SECRET_NAME)])``
Flytekit provides a shorthand for loading the requested secret inside a task: ``secret = flytekit.current_context().secrets.get(SECRET_GROUP, SECRET_NAME)``
See the python examples further down for more details on how to request and use secrets in a task.

Flytekit relies on the following environment variables to load secrets (defined `here <https://github.com/flyteorg/flytekit/blob/9d313429c577a919ec0ad4cd397a5db356a1df0d/flytekit/configuration/internal.py#L141-L159>`_). When running tasks and workflows locally you should make sure to store your secrets accordingly or to modify these:

- ``FLYTE_SECRETS_DEFAULT_DIR``: The directory Flytekit searches for secret files, default: ``"/etc/secrets"``
- ``FLYTE_SECRETS_FILE_PREFIX``: a common file prefix for Flyte secrets, default: ``""``
- ``FLYTE_SECRETS_ENV_PREFIX``: a common env var prefix for Flyte secrets, default: ``"_FSEC_"``

When running a workflow on a Flyte cluster, the configured secret manager will use the secret Group and Key to try and retrieve a secret.
If successful, it will make the secret available as either file or environment variable and will if necessary modify the above variables automatically so that the task can load and use the secrets.

Configuring a Secret Management System Plugin into Use
======================================================

When a task requests a secret Flytepropeller will try to retrieve secrets in the following order: 1.) checking for global secrets (secrets mounted as files or environment variables on the flyte-pod-webhook pod) and 2.) checking with an additional configurable secret manager.
Note that the global secrets take precedence over any secret discoverable by the secret manager plugins.

The following additional secret managers are available at the time of writing:

- `K8s secrets <https://kubernetes.io/docs/concepts/configuration/secret/>`_ (default): flyte-pod-webhook will try to look for a K8s secret named after the secret Group and retrieve the value for the secret Key.
- AWS Secret Manager: flyte-pod-webhook will add the AWS Secret Manager sidecar container to a task Pod which will mount the secret.
- `Vault Agent Injector <https://www.vaultproject.io/docs/platform/k8s/injector>`_ : flyte-pod-webhook will annotate the task Pod with the respective Vault annotations that trigger an existing Vault Agent Injector to retrieve the specified secret Key from a vault path defined as secret Group.

You can configure the additional secret manager by defining ``secretManagerType`` to be either 'K8s', 'AWS' or 'Vault' in
the `core config <https://github.com/flyteorg/flyte/blob/master/kustomize/base/single_cluster/headless/config/propeller/core.yaml#L34>`_ of the Flytepropeller.

When using the K8s secret manager plugin (enabled by default), the secrets need to be available in the same namespace as the task execution
(for example `flytesnacks-development`). K8s secrets can be mounted as either files or injected as environment variables into the task pod,
so if you need to make larger files available to the task, then this might be the better option.
Furthermore, this method also allows you to have separate credentials for different domains but still using the same name for the secret.
The `group` of the secret request corresponds to the K8s secret name, while the `name` of the request corresponds to the key of the specific entry in the secret.

When using the Vault secret manager, make sure you have Vault Agent deployed on your cluster (`step-by-step tutorial <https://learn.hashicorp.com/tutorials/vault/kubernetes-sidecar>`_).
Vault secrets can only be mounted as files and will become available under ``"/etc/flyte/secrets/SECRET_GROUP/SECRET_NAME"``. Vault comes with `two versions <https://www.vaultproject.io/docs/secrets/kv>`_ of the key-value secret store.
By default the Vault secret manager will try to retrieve Version 2 secrets. You can specify the KV version by setting ``webhook.vaultSecretManager.kvVersion`` in the configmap. Note that the version number needs to be an explicit string (e.g. ``"1"``).
You can also configure the Vault role under which Flyte will try to read the secret by setting webhook.vaultSecretManager.role (default: ``"flyte"``).


How to Use Secrets Injection in a Task
======================================

This feature is available in Flytekit v0.17.0+. This example explains how a secret can be accessed in a Flyte Task. Flyte provides different types of Secrets, as part of
SecurityContext. But, for users writing python tasks, you can only access ``secure secrets`` either as environment variable
or injected into a file.
"""

# %%
import os
from typing import Tuple

import flytekit

# %%
# Flytekit exposes a type/class called Secrets. It can be imported as follows.
from flytekit import Secret, task, workflow
from flytekit.testing import SecretsManager

# %%
# Secrets consists of a name and an enum that indicates how the secrets will be accessed. If the mounting_requirement is
# not specified then the secret will be injected as an environment variable is possible. Ideally, you need not worry
# about the mounting requirement, just specify the ``Secret.name`` that matches the declared ``secret`` in Flyte backend
#
# Let us declare a secret named user_secret in a secret group ``user-info``. A secret group can have multiple secret
# associated with the group. Optionally it may also have a group_version. The version helps in rotating secrets. If not
# specified the task will always retrieve the latest version. Though not recommended some users may want the task
# version to be bound to a secret version.

SECRET_NAME = "user_secret"
SECRET_GROUP = "user-info"


# %%
# Now declare the secret in the requests. The request tells Flyte to make the secret available to the task. The secret can
# then be accessed inside the task using the :py:class:`flytekit.ExecutionParameters`, through the global flytekit
# context as shown below. At runtime, flytekit looks inside the task pod for an environment variable or a mounted file with
# a predefined name/path and loads the value.
@task(secret_requests=[Secret(group=SECRET_GROUP, key=SECRET_NAME)])
def secret_task() -> str:
    secret_val = flytekit.current_context().secrets.get(SECRET_GROUP, SECRET_NAME)
    # Please do not print the secret value, we are doing so just as a demonstration
    print(secret_val)
    return secret_val


# %%
# .. note::
#
#   - In case of failure to access the secret (it is not found at execution time) an error is raised.
#   - Secrets group and key are required parameters during declaration and usage. Failure to specify will cause a
#     :py:class:`ValueError`
#
# In some cases you may have multiple secrets and sometimes, they maybe grouped as one secret in the SecretStore.
# For example, In Kubernetes secrets, it is possible to nest multiple keys under the same secret.
# In this case, the name would be the actual name of the nested secret, and the group would be the identifier for
# the kubernetes secret.
#
# As an example, define 2 secrets username and password, defined in the group user_info
USERNAME_SECRET = "username"
PASSWORD_SECRET = "password"


# %%
# The Secret structure allows passing two fields, matching the key and the group, as previously described:
@task(
    secret_requests=[
        Secret(key=USERNAME_SECRET, group=SECRET_GROUP),
        Secret(key=PASSWORD_SECRET, group=SECRET_GROUP),
    ]
)
def user_info_task() -> Tuple[str, str]:
    secret_username = flytekit.current_context().secrets.get(
        SECRET_GROUP, USERNAME_SECRET
    )
    secret_pwd = flytekit.current_context().secrets.get(SECRET_GROUP, PASSWORD_SECRET)
    # Please do not print the secret value, this is just a demonstration.
    print(f"{secret_username}={secret_pwd}")
    return secret_username, secret_pwd


# %%
# It is also possible to enforce Flyte to mount the secret as a file or an environment variable.
# The File type is useful for large secrets that do not fit in environment variables - typically asymmetric
# keys (certs etc). Another reason may be that a dependent library necessitates that the secret be available as a file.
# In these scenarios you can specify the mount_requirement. In the following example we force the mounting to be
# an Env variable
@task(
    secret_requests=[
        Secret(
            group=SECRET_GROUP,
            key=SECRET_NAME,
            mount_requirement=Secret.MountType.ENV_VAR,
        )
    ]
)
def secret_file_task() -> Tuple[str, str]:
    # SM here is a handle to the secrets manager
    sm = flytekit.current_context().secrets
    f = sm.get_secrets_file(SECRET_GROUP, SECRET_NAME)
    secret_val = sm.get(SECRET_GROUP, SECRET_NAME)
    # returning the filename and the secret_val
    return f, secret_val


# %%
# These tasks can be used in your workflow as usual
@workflow
def my_secret_workflow() -> Tuple[str, str, str, str, str]:
    x = secret_task()
    y, z = user_info_task()
    f, s = secret_file_task()
    return x, y, z, f, s


# %%
# The simplest way to test Secret accessibility is to export the secret as an environment variable. There are some
# helper methods available to do so

if __name__ == "__main__":
    sec = SecretsManager()
    os.environ[sec.get_secrets_env_var(SECRET_GROUP, SECRET_NAME)] = "value"
    os.environ[
        sec.get_secrets_env_var(SECRET_GROUP, USERNAME_SECRET)
    ] = "username_value"
    os.environ[
        sec.get_secrets_env_var(SECRET_GROUP, PASSWORD_SECRET)
    ] = "password_value"
    x, y, z, f, s = my_secret_workflow()
    assert x == "value"
    assert y == "username_value"
    assert z == "password_value"
    assert f == sec.get_secrets_file(SECRET_GROUP, SECRET_NAME)
    assert s == "value"

# %%
# Scaling the Webhook
# ===================
#
# Vertical Scaling
# -----------------
#
# To scale the Webhook to be able to process the number/rate of pods you need, you may need to configure a vertical `pod
# autoscaler <https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler>`_.
#
# Horizontal Scaling
# -------------------
#
# The Webhook does not make any external API Requests in response to Pod mutation requests. It should be able to handle traffic
# quickly. For horizontal scaling, adding additional replicas for the Pod in the
# deployment should be sufficient. A single MutatingWebhookConfiguration object will be used, the same TLS certificate
# will be shared across the pods and the Service created will automatically load balance traffic across the available pods.
