---
jupytext:
  formats: md:myst
  text_representation:
    extension: .md
    format_name: myst
---

(how_secret_works_in_agent)=
# How Secret Works in Agent

In Flyte agent's deployment, we mount secrets in Kubernetes with the namespace `flyte` and the name `flyteagent`.
If you want to add secrets for agents, you can use the following command:

```bash
SECRET_VALUE=$(<YOUR_SECRET_VALUE> | base64) && kubectl patch secret flyteagent -n flyte --patch "{\"data\":{\"your_agent_secret_name\":\"$SECRET_VALUE\"}}"
```
