###################################
# WORK IN PROGRESS still
###################################

SQL Database
------------
Create a SQL database (Postgres)
https://cloud.google.com/sql/docs/postgres/create-instance

Enable the the SQL server to be accessed from the GKE cluster that will host the FlyteAdmin service. This can be done using private networking mode and associating the shared network.

Create a database called "flyte" in this DB instance

Configuring Flyte to access DB
------------------------------

In this sample we pass the username and password directly in the config file.
TODO: Example of how to use kube secrets to pass the username and password.

Auth / IAM
----------

On GKE you can follow instructions listed here
https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity
to setup WorkloadIdentity and serviceAccounts.

Important commands
kubectl create serviceaccount --namespace flytekit-development flyte-sandbox
gcloud iam service-accounts add-iam-policy-binding --role roles/iam.workloadIdentityUser --member "serviceAccount:flyte-sandbox.svc.id.goog[flytekit-development/flyte-sandbox]" flyte-sandbox@flyte-sandbox.iam.gserviceaccount.com
kubectl annotate serviceaccount  --namespace flytekit-development flyte-sandbox iam.gke.io/gcp-service-account=flyte-sandbox@flyte-sandbox.iam.gserviceaccount.com


IAM For Flyte components
------------------------
Create the right service accounts in GKE cluster's flyte namespace and then add the serviceaccountname to propeller and flyteadmin deployments. You may also want to add it to the various plugin
deployments.

gcloud iam service-accounts add-iam-policy-binding --role roles/iam.workloadIdentityUser --member "serviceAccount:flyte-sandbox.svc.id.goog[flyte/flyteadmin]" flyte-sandbox@flyte-sandbox.iam.gserviceaccount.com
kubectl annotate serviceaccount  --namespace flyte flyteadmin iam.gke.io/gcp-service-account=flyte-sandbox@flyte-sandbox.iam.gserviceaccount.com
gcloud iam service-accounts add-iam-policy-binding --role roles/iam.workloadIdentityUser --member "serviceAccount:flyte-sandbox.svc.id.goog[flyte/flytepropeller]" flyte-sandbox@flyte-sandbox.iam.gserviceaccount.com
kubectl annotate serviceaccount  --namespace flyte flytepropeller iam.gke.io/gcp-service-account=flyte-sandbox@flyte-sandbox.iam.gserviceaccount.com

IAM for workflows
-----------------
As a platform admin, you will need to associate service accounts with the target namespaces (project-domain) combination.  Flyte allows launching workflows with serviceAccounts. Thus when the end user
requests a workflow launch or declares a workflow the right account should be associated within the right namespace. 

TODO: Future plans to automate this creation and association
