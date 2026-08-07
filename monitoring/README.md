# Monitoring assets

Grafana dashboards for Flyte v2.

```
dashboards/flyte-execution.json   RPC latency, throughput and error rates per service
```

## What it charts

The `rpc_*` metric family that flyte2 emits over OTLP (`rpc_server_duration_milliseconds`,
`rpc_client_duration_milliseconds`, request/response sizes), broken down by
`service_name`, `rpc_method` and `rpc_service`. It expects a Prometheus
datasource with **uid `prometheus`** — the default that kube-prometheus-stack
creates. Point it elsewhere by editing the datasource uid, or by importing
through the Grafana UI and picking a datasource.

Metrics only reach Prometheus if flyte2 exports them, so the deployment needs an
`otel` config section pointing at a collector, and the collector's Prometheus
exporter needs to be scraped.

## Using it

**Devbox** — bundled, nothing to do:

```bash
make devbox-run
make devbox-monitoring     # http://localhost:30300/d/oss/flyte-execution
```

**A cluster running kube-prometheus-stack** (typical AWS/GCP install) — the
Grafana sidecar provisions any ConfigMap carrying the `grafana_dashboard: "1"`
label, from any namespace:

```bash
kubectl create configmap flyte-execution-dashboard \
  --from-file=monitoring/dashboards/ -n default \
  --dry-run=client -o yaml | \
  kubectl label -f - --local grafana_dashboard=1 --dry-run=client -o yaml | \
  kubectl apply -f -
```

**Any other Grafana** — import `dashboards/flyte-execution.json` through the UI
(Dashboards → New → Import), or mount it via file provisioning.

## Editing

Edit in Grafana, then export via **Share → Export → Save to file** and replace
the JSON here, so changes are reviewable as a JSON diff. Keep the `uid` stable
(`oss`) — links to `/d/oss/flyte-execution` depend on it.
