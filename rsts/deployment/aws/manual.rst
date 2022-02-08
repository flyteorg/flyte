.. _deployment-aws-manual:

#######################
AWS (EKS) Manual Setup
#######################

************************************************
Flyte Deployment - Manual AWS/EKS Deployment
************************************************
This guide helps you set up Flyte from scratch, on AWS, without using an automated approach. It details step-by-step how to go from a bare AWS account, to a fully functioning Flyte deployment that members of your company can use.

Prerequisites
=============
Before you begin, please ensure that you have the following tools installed.

* AWS CLI
* ``eksctl``
* Access to AWS console
* Helm
* ``kubectl``
* Openssl

AWS Permissioning
=================
Start off by creating a series of roles. These roles control Flyte's access to the AWS account. Since this is a setup guide, we'll be making liberal use of the default policies that AWS IAM comes with, which may be too broad - please consult with your infrastructure team if necessary.

EKS Cluster Role
----------------
First create a role for the EKS cluster. This is the role that the Kubernetes platform itself will use to monitor, scale, and create ASGs, run the etcd store, and the K8s API server, etc.

* Navigate to your AWS console and choose the IAM Roles page.
* Under the EKS service, select EKS-Cluster.
* Ensure that the ``AmazonEKSClusterPolicy`` is selected.
* Create this role without any permission boundary. Advanced users can try to restrict the permissions for their usecases.
* Choose any tags that would help in you tracking this role based on your devops rules.
* Choose a good name for your cluster role which is easier to search eg: <ClusterName-EKS-Cluster-Role>.

Refer the following AWS docs for the details:
https://docs.aws.amazon.com/eks/latest/userguide/service_IAM_role.html#create-service-role

EKS Node IAM Role
-----------------
Next create a role for your compute nodes to use. This is the role that will be given to the nodes that actually run user pods (including Flyte pods).

* Navigate to your AWS console and choose IAM role service.
* Choose EC2 as service while choosing the use case.
* Choose the following policies as mentioned in the linked AWS doc:

  * AmazonEKSWorkerNodePolicy allows EKS nodes to connect to EKS clusters.
  * AmazonEC2ContainerRegistryReadOnly if using Amazon ECR for container registry.
  * AmazonEKS_CNI_Policy which allows pod and node networking within your VPC. You need this even though it is marked
    as optional in the AWS guide linked below.

* Create this role without any permission boundary. Advanced users can try to restrict the permissions for their usecases.
* Choose any tags that would help in you tracking this role based on your devops rules.
* Choose a good name for your node role which is easier to search eg: <ClusterName-EKS-Node-Role>.

Refer the following AWS docs for the details:
https://docs.aws.amazon.com/eks/latest/userguide/create-node-role.html

Flyte System Role
-----------------
Next create a role for the Flyte platform. When pods run, they shouldn't run with the node role created above; they should assume a separate role with permissions suitable for that pod's containers. This role will be used for Flyte's own API servers and associated agents.

Create a role ``iam-role-flyte`` from the IAM console. Select "AWS service" again for the type, and EC2 for the use case.
Attach the ``AmazonS3FullAccess`` policy for now. S3 access can be tweaked later to narrow down the scope.

Flyte User Role
----------------
Finally create a role for Flyte users.  This is the role that user pods will end up assuming when Flyte kicks them off.

Create a role ``flyte-user-role`` from the IAM console. Select "AWS service" again for the type, and EC2 for the use case. Also add the ``AmazonS3FullAccess`` policy for now.

Create an EKS Cluster
=====================
Create an EKS cluster from the AWS console:

* Pick a good name for your cluster eg : <Name-EKS-Cluster>
* Pick Kubernetes version >= 1.19
* Choose the EKS cluster role <ClusterName-EKS-Cluster-Role>, created in previous steps.
* Keep secrets encryption off.
* Use the same VPC where you intend to deploy your RDS instance. Keep the default VPC if none were created and choose RDS to use the default as well.
* Use the subnets for all supported AZ's in that VPC.
* Choose the security group to use for this cluster and the RDS instance (use default if using default VPC).
* Provide public access to your cluster, or otherwise according to your devops settings.
* Choose default version of the network add-ons.
* You can choose to enable the control plane logging to CloudWatch.

Connect to an EKS Cluster
=========================
* Use your AWS account access keys to run the following command to update your kubectl config and switch to the new EKS cluster context:

  .. code-block::

       export AWS_ACCESS_KEY_ID=<YOUR-AWS-ACCOUNT-ACCESS-KEY-ID>
       export AWS_SECRET_ACCESS_KEY=<YOUR-AWS-SECRET-ACCESS-KEY>
       exportAWS_SESSION_TOKEN=<YOUR-AWS-SESSION-TOKEN>

* Switch to the EKS cluster context <Name-EKS-Cluster>

  .. code-block::

     aws eks update-kubeconfig --name <Name-EKS-Cluster> --region <region>

* Verify the context is switched

.. code-block:: bash

   $ kubectl config current-context
   arn:aws:eks:<region>:<AWS_ACCOUNT_ID>:cluster/<Name-EKS-Cluster>

* Test it with ``kubectl``. It should tell you there aren't any resources.

.. code-block::

   $ kubectl get pods
   No resources found in default namespace.

OIDC Provider for the EKS Cluster
=================================
Create the OIDC provider to be used for the EKS cluster and associate a trust relationship with the EKS cluster role <ClusterName-EKS-Cluster-Role>:

* EKS cluster created should have a URL created and hence the following command would return the provider

.. code-block::

  aws eks describe-cluster --region <region> --name <Name-EKS-Cluster> --query "cluster.identity.oidc.issuer" --output text

Example output:

.. code-block::

  https://oidc.eks.<REGION>.amazonaws.com/id/<UUID-OIDC>

* The following command creates the OIDC provider using the address provided by the cluster:

.. code-block::

  eksctl utils associate-iam-oidc-provider --cluster <Name-EKS-Cluster> --approve

Follow this [AWS documentation](https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html) for additional reference.

* Verify the OIDC provider is created by navigating to https://console.aws.amazon.com/iamv2/home?#/identity_providers and confirming that a new provider entry has been created with the same <UUID-OIDC> issuer as the cluster's.

* Next we need to add a trust relationship between this OIDC provider and the two Flyte roles:
   * Navigate to the newly created `OIDC Providers <https://console.aws.amazon.com/iamv2/home?#/identity_providers>`__ with <UUID-OIDC> and copy the ARN.
   * Navigate to `IAM Roles <https://console.aws.amazon.com/iam/home#/roles>`__ and select the ``iam-role-flyte`` role.
   * Under the Trust relationships tab, hit the Edit button.
   * Replace the ``Principal:Federated`` value in the policy JSON below with the copied ARN.
   * Replace the ``<UUID-OIDC>`` placeholder in the ``Condition:StringEquals`` with the last part of the copied ARN. It'll look something like ``8DCF90D22E386AA3975FC4DCD2ECD23BC`` and should match the tail end of the issuer ID from the first step.
     Ensure you don't accidentally remove the ``:aud`` suffix. You need that.
   * Repeat these steps for the ``flyte-user-role``.

.. code-block::

   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Principal": {
           "Service": "eks.amazonaws.com"
         },
         "Action": "sts:AssumeRole"
       },
       {
         "Effect": "Allow",
         "Principal": {
           "Federated": "arn:aws:iam::<AWS_ACCOUNT_ID>:oidc-provider/oidc.eks.<REGION>.amazonaws.com/id/<UUID-OIDC>"
         },
         "Action": "sts:AssumeRoleWithWebIdentity",
         "Condition": {
           "StringEquals": {
             "oidc.eks.<REGION>.amazonaws.com/id/<UUID-OIDC>:aud": "sts.amazonaws.com"
           }
         }
       }
     ]
   }

Create an EKS Node Group
========================

The initial EKS cluster will not have any instances configured to operate the cluster. Create a node group which provides resources for the kubernetes cluster:

* Go to your EKS cluster navigate to the Configuration -> Compute tab.
* Provide a suitable name <Name-EKS-Node-Group>
* Use the EKS node IAM role <ClusterName-EKS-Node-Role> created in the above steps.
* Use without any launch template, kuebernetes labels,taints or tags.
* Choose the default Amazon EC2 AMI (AL2_x86_64).
* Capacity type on demand, Instance type and size can be chosen based on your devops requirements. Keep the default if in doubt.
* Create a node group with 5/10/5 instance min, max, desired.
* Use the default subnets selected which would be chosen based on your EKS cluster accessible subnets.
* Disallow remote access to the nodes (If needed provide the ssh access key pair to use from your account).

Create an RDS Database
======================
Next, create a relational database. This database will be used by both the primary control plane service (Flyte Admin) and the Flyte memoization service (Data Catalog).

* Navigate to `RDS <https://console.aws.amazon.com/rds/home>`__ and create an Aurora engine with Postgres compatibility database.
* Leave the Template as Production.
* Change the default cluster identifier to ``flyteadmin``.
* Set the master username to ``flyteadmin``.
* Choose a master password which you'll later use in your Helm template.

  * `Password <https://github.com/flyteorg/flyte/blob/3600badd2ad49ec2cd1f62752780f201212de3f3/helm/values-eks.yaml#L196>`_

* Leave Public access off.
* Choose the same VPC that your EKS cluster is in.
* In a separate tab, navigate to the EKS cluster page and make note of the security group attached to your cluster.
* Go back to the RDS page and in the security group section, add the EKS cluster's security group (feel free to leave the default as well). This will ensure you don't have to play around with security group rules in order for pods running in the cluster to access the RDS instance.
* Under the top level Additional configuration (there's a sub menu by the same name) under "Initial database name" enter ``flyteadmin`` as well.

Leave all the other settings as is and hit Create.

Check Connectivity to the RDS Database From the EKS Cluster
===========================================================
* Get the <RDS-HOST-NAME> by navigating to the database cluster and copying the writer instance endpoint.

We will use pgsql-postgres-client to verify DB connectivity:

* Create a testdb namespace for trial.

  .. code-block:: bash

     kubectl create ns testdb

* Run the following command with the username and password you used, and the host returned by AWS.

  .. code-block:: bash

     kubectl run pgsql-postgresql-client --rm --tty -i --restart='Never' --namespace testdb --image docker.io/bitnami/postgresql:11.7.0-debian-10-r9 --env="PGPASSWORD=<Password>" --command -- psql testdb --host <RDS-HOST-NAME> -U <Username> -d flyteadmin -p 5432

* If things are working fine then you should drop into a psql command prompt. Type ``\q`` to quit. If you make a mistake in the above command you may need to delete the pod created with ``kubectl -n testdb delete pod pgsql-postgresql-client``

* In case there are connectivity issues then you would see the following error. Please check the security groups on the Database and the EKS cluster.

.. code-block:: bash

   psql: warning: extra command-line argument "testdb" ignored
   psql: could not translate host name "database-2-instance-1.ce40o2y3b4os.us-east-2.rds.amazonaws.co" to address: Name or service not known
   pod "pgsql-postgresql-client" deleted
   pod flyte/pgsql-postgresql-client terminated (Error)

Install an Amazon Loadbalancer Ingress Controller
=================================================

The cluster doesn't come with any ingress controllers so we have to install one separately. This one will create an AWS load balancer for K8s Ingress objects.

Before we begin, make sure all the subnets are tagged correctly for subnet discovery. The controller uses this for creating the ALB's.

* Go to your default VPC subnets. There would be 3 subnets for the 3 AZ's.
* Add 2 tags on all the three subnets
  Key kubernetes.io/role/elb Value 1
  Key kubernetes.io/cluster/<Name-EKS-Cluster> Value shared
* Refer to this doc for additional details https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.1/deploy/subnet_discovery/

* Download the IAM policy for the AWS Load Balancer Controller:

  .. code-block::

     curl -o iam-policy.json https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/v2.2.0/docs/install/iam_policy.json

* Create an IAM policy called AWSLoadBalancerControllerIAMPolicy (delete it if it already exists from IAM service):

  .. code-block::

     aws iam create-policy \
       --policy-name AWSLoadBalancerControllerIAMPolicy \
       --policy-document file://iam-policy.json

* Create an IAM role and ServiceAccount for the AWS Load Balancer controller, using the ARN from the step above:

  .. code-block::

     eksctl create iamserviceaccount \
     --cluster=<cluster-name> \
     --region=<region> \
     --namespace=kube-system \
     --name=aws-load-balancer-controller \
     --attach-policy-arn=arn:aws:iam::<AWS_ACCOUNT_ID>:policy/AWSLoadBalancerControllerIAMPolicy \
     --override-existing-serviceaccounts \
     --approve

* Add the EKS chart repo to helm:

  .. code-block::

     helm repo add eks https://aws.github.io/eks-charts

* Install the TargetGroupBinding CRDs:

  .. code-block::

     kubectl apply -k "github.com/aws/eks-charts/stable/aws-load-balancer-controller//crds?ref=master"

* Install the load balancer controller using helm:

.. code-block::

   helm install aws-load-balancer-controller eks/aws-load-balancer-controller -n kube-system --set clusterName=<Name-EKS-Cluster> --set serviceAccount.create=false --set serviceAccount.name=aws-load-balancer-controller


* Verify load balancer webhook service is running in kube-system ns:

.. code-block::

   kubectl get service -n kube-system

Sample o/p

.. code-block::

   NAME                                TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)         AGE
   aws-load-balancer-webhook-service   ClusterIP   10.100.255.5   <none>        443/TCP         95s
   kube-dns                            ClusterIP   10.100.0.10    <none>        53/UDP,53/TCP   75m

.. code-block::

   $ kubectl get pods -n kube-system
   NAME                                            READY   STATUS    RESTARTS   AGE
   aws-load-balancer-controller-674869f987-brfkj   1/1     Running   0          11s
   aws-load-balancer-controller-674869f987-tpwvn   1/1     Running   0          11s


* Use this doc for any additional installation instructions: 
  https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.2/deploy/installation/


SSL Certificate
===============
In order to use SSL (which we need to use gRPC clients), we now need to create an SSL certificate. We realize that
you may need to work with your infrastructure team to acquire a legitimate certificate, so the first set of instructions
help you get going with a self-signed certificate. These are of course not secure and will show up as a security warning
to any users, so we recommend deploying a legitimate certificate as soon as possible.

Self-Signed Method (Insecure)
-----------------------------

Generate a self signed cert using open ssl and get the <KEY> and <CRT> file.

#. Define req.conf file with the following contents.

  .. code-block::

       [req]
       distinguished_name = req_distinguished_name
       x509_extensions = v3_req
       prompt = no
       [req_distinguished_name]
       C = US
       ST = WA
       L = Seattle
       O = Flyte
       OU = IT
       CN = flyte.example.org
       emailAddress = dummyuser@flyte.org
       [v3_req]
       keyUsage = keyEncipherment, dataEncipherment
       extendedKeyUsage = serverAuth
       subjectAltName = @alt_names
       [alt_names]
       DNS.1 = flyte.example.org

#. Use openssl to generate the KEY and CRT files.

   .. code-block::

      openssl req -x509 -nodes -days 3649 -newkey rsa:2048 -keyout key.out -out crt.out -config req.conf -extensions 'v3_req'

#. Create ARN for the cert.

   .. code-block::

      aws acm import-certificate --certificate fileb://crt.out --private-key fileb://key.out --region <REGION>

Production
----------

Generate a cert from the CA used by your org and get the <KEY> and <CRT>.
Flyte doesn't manage the lifecycle of certificates so this will need to be managed by your security or infrastructure team.

Please refer to:
AWS docs for importing the cert https://docs.aws.amazon.com/acm/latest/userguide/import-certificate-prerequisites.html
Requesting a public cert issued by ACM Private CA https://docs.aws.amazon.com/acm/latest/userguide/gs-acm-request-public.html#request-public-console

Note the generated ARN. Let's calls it <CERT-ARN> in this doc which we will use to replace in our values-eks.yaml

Use the AWS Certificate manager for generating the SSL certificate to host your hosted Flyte installation.


Create an S3 Bucket
===================
* Create an S3 bucket without public access.
* Choose a good name for it  <ClusterName-Bucket>
* Use the same region as the EKS cluster.


Create a Log Group
==================
Navigate to the `AWS Cloudwatch <https://console.aws.amazon.com/cloudwatch/home>`__ page and create a Log Group.
Give it a reasonable name like ``flyteplatform``.

Time for Helm
=============

Installing Flyte
-----------------

1. Add the Flyte chart repo to Helm

.. code-block::

 helm repo add flyteorg https://flyteorg.github.io/flyte


2. Download EKS values for Helm

.. tabbed:: Flyte Native Scheduler

  * Download EKS Helm values (it enables Flyte native scheduler by default)

    .. code-block:: bash

       curl -sL https://raw.githubusercontent.com/flyteorg/flyte/master/charts/flyte-core/values-eks.yaml

.. tabbed:: AWS scheduler

  * Download EKS helm values for AWS Scheduler

    .. code-block:: bash

       curl -sL https://raw.githubusercontent.com/flyteorg/flyte/master/charts/flyte-core/values-eks.yaml
       curl -sL https://raw.githubusercontent.com/flyteorg/flyte/master/charts/flyte-core/values-eks-override.yaml


3. Update values in the YAML file

Search and replace the following:

.. list-table:: Helm EKS Values
   :widths: 25 25 75
   :header-rows: 1

   * - Placeholder
     - Description
     - Sample Value
   * - ``<ACCOUNT_NUMBER>``
     - The AWS Account ID
     - ``173113148371``
   * - ``<AWS_REGION>``
     - The region your EKS cluster is in
     - ``us-east-2``
   * - ``<RDS_HOST_DNS>``
     - DNS entry for your Aurora instance
     - ``flyteadmin.cluster-cuvm8rpzqloo.us-east-2.rds.amazonaws.com``
   * - ``<BUCKET_NAME>``
     - Bucket used by Flyte
     - ``my-sample-s3-bucket``
   * - ``<DB_PASSWORD>``
     - The password in plaintext for your RDS instance
     - awesomesauce
   * - ``<LOG_GROUP_NAME>``
     - CloudWatch Log Group
     - ``flyteplatform``
   * - ``<CERTIFICATE_ARN>``
     - ARN of the self-signed (or official) certificate
     - ``arn:aws:acm:us-east-2:173113148371:certificate/763d12d5-490d-4e1e-a4cc-4b28d143c2b4``


4. (Optional) Configure Flyte project and domain

To restrict projects, update Helm values. By default, Flyte creates three projects: Flytesnacks, Flytetester, and Flyteexample.

.. code-block::

   # you can define projects as per your need
   flyteadmin:
    initialProjects:
       - flytesnacks
       - flytetester
       - flyteexamples

To restrict domains, update the Helm values again. By default, Flyte creates three domains per project: development, staging and production.

.. code-block::

   # -- Domain configuration for Flyte project. This enables the specified number of domains across all projects in Flyte.
   configmap
     domain:
       domains:
         - id: development
           name: development
         - id: staging
           name: staging
         - id: production
           name: production

   # Update Cluster resource manager only if you are using Flyte resource manager. It will create the required resource in the project-domain namespace.
   cluster_resource_manager:
     enabled: true
     config:
       cluster_resources:
          customData:
            - development:
                - projectQuotaCpu:
                  value: "5"
                - projectQuotaMemory:
                  value: "4000Mi"
                - defaultIamRole:
                  value: "arn:aws:iam::{{ .Values.userSettings.accountNumber }}:role/flyte-user-role"
            - staging:
                - projectQuotaCpu:
                  value: "2"
                - projectQuotaMemory:
                  value: "3000Mi"
                - defaultIamRole:
                  value: "arn:aws:iam::{{ .Values.userSettings.accountNumber }}:role/flyte-user-role"
            - production:
                - projectQuotaCpu:
                  value: "2"
                - projectQuotaMemory:
                  value: "3000Mi"
                - defaultIamRole:
                  value: "arn:aws:iam::{{ .Values.userSettings.accountNumber }}:role/flyte-user-role"

5. Install Flyte

.. tabbed:: Flyte Native scheduler

  * Install Flyte with Flyte native scheduler

    .. code-block:: bash

       helm install -n flyte -f values-eks.yaml --create-namespace flyte flyteorg/flyte-core

.. tabbed:: AWS scheduler

  * Install Flyte with Flyte AWS Scheduler

    .. code-block:: bash

       helm install -n flyte -f values-eks.yaml -f values-eks-override.yaml --create-namespace flyte flyteorg/flyte-core


6. Verify if all of the pods have come up correctly

.. code-block:: bash

   kubectl get pods -n flyte

Uninstalling Flyte
------------------

.. code-block:: bash

   helm uninstall -n flyte flyte

Upgrading Flyte
---------------

.. tabbed:: Flyte Native Scheduler

  * Install Flyte with flyte native scheduler:

    .. code-block:: bash

       helm upgrade -n flyte -f values-eks.yaml --create-namespace flyte flyteorg/flyte-core

.. tabbed:: AWS Scheduler

  * Install Flyte with flyte aws scheduler

    .. code-block:: bash

       helm upgrade -n flyte -f values-eks.yaml -f values-eks-override.yaml --create-namespace flyte flyteorg/flyte-core

Connecting to Flyte
===================

Flyte can be accessed using the UI console or your terminal.

* First, find the Flyte endpoint created by the ALB ingress controller.

.. code-block:: bash

   $ kubectl -n flyte get ingress

   NAME         CLASS    HOSTS   ADDRESS                                                       PORTS   AGE
   flyte        <none>   *       k8s-flyte-8699360f2e-1590325550.us-east-2.elb.amazonaws.com   80      3m50s
   flyte-grpc   <none>   *       k8s-flyte-8699360f2e-1590325550.us-east-2.elb.amazonaws.com   80      3m49s

<FLYTE-ENDPOINT> = Value in ADDRESS column and both will be the same as the same port is used for both GRPC and HTTP. Own host website may work too.


* Connect to flytectl CLI.

Add :<FLYTE-ENDPOINT>  to ~/.flyte/config.yaml eg ;

.. code-block::

    admin:
     # For GRPC endpoints you might want to use dns:///flyte.myexample.com
     endpoint: dns:///<FLYTE-ENDPOINT>
     insecureSkipVerify: true # only required if using a self-signed cert. Caution: not to be used in production
     insecure: true # only required when using insecure ingress. Secure ingress may cause an unavailable desc error to true option.
    logger:
     show-source: true
     level: 0
    storage:
      kind: s3
      config:
        auth_type: iam
        region: <REGION> # Example: us-east-2
      container: <ClusterName-Bucket> # Example my-bucket. Flyte k8s cluster / service account for execution should have access to this bucket

Accessing Flyte Console (Web UI)
================================

* Use the https://<FLYTE-ENDPOINT>/console to get access to flyteconsole UI
* Ignore the certificate error if using a self signed cert

Troubleshooting
===============


* If a flyteadmin pod is not coming up, then describe the pod and check which of the container or init-containers had an error.

.. code-block:: bash

   kubectl describe pod/<flyteadmin-pod-instance> -n flyte

Then check the logs for the container which failed.

Eg: to check for run-migrations init container do this:

.. code-block:: bash

   kubectl logs -f <flyteadmin-pod-instance> run-migrations -n flyte

If you see connectivity issues, then check your security group rules on the DB and eks cluster. 

For authentication issues, check that you have used the same password in helm and RDS DB creation. 

(Note : When using Cloud formation templates, make sure the passwords are not double/single quoted.)

* Increasing log level for flytectl
  Change your logger config to this
  .. code-block::

     logger:
     show-source: true
     level: 6
