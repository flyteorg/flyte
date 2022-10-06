

Command to generate yaml files.

To generate one file: `make helm`
To generate multiple files:

helm template lev0 ./ -n flyte --output-dir gen --dependency-update --debug --create-namespace -a rbac.authorization.k8s.io/v1 -a networking.k8s.io/v1/Ingress -a apiextensions.k8s.io/v1/CustomResourceDefinition


Debug install command
helm install tstinstall ./ -n flyte --debug --dry-run --dependency-update --create-namespace


