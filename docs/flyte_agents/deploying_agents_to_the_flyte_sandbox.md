---
jupytext:
  formats: md:myst
  text_representation:
    extension: .md
    format_name: myst
---

(deploying_agents_to_the_flyte_sandbox)=
# Deploying agents to the Flyte sandbox

After you have finished {ref}`testing an agent locally <testing_agents_locally>`, you can deploy your agent to the Flyte sandbox.

Here's a step by step guide to deploying your agent image to the Flyte sandbox.

1. Start the Flyte sandbox:
```bash
flytectl demo start
```

2. Build an agent image:
You can go to [here](https://github.com/flyteorg/flytekit/blob/master/Dockerfile.agent) to see the Dockerfile we use in flytekit python.
Take Databricks agent as an example:
```Dockerfile
FROM python:3.9-slim-bookworm

RUN apt-get update && apt-get install build-essential git -y
RUN pip install prometheus-client grpcio-health-checking
RUN pip install --no-cache-dir -U flytekit \
    git+https://github.com/flyteorg/flytekit.git@<gitsha>#subdirectory=plugins/flytekit-spark \
    && apt-get clean autoclean \
    && apt-get autoremove --yes \
    && rm -rf /var/lib/{apt,dpkg,cache,log}/ \
    && :

CMD pyflyte serve agent --port 8000
```
```bash
docker buildx build -t localhost:30000/flyteagent:example -f Dockerfile.agent . --load
docker push localhost:30000/flyteagent:example
```

2. Deploy your agent image to the Kubernetes cluster:
```bash
kubectl edit deployment flyteagent -n flyte
```
Search for the `image` key and change its value to your agent image:
```yaml
image: localhost:30000/flyteagent:example
```

3. Set up your secrets:
Let's take Databricks agent as an example:
```bash
kubectl edit secret flyteagent -n flyte
```
Get your `BASE64_ENCODED_DATABRICKS_TOKEN`:
```bash
echo -n "<DATABRICKS_TOKEN>" | base64
```
Add your token to the `data` field:
```yaml
apiVersion: v1
data:
  flyte_databricks_access_token: <BASE64_ENCODED_DATABRICKS_TOKEN>
kind: Secret
metadata:
  annotations:
    meta.helm.sh/release-name: flyteagent
    meta.helm.sh/release-namespace: flyte
  creationTimestamp: "2023-10-04T04:09:03Z"
  labels:
    app.kubernetes.io/managed-by: Helm
  name: flyteagent
  namespace: flyte
  resourceVersion: "753"
  uid: 5ac1e1b6-2a4c-4e26-9001-d4ba72c39e54
type: Opaque
```
:::{note}
Please ensure two things:
1. The secret name consists only of lowercase English letters.
2. The secret value is encoded in Base64.
:::

4. Restart development:
```bash
kubectl rollout restart deployment flyte-sandbox -n flyte
```

5. Test your agent remotely in the Flyte sandbox:
```bash
pyflyte run --remote agent_workflow.py agent_task
```

:::{note}
You must build an image that includes the plugin for the task and specify its config with the [`--image` flag](https://docs.flyte.org/en/latest/api/flytekit/pyflyte.html#cmdoption-pyflyte-run-i) when running `pyflyte run` or in an {ref}`ImageSpec <imagespec>` definition in your workflow file.
:::