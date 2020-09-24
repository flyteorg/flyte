[Back to main menu](../)
# Base Components for Flyte
These deployments provide individual deployment units of the Flyte Backend.

As a user it might be preferable to use the `single_cluster` deployment base to create an overlay on top of, or directly edit on top of one of the existing overlays.

## To create a new flyte overlay for one K8s cluster
 Start here 
- [Single Cluster Flyte Deployment configuration](./single_cluster)

## To create a completely custom overlay refer to components
1. FlyteAdmin [Deployment](./admindeployment) | [ServiceAccount](./adminserviceaccount)
1. [Core Flyte namespace creation](./namespace)
1. [FlytePropeller](./propeller) & its [CRD](./wf_crd)
1. [DataCatalog](./datacatalog)
1. [FlyteConsole](./console)
1. [Overall Ingress for Flyte (optional)](./ingress)
1. [Additional plugin components for Flyte using K8s operators](./operators)

