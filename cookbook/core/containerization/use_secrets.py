"""

.. _secrets:

Using Secrets in a Task
-----------------------

Flyte supports running a wide variety of tasks, from containers to SQL queries and service calls. In order for Flyte-run
containers to request and access secrets, Flyte provides a native Secret construct.

For a simple task that launches a Pod, the flow will look something like this:

.. image:: https://mermaid.ink/img/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4gICAgUHJvcGVsbGVyLT4-K1BsdWdpbnM6IENyZWF0ZSBLOHMgUmVzb3VyY2VcbiAgICBQbHVnaW5zLT4-LVByb3BlbGxlcjogUmVzb3VyY2UgT2JqZWN0XG4gICAgUHJvcGVsbGVyLT4-K1Byb3BlbGxlcjogU2V0IExhYmVscyAmIEFubm90YXRpb25zXG4gICAgUHJvcGVsbGVyLT4-K0FwaVNlcnZlcjogQ3JlYXRlIE9iamVjdCAoZS5nLiBQb2QpXG4gICAgQXBpU2VydmVyLT4-K1BvZCBXZWJob29rOiAvbXV0YXRlXG4gICAgUG9kIFdlYmhvb2stPj4rUG9kIFdlYmhvb2s6IExvb2t1cCBnbG9iYWxzXG4gICAgUG9kIFdlYmhvb2stPj4rUG9kIFdlYmhvb2s6IEluamVjdCBTZWNyZXQgQW5ub3RhdGlvbnMgKGUuZy4gSzhzLCBWYXVsdC4uLiBldGMuKVxuICAgIFBvZCBXZWJob29rLT4-LUFwaVNlcnZlcjogTXV0YXRlZCBQb2RcbiAgICBcbiAgICAgICAgICAgICIsIm1lcm1haWQiOnt9LCJ1cGRhdGVFZGl0b3IiOmZhbHNlfQ
   :target: https://mermaid.ink/img/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4gICAgUHJvcGVsbGVyLT4-K1BsdWdpbnM6IENyZWF0ZSBLOHMgUmVzb3VyY2VcbiAgICBQbHVnaW5zLT4-LVByb3BlbGxlcjogUmVzb3VyY2UgT2JqZWN0XG4gICAgUHJvcGVsbGVyLT4-K1Byb3BlbGxlcjogU2V0IExhYmVscyAmIEFubm90YXRpb25zXG4gICAgUHJvcGVsbGVyLT4-K0FwaVNlcnZlcjogQ3JlYXRlIE9iamVjdCAoZS5nLiBQb2QpXG4gICAgQXBpU2VydmVyLT4-K1BvZCBXZWJob29rOiAvbXV0YXRlXG4gICAgUG9kIFdlYmhvb2stPj4rUG9kIFdlYmhvb2s6IExvb2t1cCBnbG9iYWxzXG4gICAgUG9kIFdlYmhvb2stPj4rUG9kIFdlYmhvb2s6IEluamVjdCBTZWNyZXQgQW5ub3RhdGlvbnMgKGUuZy4gSzhzLCBWYXVsdC4uLiBldGMuKVxuICAgIFBvZCBXZWJob29rLT4-LUFwaVNlcnZlcjogTXV0YXRlZCBQb2RcbiAgICBcbiAgICAgICAgICAgICIsIm1lcm1haWQiOnt9LCJ1cGRhdGVFZGl0b3IiOmZhbHNlfQ

Where:

1. Flyte invokes a plugin to create the K8s object. This can be a Pod or a more complex CRD (e.g. Spark, PyTorch, etc.)

   .. tip:: The plugin will ensure that labels and annotations are passed through to any Pod that will be spawned due to the creation of the CRD.

2. Flyte will apply labels and annotations that are referenced to all secrets the task is requesting access to.
3. Flyte will send a POST request to ApiServer to create the object.
4. Before persisting the Pod, ApiServer will invoke all registered Pod Webhooks. Flyte's Pod Webhook will be called.
5. Flyte Pod Webhook will then lookup globally mounted secrets for each of the requested secrets. 
6. If found, Pod Webhook will mount them directly in the Pod. If not found, it will inject the appropriate annotations to load the secrets for K8s (or Vault or Confidant or any other secret management system plugin configured) into the Pod.

The webhook is included in all overlays in the Flytekit repo. The deployment file creates (mainly) two things; a Job and a Deployment.

1) flyte-pod-webhook-secrets Job: This job runs ``flytepropeller webhook init-certs`` command that issues self-signed CA Certificate as well as a derived TLS certificate and its private key. It stores them into a new secret ``flyte-pod-webhook-secret``.
2) flyte-pod-webhook Deployment: This deployment creates the Webhook pod which creates a MutatingWebhookConfiguration on startup. This serves as the registration contract with the ApiServer to know about the Webhook before it starts serving traffic.

How to Enable Secrets Injection
###############################

This feature is available in Flytekit v0.17.0+. This example explains how a secret can be accessed in a Flyte Task. Flyte provides different types of Secrets, as part of
SecurityContext. But, for users writing python tasks, you can only access ``secure secrets`` either as environment variable
or injected into a file.
"""

# %%
import os
import flytekit

# %%
# Flytekit exposes a type/class called Secrets. It can be imported as follows.
from flytekit import Secret, task, workflow

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
# Now declare the secret in the requests. The secret can be accessed using the :py:class:`flytekit.ExecutionParameters`,
# through the global flytekit context as shown below
#
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
    secret_requests=[Secret(key=USERNAME_SECRET, group=SECRET_GROUP), Secret(key=PASSWORD_SECRET, group=SECRET_GROUP)])
def user_info_task() -> (str, str):
    secret_username = flytekit.current_context().secrets.get(SECRET_GROUP, USERNAME_SECRET)
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
@task(secret_requests=[Secret(group=SECRET_GROUP, key=SECRET_NAME, mount_requirement=Secret.MountType.ENV_VAR)])
def secret_file_task() -> (str, str):
    # SM here is a handle to the secrets manager
    sm = flytekit.current_context().secrets
    f = sm.get_secrets_file(SECRET_GROUP, SECRET_NAME)
    secret_val = sm.get(SECRET_GROUP, SECRET_NAME)
    # returning the filename and the secret_val
    return f, secret_val


# %%
# These tasks can be used in your workflow as usual
@workflow
def my_secret_workflow() -> (str, str, str, str, str):
    x = secret_task()
    y, z = user_info_task()
    f, s = secret_file_task()
    return x, y, z, f, s


# %%
# The simplest way to test Secret accessibility is to export the secret as an environment variable. There are some
# helper methods available to do so
from flytekit.testing import SecretsManager

if __name__ == "__main__":
    sec = SecretsManager()
    os.environ[sec.get_secrets_env_var(SECRET_GROUP, SECRET_NAME)] = "value"
    os.environ[sec.get_secrets_env_var(SECRET_GROUP, USERNAME_SECRET)] = "username_value"
    os.environ[sec.get_secrets_env_var(SECRET_GROUP, PASSWORD_SECRET)] = "password_value"
    x, y, z, f, s = my_secret_workflow()
    assert x == "value"
    assert y == "username_value"
    assert z == "password_value"
    assert f == sec.get_secrets_file(SECRET_GROUP, SECRET_NAME)
    assert s == "value"

# %%
# Scaling the Webhook
# ###################
#
# Vertical Scaling
# =================
#
# To scale the Webhook to be able to process the number/rate of pods you need, you may need to configure a vertical `pod
# autoscaler <https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler>`_.
#
# Horizontal Scaling
# ==================
#
# The Webhook does not make any external API Requests in response to Pod mutation requests. It should be able to handle traffic
# quickly. For horizontal scaling, adding additional replicas for the Pod in the
# deployment should be sufficient. A single MutatingWebhookConfiguration object will be used, the same TLS certificate
# will be shared across the pods and the Service created will automatically load balance traffic across the available pods.
