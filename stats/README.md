## Developing stats

These scripts generate dashboards by running the following ``make`` target in the
flyteorg/flyte root folder.

```bash
make stats
```

Ensure that you have installed the requirements.txt

```bash
pip install -r requirements.txt
```

Refer to [Grafanalib](https://github.com/weaveworks/grafanalib) to understand
how to write the dashboards.

Currently the dashboards are manually uploaded to [Grafana marketplace](https://grafana.com/grafana/dashboards?search=flyte)
