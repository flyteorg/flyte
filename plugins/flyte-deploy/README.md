# flyte-deploy

A [Claude Code](https://docs.claude.com/en/docs/claude-code) skill that walks an agent
through deploying a **Flyte v2 (`flyte-binary`) cluster on AWS from scratch**: EKS + S3 +
RDS PostgreSQL + AWS Load Balancer Controller + `helm`, with optional TLS (ACM, incl.
cross-account DNS) and Okta/OIDC SSO (console edge SSO + CLI Bearer-bypass).

It encodes the full runbook plus the non-obvious gotchas that break a first deploy
(eksctl version vs EKS, the RDS source security group, ALB controller IAM drift, the
`runs` vs `runs.server` config nesting, and the task-pod control-plane callback env vars).

## Install (Claude Code plugin marketplace)

```
/plugin marketplace add flyteorg/flyte
/plugin install flyte-deploy@flyte-skills
```

Then ask Claude to "deploy a Flyte v2 cluster on AWS", or invoke it directly with
`/flyte-deploy:flyte-deploy`.

## Install (manual)

Copy `skills/flyte-deploy/` into your `~/.claude/skills/` directory.

## Scope & safety

The skill provisions real, billable AWS resources and runs `aws`/`eksctl`/`helm`
commands. Every value in `<angle brackets>` and the example hostnames/IDs are
placeholders — replace them with your own. Review each step before running it.
