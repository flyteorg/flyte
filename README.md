# Flyte

This is a work in progress. Please ignore this for now.


## Registration

`cat > ~/.flyte/config-k3d.yaml`
```
admin:
  # For GRPC endpoints you might want to use dns:///flyte.myexample.com
  endpoint: localhost:30081
  insecure: true
logger:
  show-source: true
  level: 0
```

```
flytectl -c ~/.flyte/config-k3d.yaml -p flytesnacks -d dev register files -a ./cookbook/flyte-package.tgz
```

