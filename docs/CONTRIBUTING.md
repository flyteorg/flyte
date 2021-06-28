# Developing FlyteCtl

A local cluster can be setup via --> https://docs.flyte.org/en/latest/getting_started.html

Then, if having trouble connecting to local cluster see the following:

#1) Find/Set/Verify gRPC port for your local Flyte service:

FLYTECTL_GRPC_PORT=`kubectl get service -n flyte flyteadmin -o json | jq '.spec.ports[] | select(.name=="grpc").port'`

#2) Setup Port forwarding: kubectl port-forward -n flyte service/flyteadmin 8081:$FLYTECTL_GRPC_PORT

#3) Update config line in https://github.com/flyteorg/flytectl/blob/master/config.yaml to dns:///localhost:8081

#4) All new flags introduced for flytectl commands and subcommands should be camelcased. eg: bin/flytectl update project -p flytesnacks --activateProject

# DCO: Sign your work

Flyte ships commit hooks that allow you to auto-generate the DCO signoff line if
it doesn't exist when you run `git commit`. Simply navigate to the flytectl project root and run

```bash
make dco
```

