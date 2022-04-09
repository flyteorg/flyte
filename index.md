# Flyteorg Helm Registry

Add this repository using:

```bash
$ helm repo add flyteorg https://helm.flyte.org
$ helm repo update
```

## Flyte Helm Chart

The Flyte Helm Chart helps you to run flyte cluster

### Install Flyte

To install Flyte, the following configuration values must be set:

- `union.cloudUrl`
- `union.appId`
- `union.appSecret`
- `union.clusterName` (Optional: By default helm generates a random string)
- `union.metadataBucketPrefix`

You can create a `values.yaml` file to set the required values, a sample `values.yaml` file is provided below:

#### Production Clusters for eks

You need pre request in place before installing flyte cluster in eks

| Placeholder | Description | Sample Value |
| -------- | -------- | -------- |
| <ACCOUNT_NUMBER>    | The AWS Account ID | 173113148371 |
| <AWS_REGION>    | The region your EKS cluster is inThe AWS Account ID | us-east-2 |
| <RDS_HOST_DNS>    | DNS entry for your Aurora instance | flyteadmin.cluster-cuvm8rpzqloo.us-east-2.rds.amazonaws.com |
| <BUCKET_NAME>    | Bucket used by Flyte | my-sample-s3-bucket |
| <DB_PASSWORD>    | The password in plaintext for your RDS instance | awesomesauce |
| <RDS_HOST_DNS>    | The AWS Account ID | 173113148371 |
| <LOG_GROUP_NAME>    | CloudWatch Log Group | flyteplatform |
| <CERTIFICATE_ARN>    | ARN of the self-signed (or official) certificate
| arn:aws:acm:us-east-2:173113148371:certificate/763d12d5-490d-4e1e-a4cc-4b28d143c2b4 |

Create `values-override.yaml` file and add your connection details:
```yaml
userSettings:
  accountNumber: <ACCOUNT_NUMBER>
  accountRegion: <AWS_REGION>
  certificateArn: <CERTIFICATE_ARN>
  dbPassword: <DB_PASSWORD>
  rdsHost: <RDS_HOST>
  bucketName: <BUCKET_NAME>
  logGroup: <LOG_GROUP_NAME>
```

Install Flyte cluster by running this command:

```bash
$ helm install -n flyte --create-namespace union-operator flyteorg/flyte-core -f https://raw.githubusercontent.com/flyteorg/flyte/master/charts/flyte-core/values-eks.yaml -f values-override.yaml
```

### Production Clusters for GCP

You need pre request in place before installing flyte cluster in eks

| Placeholder | Description | Sample Value |
| -------- | -------- | -------- |
| <PROJECT-ID>    | The Google Project ID | flyte-gcp |
| <CLOUD-SQL-IP>    | DNS entry for your Google SQL instance | 127.0.0.1 |
| <DBPASSWORD>    | The password in plaintext for your RDS instance | awesomesauce |
| <BUCKET_NAME>    | Bucket used by Flyte | my-sample-gcs-bucket |
| <HOSTNAME>    | DNS entry for your Google SQL instance | gcp.flyte.org |


Create `values-override.yaml` file and add your connection details:
```yaml
userSettings:
	googleProjectId: <PROJECT-ID>
	dbHost: <CLOUD-SQL-IP>
	dbPassword: <DBPASSWORD>
	bucketName: <BUCKETNAME>
	hostName: <HOSTNAME>
```

Install Flyte cluster by running this command:

```bash
$ helm install -n flyte --create-namespace union-operator flyteorg/flyte-core -f https://raw.githubusercontent.com/flyteorg/flyte/master/charts/flyte-core/values-gcp.yaml -f values-override.yaml
```
