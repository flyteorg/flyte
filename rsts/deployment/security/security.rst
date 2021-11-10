.. _security-overview:

###################
Security Overview
###################

Here we cover security aspects of running your flyte deployments. In the current state we will be covering the user
which is used for running the flyte services and we will go through why we do this and not run them as root user

#################
Why non-root user
#################
Flyte uses docker for container packaging and by default its containers run as root. This gives full
permissions on the system but may not be suitable for production deployments where a security breach could comprise your
application deployments.
It's considered to be a best practice for security because running in constrained permission environment will prevent any
malicious code from utilizing the full permissions of the host `Ref <https://kubernetes.io/blog/2018/07/18/11-ways-not-to-get-hacked/#8-run-containers-as-a-non-root-user>`__
Also in certain container platforms like `OpenShift <https://engineering.bitnami.com/articles/running-non-root-containers-on-openshift.html>`__ running non-root containers is mandatory.


*******
Changes
*******
New user group and user have been added to the Docker files for all the flyte components
`Flyteadmin <https://github.com/flyteorg/flyteadmin/blob/master/Dockerfile>`__
`Flytepropeller <https://github.com/flyteorg/flytepropeller/blob/master/Dockerfile>`__
`Datacatalog <https://github.com/flyteorg/datacatalog/blob/master/Dockerfile>`__
`Flyteconsole <https://github.com/flyteorg/flyteconsole/blob/master/Dockerfile>`__

And Dockerfile uses `USER command <https://docs.docker.com/engine/reference/builder/#user>`__ which sets user
and group which will be used for running the container.

Additionally the k8s manifest files for the flyte components define the overridden security context with the created
user and group to run them. Following shows the overriden security context added for flyteadmin
`Flyteadmin <https://github.com/flyteorg/flyte/blob/master/charts/flyte/templates/admin/deployment.yaml>`__


************
Why override
************
There are certain init-containers that still require root permissions and hence we required to override the security
context for these.
For example: in case of `Flyteadmin <https://github.com/flyteorg/flyte/blob/master/charts/flyte/templates/admin/deployment.yaml>`__
the init container of check-db-ready that runs postgres-provided docker image is not able to resolve the host for the checks and fails. This is mostly due to no read
permissions on etc/hosts file. Only the check-db-ready container is run using the root user and which we will plan to fix as well.
