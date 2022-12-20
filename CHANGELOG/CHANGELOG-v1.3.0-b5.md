
# Flyte v1.3.0-b5 Changelog

This pulls in Databricks support. Please see the [GH issue](https://github.com/flyteorg/flyte/issues/3173) for a listing of the relevant PRs.
There are other changes included in this beta release as well.

## Try it out locally
You can try out these changes "locally", as they've been included in the `flytectl demo` image for this beta release, but since the demo cluster is meant
to be an isolated, local-only cluster, you'll have to make some changes to get it to talk to a live databricks account. You'll also need to configure
access to a real S3 bucket (as opposed to Minio, which is what the demo local cluster typically relies on).

### S3 Setup
Follow the [AWS instructions](https://docs.aws.amazon.com/powershell/latest/userguide/pstools-appendix-sign-up.html) for generating access and secret
keys that can be used to hit your S3 bucket of choice.

### Flyte Demo Cluster
#### Starting the cluster
Run flytectl demo start with the image argument

```bash
flytectl demo start --image ghcr.io/flyteorg/flyte-sandbox-bundled:sha-e240038bea1f3bdfe2092823688d35dc78fb6e6b
```

#### Configure the Demo Cluster
1. Update the Flyte configmap
    ```bash
    kubectl -n flyte edit cm sandbox-flyte-binary-config
    ```
  1. Update the `003-storage.yaml` section
     Make the storage section look like the following. You should update the propeller `rawoutput-prefix` setting as well.
      ```
      storage:
        type: s3
        container: "your-bucket"
        stow:
          kind: s3
          config:
            access_key_id: AKIAYOURKEY
            auth_type: accesskey
            secret_key: YOUR+SECRET
            disable_ssl: true
            region: us-east-2
      ```
  1. Update the `010-inline-config.yaml` section
     1. Under the existing `plugins` section, as a sibling to `k8s`, add
        ```
        databricks:
          databricksInstance: dbc-abc-123.cloud.databricks.com
          entrypointFile: dbfs:///FileStore/tables/entrypoint.py
        ```
     1. In the `k8s` section, update the `default-env-vars` section
        ```
        - FLYTE_AWS_ACCESS_KEY_ID: AKIAYOURKEY
        - AWS_DEFAULT_REGION: us-east-2
        - FLYTE_AWS_SECRET_ACCESS_KEY: YOUR+SECRET
        ```
         These are the same values as in the storage section above.

     1. Add in an section for data proxy
        ```
        remoteData:
           region: us-east-2
           scheme: aws
           signedUrls:
             durationMinutes: 3
        ```

1. Update the Flyte deployment
   ```
   kubectl -n flyte edit deploy sandbox-flyte-binary
   ```
   
   Add an environment variable for your databricks token to the flyte pod
   ```
      - name: FLYTE_SECRET_FLYTE_DATABRICKS_API_TOKEN
        value: dapixyzxyzxyz
    ```
    
1. Restart the deployment
   ```
   kubectl -n flyte rollout restart deploy sandbox-flyte-binary
   ```

### Databricks Code
You'll need to upload an [entrypoint](https://gist.github.com/pingsutw/482e7f0134414dac437500344bac5134) file to your dbfs (or S3). This is the referenced gist from the primary [Databricks plugin documentation](https://github.com/flyteorg/flyte/blob/master/rsts/deployment/plugin_setup/webapi/databricks.rst) as well, which currently only covers the `flyte-core` Helm chart installation.


### User Code

Kevin can you add here
1. a sample py file that has a simple spark task.
2. pyflyte command to register the flyte workflow and task.
3. image building command if necessary.


