.. _security-overview:

###################
Security Overview
###################

Here we cover the security aspects of running your flyte deployments. In the current state, we will cover the user
used for running the flyte services, and go through why we do this and not run them as a root user.

*****************
Why non-root user
*****************
Flyte uses docker for container packaging, and by default, its containers run as root. This gives full
permissions to the system but may not be suitable for production deployments where a security breach could comprise your
application deployments.
It's considered to be a best practice for security because running in a constrained permission environment will prevent any
malicious code from utilizing the full permissions of the host `Ref <https://kubernetes.io/blog/2018/07/18/11-ways-not-to-get-hacked/#8-run-containers-as-a-non-root-user>`__
Also, in certain container platforms like `OpenShift <https://engineering.bitnami.com/articles/running-non-root-containers-on-openshift.html>`__, running non-root containers is mandatory.


*******
Changes
*******
A new user group and user have been added to the Docker files for all the flyte components
`Flyteadmin <https://github.com/flyteorg/flyteadmin/blob/master/Dockerfile>`__,
`Flytepropeller <https://github.com/flyteorg/flytepropeller/blob/master/Dockerfile>`__,
`Datacatalog <https://github.com/flyteorg/datacatalog/blob/master/Dockerfile>`__,
`Flyteconsole <https://github.com/flyteorg/flyteconsole/blob/master/Dockerfile>`__.

And Dockerfile uses the `USER command <https://docs.docker.com/engine/reference/builder/#user>`__, which sets the user
and group, that's used for running the container.

Additionally, the k8s manifest files for the flyte components define the overridden security context with the created
user and group to run them. The following shows the overridden security context added for flyteadmin
`Flyteadmin <https://github.com/flyteorg/flyte/blob/master/charts/flyte/templates/admin/deployment.yaml>`__.


************
Why override
************
Certain init-containers still require root permissions, and hence we are required to override the security
context for these.
For example: in the case of `Flyteadmin <https://github.com/flyteorg/flyte/blob/master/charts/flyte/templates/admin/deployment.yaml>`__,
the init container of check-db-ready that runs postgres-provided docker image cannot resolve the host for the checks and fails. This is mostly due to no read
permissions on etc/hosts file. Only the check-db-ready container is run using the root user, which we will also plan to fix.


************
OAuth
************
Flytectl requires CA-certified SSL cert for OAuth to work. Using a self-signed certificate throws the following error:

.. code-block::
    
    certificate is not standards compliant.

There are two options to fix this:

#. Switch to a full external OAuth Provider (okta, GCP cloud identity, keycloak, Azure AD, etc.).
#. Use a CA-certified SSL cert.

********************************************************
Running flyteadmin and flyteconsole on different domains
********************************************************
In some cases when flyteadmin and flyteconsole are running on different domains then you would need to allow the flyteadmin's domain to allow cross origin request from the flyteconsole's domain
This can be done by changing the flyteadmin config in the following manner

#. Assuming <flyte-admin-domain> is the domain which will be get request
#. And <flyte-console-domain> is the domain which will be sending the request as the originator

   .. code-block::

             security:
               allowCors: true
               ......
               allowedOrigins:
               - 'https://<flyte-console-domain>'
               ......

#. And then rollout admin

   .. prompt:: bash

      kubectl rollout restart deployment/flyteadmin -n <flyte-admin-domain>
