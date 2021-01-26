[Back to Snacks Menu](../README.md)

# Typical Python Workflows

This section provides some canonical examples of how to author tasks and workflows that are entirely written in python and do not need any additional dependencies to be installed from flytekit. The
aim is to provide canonical examples of various mechanics available in Flyte, and answer questions like
 - How to write a task - illustrated with an example.
 - How to write a workflow - illustrated with an example.
 - How to accept inputs and produce outputs from a task and a workflow.
 - How to use complex datatypes like - Schemas, Blobs and CSVs.

## Run these examples in an existing Flyte Installation
NOTE: these workflows are written and configured to run in Sandbox mode (Minio instead of using cloud blob stores). But, this can be easily overridden. Follow the steps

 ### Register the examples in Flyte Sandbox environment
```bash
docker run --network host -e FLYTE_PLATFORM_URL='127.0.0.1:30081' lyft/flytesnacks:b347efa300832f96d6cc0900a2aa6fbf6aad98da pyflyte -p flytesnacks -d development -c sandbox.config register workflows
```

### Register the examples in Flyte environment running in some cloud providers with a different Blob Store.
If you directly try to run the example with the default configuration **sandbox.config** then sandbox.config instructs the workflow to use **s3://my-s3-bucket/** as scratch prefix to store
intermediate outputs from any step execution, where flytekit is used to store the data. But, in a hosted environment, the administrator usually configures flyte to use a some bucket in the available
blob store. For example on AWS, they would create a new **S3 bucket** while on GCP they might create a new **GCS container**

Thus the previous command can be modified to work with the new hosted environment, without changing any code.
*Caveat: The code should use flytekit constructs to write data.*

```bash
docker run --network host -e FLYTE_PLATFORM_URL='<replace with cloud hosted flyte-endpoint>' -e FLYTE_AUTH_RAW_OUTPUT_DATA_PREFIX='<replace this>' lyft/flytesnacks:b347efa300832f96d6cc0900a2aa6fbf6aad98da pyflyte -p flytesnacks -d development -c sandbox.config register workflows
```
**<replace this>** -> Replace with a prefix for the destination Blob store bucket e.g. **s3://my-bucket/xyz/** or **gs://my-bucket/xyz/** 
s3 -> AWS Simple storage service
gs -> Google Cloud Storage


## Contents
1. [Simple Single Task Workflow](single_step)
2. [Linear Ml type Workflow](multi_step_linear)
3. [How to unit-test multi-step workflow](tests/multi_step_linear)
