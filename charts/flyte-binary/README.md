

Command to generate yaml files.

To generate one file: `make helm`
To generate multiple files:

helm template lev0 ./ -n flyte --output-dir gen --dependency-update --debug --create-namespace -a rbac.authorization.k8s.io/v1 -a networking.k8s.io/v1/Ingress -a apiextensions.k8s.io/v1/CustomResourceDefinition


Debug install command
helm install tstinstall ./ -n flyte --debug --dry-run --dependency-update --create-namespace


For flyte-binary. Reason this is here instead of makefile is because the kubectl command seems to hit the server for the dryrun, if not signed in it can fail.
```
helm template flytedemo ./ -f flytectldemo.yaml -n flyte --dependency-update --debug --create-namespace -a rbac.authorization.k8s.io/v1 -a networking.k8s.io/v1/Ingress -a apiextensions.k8s.io/v1/CustomResourceDefinition | kubectl create -f - --namespace=flyte --dry-run=client -o yaml > gen/tmpoutput.yaml
```
