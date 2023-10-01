# Google Cloud Platform Overlay

This overlay serves as an example to bootstrap Flyte setup on Google Cloud Platform (GCP). It is not
designed to work out of the box due to the need of GCP resources. Please follow the instruction
below to further configure.

_Hint_: searching `TODO:` through this directory would help to understand what needs to be done.

## Cloud SQL

[Cloud SQL](https://cloud.google.com/sql) is used as persistence layer. To set it up, please
follow standard GCP documentation.

A few things are required for this overlay to function:

* Two databases named as `flyte` and `datacatalog`
* A database user named as `flyte`
* Password of the database user should be uploaded to GKE as a k8s secret named as `db-user-pass`
  containing of a file named as `db_pwd.txt`of which the content is the plain text password
* Service account(s) associated with `flyteadmin` and `datacatalog` pods (either as GKE cluster
  service account or through workload identity) should have `Cloud SQL Editor` role

To securely access Cloud SQL instance, [Cloud SQL
Proxy](https://cloud.google.com/sql/docs/postgres/connect-admin-proxy) is launched as a pod sitting
in between Flyte and Cloud SQL instance.

The kustomization files can be found under [cloudsqlproxy](cloudsqlproxy). Please note that one
needs to replace `<project-id>` and `<region>` accordingly in
[cloudsqlproxy/deployment.yaml](cloudsqlproxy/deployment.yaml).

## flyteadmin

flyteadmin configuration is kept as similar as [sandbox](../sandbox) overlay, with only necessary
modifications such as database, storage and CORS.

If one has followed [Cloud SQL](#cloud-sql) section, there is nothing to be done for database.

For storage layer, a few things needs to be done:

* Create a GCS bucket named as `flyte` in a GCP project
* Replace `<project-id>` in [admin/flyteadmin_config.yaml](admin/flyteadmin_config.yaml) with the
  GCP project ID

For CORS to work properly, one needs to use real origin in
[admin/flyteadmin_config.yaml](admin/flyteadmin_config.yaml) `server -> security -> allowedOrigins`.

flyteadmin (including metrics endpoint) is exposed as a service using [internal load
balancer](https://cloud.google.com/kubernetes-engine/docs/how-to/internal-load-balancing).

## flyteconsole

[flyteconsole configmap](console/configmap.yaml) needs to be updated with flyteadmin internal load
balancer IP address or the DNS name associated with it if any.

flyteconsole is exposed as a service using [internal load
balancer](https://cloud.google.com/kubernetes-engine/docs/how-to/internal-load-balancing).

## flytepropeller

flytepropeller configuration is kept as similar as [sandbox](../sandbox) overlay, with only
necessary modifications such as storage.

For storage layer, a few things needs to be done:

* Create a GCS bucket named as `flyte` in a GCP project (skip this if already done)
* Replace `<project-id>` in [propeller/config.yaml](propeller/config.yaml) with the
  GCP project ID
* Replace `<project-id>` in [propeller/plugins/config.yaml](propeller/plugins/config.yaml) with the
  GCP project ID

By default, three plugins are enabled:

* container
* k8s-array
* sidecar

flytepropeller metrics endpoint is exposed as a service using [internal load
balancer](https://cloud.google.com/kubernetes-engine/docs/how-to/internal-load-balancing).

## datacatalog

datacatalog configuration is kept as similar as [sandbox](../sandbox) overlay, with only
necessary modifications such as database and storage.

If one has followed [Cloud SQL](#cloud-sql) section, there is nothing to be done for database.

For storage layer, a few things needs to be done:

* Create a GCS bucket named as `flyte` in a GCP project (skip this if already done)
* Replace `<project-id>` in [datacatalog/datacatalog_config.yaml](propeller/config.yaml) with the
  GCP project ID

datacatalog metrics endpoint is exposed as a service using [internal load
balancer](https://cloud.google.com/kubernetes-engine/docs/how-to/internal-load-balancing).


## Now ship it

``` shell
make
kubectl apply -f deployment/gcp/flyte_generated.yaml
```
