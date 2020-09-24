[All Overlays](./)
# :construction: Amazon Webservices Elastic Kubernetes Service Overlay

This overlay serves as an example to bootstrap Flyte setup on AWS. It is not
designed to work out of the box due to the need of AWS resources. Please follow the instruction
below to further configure.

_Hint_: searching `TODO:` through this directory would help to understand what needs to be done.

## AWS RDS


## FlyteAdmin


## FlyteConsole


## DataCatalog


## Build it
Refer to previous documentation

## Now ship it

``` shell
make
kubectl apply -f deployment/gcp/flyte_generated.yaml
```
