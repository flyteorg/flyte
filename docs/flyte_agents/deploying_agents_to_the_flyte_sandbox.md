---
jupytext:
  formats: md:myst
  text_representation:
    extension: .md
    format_name: myst
---

(deploying_agents_to_the_flyte_sandbox)=
# Deploying agents to the Flyte Sandbox

After you have finished {ref}`testing an agent locally <testing_agents_locally>`, you can deploy your agent to the flyte sandbox.

Here's a step by step guide to deploy your agent to the flyte sandbox.

1. Start the flyte sandbox.
```bash
flytectl demo start
```

2. Build an agent image.
You can go to [here](https://github.com/flyteorg/flytekit/blob/master/Dockerfile.agent) to see the dockerfile we use in flytekit python.
Take Databricks agent as an example.
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

2. Deploy your agent image to the kubernetes cluster
```bash
kubectl set image deployment/flyteagent flyteagent=localhost:30000/flyteagent:example
```

3. Set up your secrets
Let's take databricks agent as the example.
```bash
kubectl edit secret flyteagent -n flyte
```
Get your `BASE64_ENCODED_DATABRICKS_TOKEN`.
```bash
echo -n "<DATABRICKS_TOKEN>" | base64
```
Add your token to the `data` field.
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
Please ensure 2 things.
1. The secret name should not include uppercase English letters.
2. The secret value is encoded in Base64.
:::

4. Restart development
```bash
kubectl rollout restart deployment flyte-sandbox -n flyte
```

5. Test your agent remotely in flyte sandbox
```bash
pyflyte run --remote agent_workflow.py agent_task
```

:::{note}
Please ensure you've built an image and specified it by `--image` flag or use ImageSpec to include the plugin for the task.
:::
