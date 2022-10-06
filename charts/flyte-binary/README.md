

Command to generate yaml files.

helm template lev0 ./ -n flyte --include-crds --output-dir gen --dependency-update --debug --create-namespace -a rbac.authorization.k8s.io/v1 -a networking.k8s.io/v1/Ingress -a apiextensions.k8s.io/v1/CustomResourceDefinition

The crd is in a different folder and is expected to be created separately. This reduces the permissions the fltye binary needs.
