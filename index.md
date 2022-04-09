# Flyteorg Helm Registry

Add this repository using:

```bash
$ helm repo add flyteorg https://helm.flyte.org
$ helm repo update
```

## Flyte Helm Chart

The Flyte Helm Chart helps you to run flyte cluster

### Prerequisite

- You need production grade Postgres Aurora or equivalent database
- You need s3/GCS storage
- You need a production grade kubernetes cluster
- Please make sure you have connectivity with database from k8s cluster
- Please make sure you have correct role & service account in place

#### Production Clusters for eks

NOTE: Before installing please read our [manual guide](https://docs.flyte.org/en/latest/deployment/aws/manual.html#aws-permissioning) for aws and create roles & service account. They are required for communication with s3.

You need Prerequisite in place before installing flyte cluster in eks

| Placeholder | Description | Sample Value |
| -------- | -------- | -------- |
| <ACCOUNT_NUMBER>    | The AWS Account ID | 173113148371 |
| <AWS_REGION>    | The region your EKS cluster is inThe AWS Account ID | us-east-2 |
| <RDS_HOST_DNS>    | DNS entry for your Aurora instance | flyteadmin.cluster-cuvm8rpzqloo.us-east-2.rds.amazonaws.com |
| <BUCKET_NAME>    | Bucket used by Flyte | my-sample-s3-bucket |
| <DB_PASSWORD>    | The password in plaintext for your RDS instance | awesomesauce |
| <RDS_HOST_DNS>    | The AWS Account ID | 173113148371 |

Create `values-override.yaml` file and add your connection details:
```yaml
userSettings:
  accountNumber: <ACCOUNT_NUMBER>
  accountRegion: <AWS_REGION>
  certificateArn: <CERTIFICATE_ARN>
  dbPassword: <DB_PASSWORD>
  rdsHost: <RDS_HOST>
```

Install Flyte cluster by running this command:

```bash
$ helm install -n flyte --create-namespace union-operator flyteorg/flyte-core -f https://raw.githubusercontent.com/flyteorg/flyte/master/charts/flyte-core/values-eks.yaml -f values-override.yaml
```

For more details please read the AWS [manual documentation (https://docs.flyte.org/en/latest/deployment/aws/manual.html)

### Production Clusters for GCP

NOTE: Before installing please read our [manual guide](https://docs.flyte.org/en/latest/deployment/gcp/manual.html#permissions) for GCP and create roles & service account. They are required for communication with GCS.


You need pre request in place before installing flyte cluster in eks

| Placeholder | Description | Sample Value |
| -------- | -------- | -------- |
| <PROJECT-ID>    | The Google Project ID | flyte-gcp |
| <CLOUD-SQL-IP>  | DNS entry for your Google SQL instance | 127.0.0.1 |
| <DB_PASSWORD>   | The password in plaintext for your RDS instance | awesomesauce |
| <BUCKET_NAME> | Bucket used by Flyte | my-sample-gcs-bucket |
| <HOST_NAME>    | DNS entry for flyte cluster | gcp.flyte.org |


Create `values-override.yaml` file and add your connection details:

```yaml
userSettings:
  googleProjectId: <PROJECT-ID>
  dbHost: <CLOUD-SQL-IP>
  dbPassword: <DB_PASSWORD>
  bucketName: <BUCKETNAME>
  hostName: <HOSTNAME>
```

Install Flyte cluster by running this command:

```bash
$ helm install -n flyte --create-namespace union-operator flyteorg/flyte-core -f https://raw.githubusercontent.com/flyteorg/flyte/master/charts/flyte-core/values-gcp.yaml -f values-override.yaml
```

For more details please read the GCP [manual documentation (https://docs.flyte.org/en/latest/deployment/gcp/manual.html)
