# Flyte

This is a work in progress. This was branched off of this flyte repo initially, with most everything removed save the helm template.

Changes from the helm template on master are:

* Spark/K8s dashboard disabled.
* Removed the Admin cluster resource manager.
* Slimmed down the role that Admin runs with since most of the permissions were for the cluster resource manager
* Changed the Admin role to a role from a cluster role. Because of this there will need to be roles/bindings for each namespace that Admin will need to operate in, I've added one.

I've kept the following for now. These can be removed later when deployed on prem when there are suitable replacements.
* Contour/Envoy
* Minio
* Postgres

## Iterating
To rebuild the Helm chart run `make helm` - this is just for us to look at, Helm install will recompile it.

## Local testing

To play around with this version in sandbox mode, follow the instructions on the K3d tab on the [Sandbox deployment guide](https://docs.flyte.org/en/latest/deployment/sandbox.html#deploy-flyte-sandbox-environment-laptop-workstation-single-machine). Basically:

1. Make sure that other Flyte sandboxes are shut down since the ports will interfere, and `unset KUBECONFIG` if you have it set since it interferes with `k3d`.
1. Start the cluster
  ```k3d cluster create -p "30081:30081" -p "30084:30084" --no-lb --k3s-server-arg '--no-deploy=traefik' --k3s-server-arg '--no-deploy=servicelb' flyte
  ```
1. Enable it with `kubectl config use-context k3d-flyte`
1. Create the Flyte namespace.  `kubectl create ns flyte`. Newer versions of Helm no longer create one for you.
1. Create a `flytectl` config
    ```
    cat > ~/.flyte/config-k3d.yaml

    admin:
      # For GRPC endpoints you might want to use dns:///flyte.myexample.com
      endpoint: localhost:30081
      insecure: true
    logger:
      show-source: true
      level: 0
    ```
1. Install Flyte: `helm install -f helm/values-jpm.yaml -n flyte jpm ./helm`
1. Register some toy examples:
```
flytectl -c ~/.flyte/config-k3d.yaml -p flytesnacks -d dev register files -a https://github.com/flyteorg/flytesnacks/releases/download/v0.2.152/flytesnacks-core.tgz
```
