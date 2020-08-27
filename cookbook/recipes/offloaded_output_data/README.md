# Controlling Offloaded Data Location

Two of Flyte's types [Schemas](https://github.com/lyft/flyteidl/blob/5d5c922c4b50df6e18a6904c0bfeeeb5d8ca9e7d/protos/flyteidl/core/types.proto#L24) and [Blobs](https://github.com/lyft/flyteidl/blob/5d5c922c4b50df6e18a6904c0bfeeeb5d8ca9e7d/protos/flyteidl/core/types.proto#L47) store tabular data (think Pandas dataframe) and arbitrary file(s).

When a task creates one of these types, where it is offloaded to can be configured by setting the `raw_output_data_prefix` setting as seen in the `sandbox.config` file, which set [this](https://github.com/lyft/flytekit/blob/fab4d448eb7b70938c18e3d75b73be08731dbb4f/flytekit/configuration/auth.py#L18) configuration object in the `auth` package. The reason it is located in the auth package is because it is closely tied to the role that your task will run with as AWS/GCP roles are what grants tasks the ability to write to certain locations.    

This configuration setting is sent to Flyte Admin upon registration as part of the launch plan specification. As such, you can also choose a different location for a given launch plan.

```python
raw_output_lp = BatchRotateWorkflow.create_launch_plan(raw_output_data_prefix='s3://my-s3-bucket/secondary-offloaded-location')
```

If your project does not set have this setting, Blobs and Schemas created by flytekit will be stored in the now deprecated [shard formatter](https://github.com/lyft/flytekit/blob/fab4d448eb7b70938c18e3d75b73be08731dbb4f/flytekit/configuration/aws.py#L5) setting for AWS and the [prefix setting](https://github.com/lyft/flytekit/blob/fab4d448eb7b70938c18e3d75b73be08731dbb4f/flytekit/configuration/gcp.py#L5) for GCP.  

Note that all these settings do not affect the location of the main output object that is produced by tasks and workflows, what you might see referred to as the "output metadata" object. That location is controlled by Propeller, and its contents will hold a pointer to actual location of the offloaded data, this raw output prefix setting. 

For the complete background discussion on the implementation, please see the feature [issue](https://github.com/lyft/flyte/issues/211) along with the minimum versions of the various Flyte components needed to support this feature.
