Install flyte

The simplest Flyte deployment is the “sandbox” deployment, which includes everything you need in order to use Flyte. The Flyte sandbox can be deployed with a single command

`kubectl create -f https://raw.githubusercontent.com/lyft/flyte/master/deployment/sandbox/flyte_generated.yaml`{{execute HOST1}}

This deployment uses a kubernetes NodePort for Flyte ingress. Once deployed, you can access the Flyte console on any kubernetes node at http://{{ any kubernetes node }}:30081/console (note that it will take a moment to deploy).

Wait until all pods are not in running condition

Verify flyteadmin deployment status
`kubectl -n flyte rollout status deployment flyteadmin`{{execute HOST1}}

Verify flytepropeller deployment status
`kubectl -n flyte rollout status deployment flytepropeller`{{execute HOST1}}

Verify minio deployment status
`kubectl -n flyte rollout status deployment minio`{{execute HOST1}}

Verify minio deployment status
`kubectl -n flyte rollout status deployment contour`{{execute HOST1}}

We are ready for the Demo

After All pods are in running condition then visit flyte console https://[[HOST_SUBDOMAIN]]-30081-[[KATACODA_HOST]].environments.katacoda.com/console
