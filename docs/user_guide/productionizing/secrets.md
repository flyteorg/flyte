(secrets)=

# Secrets

```{eval-rst}
.. tags:: Kubernetes, Intermediate
```

Flyte supports running a variety of tasks, from containers to SQL queries and
service calls, and it provides a native Secret construct to request and access
secrets.

This example explains how you can access secrets in a Flyte Task. Flyte provides
different types of secrets, but for users writing Python tasks, you can only access
secure secrets either as environment variables or as a file injected into the
running container.

## Creating secrets with a secrets manager

:::{admonition} Prerequisites
:class: important

- Install [kubectl](https://kubernetes.io/docs/tasks/tools/).
- Have access to a Flyte cluster, for e.g. with `flytectl demo start` as
  described {ref}`here <getting_started_running_workflow_local_cluster>`.
:::

The first step to using secrets in Flyte is to create one on the backend.
By default, Flyte uses the K8s-native secrets manager, which we'll use in this
example, but you can also {ref}`configure different secret managers <configure_secret_management>`.

First, we use `kubectl` to create a secret called `user-info` with a
`user_secret` key:

```{eval-rst}
.. prompt:: bash $

   kubectl create secret -n <project>-<domain> generic user-info --from-literal=user_secret=mysecret
```

:::{note}
Be sure to specify the correct Kubernetes namespace when creating a secret. If you plan on accessing
the secret in the `flytesnacks` project under the `development` domain, replace `<project>-<domain>`
with `flytesnacks-development`. This is because secrets need to be in the same namespace as the
workflow execution.
:::

:::{important}
The imperative command above is useful for creating secrets in an ad hoc manner,
but it may not be the most secure or sustainable way to do so. You can, however,
define secrets using a [configuration file](https://kubernetes.io/docs/tasks/configmap-secret/managing-secret-using-config-file/)
or tools like [Kustomize](https://kubernetes.io/docs/tasks/configmap-secret/managing-secret-using-kustomize/).
:::

## Using secrets in tasks

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

Once you've defined a secret on the Flyte backend, `flytekit` exposes a class
called {py:class}`~flytekit.Secret`s, which allows you to request a secret
from the configured secret manager:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/use_secrets.py
:caption: productionizing/use_secrets.py
:lines: 1-6, 49-53
```

Secrets consists of `group`, `key`, and `mounting_requirement` arguments,
where a secret group can have multiple secrets associated with it.
If the `mounting_requirement` argument is not specified, the secret will
be injected as an environment variable by default.

In the code below we specify two variables, `SECRET_GROUP` and
`SECRET_NAME`, which maps onto the `user-info` secret that we created
with `kubectl` above, with a key called `user_secret`.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/use_secrets.py
:caption: productionizing/use_secrets.py
:lines: 66-67
```

Now we declare the secret in the `secret_requests` argument of the
{py:func}`@task <flytekit.task>` decorator. The request tells Flyte to make
the secret available to the task.

The secret can then be accessed inside the task using the
{py:class}`~flytekit.ExecutionParameters` object, which is returned by
invoking the {py:func}`flytekit.current_context` function, as shown below.

At runtime, flytekit looks inside the task pod for an environment variable or
a mounted file with a predefined name/path and loads the value.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/use_secrets.py
:caption: productionizing/use_secrets.py
:pyobject: secret_task
```

:::{warning}
Never print secret values! The example above is just for demonstration purposes.
:::

:::{note}
- In case Flyte fails to access the secret, an error is raised.
- The `Secret` group and key are required parameters during declaration
  and usage. Failure to specify will cause a {py:class}`ValueError`.
:::

### Multiple keys grouped into one secret

In some cases you may have multiple secrets and sometimes, they maybe grouped
as one secret in the SecretStore.

For example, In Kubernetes secrets, it is possible to nest multiple keys under
the same secret:

```{eval-rst}
.. prompt:: bash $

   kubectl create secret generic user-info \
       --from-literal=user_secret=mysecret \
       --from-literal=username=my_username \
       --from-literal=password=my_password
```

In this case, the secret group will be `user-info`, with three available
secret keys: `user_secret`, `username`, and `password`:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/use_secrets.py
:caption: productionizing/use_secrets.py
:lines: 107-108
```

The Secret structure allows passing two fields, matching the key and the group, as previously described:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/use_secrets.py
:caption: productionizing/use_secrets.py
:lines: 113-124
```

:::{warning}
Never print secret values! The example above is just for demonstration purposes.
:::

### Mounting secrets as files or environment variables

It is also possible to make Flyte mount the secret as a file or an environment
variable.

The file type is useful for large secrets that do not fit in environment variables,
which are typically asymmetric keys (like certs, etc). Another reason may be that a
dependent library requires the secret to be available as a file.
In these scenarios you can specify the `mount_requirement=Secret.MountType.FILE`.

In the following example we force the mounting to be an environment variable:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/use_secrets.py
:caption: productionizing/use_secrets.py
:lines: 139-158
```

These tasks can be used in your workflow as usual

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/use_secrets.py
:caption: productionizing/use_secrets.py
:pyobject: my_secret_workflow
```

### Testing with mock secrets

The simplest way to test secret accessibility is to export the secret as an
environment variable. There are some helper methods available to do so:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/use_secrets.py
:caption: productionizing/use_secrets.py
:lines: 172-182
```

## Using secrets in task templates

For task types that connect to a remote database, you'll need to specify
secret request as well. For example, for the {py:class}`~flytekitplugins.sqlalchemy.SQLAlchemyTask`
you need to:

1. Specify the `secret_requests` argument.
2. Configure the {py:class}`~flytekitplugins.sqlalchemy.SQLAlchemyConfig` to
   declare which secret maps onto which connection argument.

```python
from flytekit import kwtypes
from flytekitplugins.sqlalchemy import SQLAlchemyTask, SQLAlchemyConfig


# define the secrets
secrets = {
    "username": Secret(group="<SECRET_GROUP>", key="<USERNAME_SECRET>"),
    "password": Secret(group="<SECRET_GROUP>", key="<PASSWORD_SECRET>"),
}


sql_query = SQLAlchemyTask(
    name="sql_query",
    query_template="""SELECT * FROM my_table LIMIT {{ .inputs.limit }}""",
    inputs=kwtypes(limit=int),

    # request secrets
    secret_requests=[*secrets.values()],

    # specify username and password credentials in the configuration
    task_config=SQLAlchemyConfig(
        uri="<DATABASE_URI>",
        secret_connect_args=secrets,
    ),
)
```

:::{note}
Here the `secret_connect_args` map to the
[SQLAlchemy engine configuration](https://docs.sqlalchemy.org/en/20/core/engines.html)
argument names for the username and password.
:::

You can then use the `sql_query` task inside a workflow to grab data and
perform downstream transformations on it.


## How secrets injection works

The rest of this page describes how secrets injection works under the hood.
For a simple task that launches a Pod, the flow would look something like this:

```{image} https://mermaid.ink/img/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4gICAgUHJvcGVsbGVyLT4-K1BsdWdpbnM6IENyZWF0ZSBLOHMgUmVzb3VyY2VcbiAgICBQbHVnaW5zLT4-LVByb3BlbGxlcjogUmVzb3VyY2UgT2JqZWN0XG4gICAgUHJvcGVsbGVyLT4-K1Byb3BlbGxlcjogU2V0IExhYmVscyAmIEFubm90YXRpb25zXG4gICAgUHJvcGVsbGVyLT4-K0FwaVNlcnZlcjogQ3JlYXRlIE9iamVjdCAoZS5nLiBQb2QpXG4gICAgQXBpU2VydmVyLT4-K1BvZCBXZWJob29rOiAvbXV0YXRlXG4gICAgUG9kIFdlYmhvb2stPj4rUG9kIFdlYmhvb2s6IExvb2t1cCBnbG9iYWxzXG4gICAgUG9kIFdlYmhvb2stPj4rUG9kIFdlYmhvb2s6IEluamVjdCBTZWNyZXQgQW5ub3RhdGlvbnMgKGUuZy4gSzhzLCBWYXVsdC4uLiBldGMuKVxuICAgIFBvZCBXZWJob29rLT4-LUFwaVNlcnZlcjogTXV0YXRlZCBQb2RcbiAgICBcbiAgICAgICAgICAgICIsIm1lcm1haWQiOnt9LCJ1cGRhdGVFZGl0b3IiOmZhbHNlfQ
:target: https://mermaid.ink/img/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4gICAgUHJvcGVsbGVyLT4-K1BsdWdpbnM6IENyZWF0ZSBLOHMgUmVzb3VyY2VcbiAgICBQbHVnaW5zLT4-LVByb3BlbGxlcjogUmVzb3VyY2UgT2JqZWN0XG4gICAgUHJvcGVsbGVyLT4-K1Byb3BlbGxlcjogU2V0IExhYmVscyAmIEFubm90YXRpb25zXG4gICAgUHJvcGVsbGVyLT4-K0FwaVNlcnZlcjogQ3JlYXRlIE9iamVjdCAoZS5nLiBQb2QpXG4gICAgQXBpU2VydmVyLT4-K1BvZCBXZWJob29rOiAvbXV0YXRlXG4gICAgUG9kIFdlYmhvb2stPj4rUG9kIFdlYmhvb2s6IExvb2t1cCBnbG9iYWxzXG4gICAgUG9kIFdlYmhvb2stPj4rUG9kIFdlYmhvb2s6IEluamVjdCBTZWNyZXQgQW5ub3RhdGlvbnMgKGUuZy4gSzhzLCBWYXVsdC4uLiBldGMuKVxuICAgIFBvZCBXZWJob29rLT4-LUFwaVNlcnZlcjogTXV0YXRlZCBQb2RcbiAgICBcbiAgICAgICAgICAgICIsIm1lcm1haWQiOnt9LCJ1cGRhdGVFZGl0b3IiOmZhbHNlfQ
```

Breaking down this sequence diagram:

1. Flyte invokes a plugin to create the K8s object. This can be a Pod or a more complex CRD (e.g. Spark, PyTorch, etc.)

   :::{note}
   The plugin ensures that the labels and annotations are passed to any Pod that is spawned due to the creation of the CRD.
   :::

2. Flyte applies labels and annotations that are referenced to all secrets the task is requesting access to. Note that secrets are not case sensitive.

3. Flyte sends a `POST` request to `ApiServer` to create the object.

4. Before persisting the Pod, `ApiServer` invokes all the registered Pod Webhooks and Flyte's Pod Webhook is called.

5. Using the labels and annotiations attached in **step 2**, Flyte Pod Webhook looks up globally mounted secrets for each of the requested secrets.

6. If found, the Pod Webhook mounts them directly in the Pod. If not found, the Pod Webhook injects the appropriate annotations to load the secrets for K8s (or Vault or Confidant or any secret management system plugin configured) into the task pod.

Once the secret is injected into the task pod, Flytekit can read it using the secret manager.

The webhook is included in all overlays in the Flytekit repo. The deployment file creates two things; a **Job** and a **Deployment**.

1. `flyte-pod-webhook-secrets` **Job**: This job runs `flytepropeller webhook init-certs` command that issues self-signed CA Certificate as well as a derived TLS certificate and its private key. Ensure that the private key is in lower case, that is, `my_token` in contrast to `MY_TOKEN`. It stores them into a new secret `flyte-pod-webhook-secret`.
2. `flyte-pod-webhook` **Deployment**: This deployment creates the Webhook pod which creates a MutatingWebhookConfiguration on startup. This serves as the registration contract with the ApiServer to know about the Webhook before it starts serving traffic.

## Secret discovery

Flyte identifies secrets using a secret group and a secret key, which can
be accessed by {py:func}`flytekit.current_context` in the task function
body, as shown in the code examples above.

Flytekit relies on the following environment variables to load secrets (defined [here](https://github.com/flyteorg/flytekit/blob/9d313429c577a919ec0ad4cd397a5db356a1df0d/flytekit/configuration/internal.py#L141-L159)). When running tasks and workflows locally you should make sure to store your secrets accordingly or to modify these:

- `FLYTE_SECRETS_DEFAULT_DIR`: The directory Flytekit searches for secret files. **Default:** `"/etc/secrets"`
- `FLYTE_SECRETS_FILE_PREFIX`: a common file prefix for Flyte secrets. **Default:** `""`
- `FLYTE_SECRETS_ENV_PREFIX`: a common env var prefix for Flyte secrets. **Default:** `"_FSEC_"`

When running a workflow on a Flyte cluster, the configured secret manager will use the secret Group and Key to try and retrieve a secret.
If successful, it will make the secret available as either file or environment variable and will if necessary modify the above variables automatically so that the task can load and use the secrets.

(configure_secret_management)=

## Configuring a secret management system plugin

When a task requests a secret, Flytepropeller will try to retrieve secrets in the following order:

1. Checking for global secrets, i.e. secrets mounted as files or environment variables on the `flyte-pod-webhook` pod
2. Checking with an additional configurable secret manager.

:::{important}
The global secrets take precedence over any secret discoverable by the secret manager plugins.
:::

The following secret managers are available at the time of writing:

- [K8s secrets](https://kubernetes.io/docs/concepts/configuration/secret/#creating-a-secret) (**default**): `flyte-pod-webhook` will try to look for a K8s secret named after the secret Group and retrieve the value for the secret Key.
- [AWS Secret Manager](https://docs.aws.amazon.com/secretsmanager/latest/userguide/create_secret.html): `flyte-pod-webhook` will add the AWS Secret Manager sidecar container to a task Pod which will mount the secret.
- [Vault Agent Injector](https://developer.hashicorp.com/vault/tutorials/getting-started/getting-started-first-secret#write-a-secret) : `flyte-pod-webhook` will annotate the task Pod with the respective Vault annotations that trigger an existing Vault Agent Injector to retrieve the specified secret Key from a vault path defined as secret Group.


When using the K8s secret manager plugin, which is enabled by default, the secrets need to be available in the same namespace as the task execution
(for example `flytesnacks-development`). K8s secrets can be mounted as either files or injected as environment variables into the task pod,
so if you need to make larger files available to the task, then this might be the better option.

Furthermore, this method also allows you to have separate credentials for different domains but still using the same name for the secret.

### AWS secrets manager

When using the AWS secret management plugin, secrets need to be specified by naming them in the format
`<SECRET_GROUP>:<SECRET_KEY>`, where the secret string is a plain-text value, **not** key/value json.

### Vault secrets manager

When using the Vault secret manager, make sure you have Vault Agent deployed on your cluster as described in this [step-by-step tutorial](https://learn.hashicorp.com/tutorials/vault/kubernetes-sidecar).
Vault secrets can only be mounted as files and will become available under `"/etc/flyte/secrets/SECRET_GROUP/SECRET_NAME"`.

Vault comes with various secrets engines. Currently Flyte supports working with both version 1 and 2 of the `Key Vault engine <https://developer.hashicorp.com/vault/docs/secrets/kv>` as well as the `databases secrets engine <https://developer.hashicorp.com/vault/docs/secrets/databases>`.
You can use use the `group_version` parameter to specify which secret backend engine to use. Available choices are: "kv1", "kv2", "db":

#### Requesting secrets with the Vault secret manager

```python
secret = Secret(
    group="<Vault path>",
    key="<Secret key for KV engine>",
    group_version="<kv1|kv2|db>",
)
```

The group parameter is used to specify the path to the secret in the Vault backend. For example, if you have a secret stored in Vault at `"secret/data/flyte/secret"` then the group parameter should be `"secret/data/flyte"`.
When using either of the Key Vault engine versions, the secret key is the name of a specific secret entry to be retrieved from the group path.
When using the database secrets engine, the secret key itself is arbitrary but is required by Flyte to name and identify the secret file. It is arbitrary because the database secrets engine returns always two keys, `username` and `password` and we need to retrieve a matching pair in one request.

**Configuration**

You can configure the Vault role under which Flyte will try to read the secret by setting webhook.vaultSecretManager.role (default: `"flyte"`).
There is also a deprecated `webhook.vaultSecretManager.kvVersion` setting in the configmap that can be used to specify the version but only for the Key Vault backend engine.
Available choices are: "1", "2". Note that the version number needs to be an explicit string (e.g. `"1"`).

**Annotations**

By default, `flyte-pod-webhook` injects following annotations to task pod:

1. `vault.hashicorp.com/agent-inject` to configure whether injection is explicitly enabled or disabled for a pod.
2. `vault.hashicorp.com/secret-volume-path` to configure where on the filesystem a secret will be rendered.
3. `vault.hashicorp.com/role` to configure the Vault role used by the Vault Agent auto-auth method.
4. `vault.hashicorp.com/agent-pre-populate-only` to configure whether an init container is the only injected container.
5. `vault.hashicorp.com/agent-inject-secret` to configure Vault Agent to retrieve the secrets from Vault required by the container.
6. `vault.hashicorp.com/agent-inject-file` to configure the filename and path in the secrets volume where a Vault secret will be written.
7. `vault.hashicorp.com/agent-inject-template` to configure the template Vault Agent should use for rendering a secret.

It is possible to add extra annotations or override the existing ones in Flyte either at the task level using pod annotations or at the installation level.
If Flyte administrator wants to set up annotations for the entire system, they can utilize `webhook.vaultSecretManager.annotations` to accomplish this.

## Scaling the webhook

### Vertical scaling

To scale the Webhook to be able to process the number/rate of pods you need, you may need to configure a vertical [pod autoscaler](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler).

### Horizontal scaling

The Webhook does not make any external API Requests in response to Pod mutation requests. It should be able to handle traffic quickly. For horizontal scaling, adding additional replicas for the Pod in the
deployment should be sufficient. A single `MutatingWebhookConfiguration` object will be used, the same TLS certificate will be shared across the pods and the Service created will automatically load balance traffic across the available pods.

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/productionizing/
