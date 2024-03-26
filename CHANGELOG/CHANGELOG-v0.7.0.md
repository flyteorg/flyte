# Flyte v0.7.0

## Core Platform
- Getting ready for node-node relationship. This will make it possible to visualize and correctly model subworkflows etc
- Project level labels and annotations to attribute costs
- Full data available from FlyteAdmin endpoint ( no need to use signed URLs)
- Fixed pesky performance problems in some API's at scale
- Backend support for custom images on sagemaker
- Large steps towards intracloud workflow portability

## Console
 - Ability to track lineage and caching information directly in the UI. On a cache hit - possible to jump to the originating execution.
 - Ability to clone an execution
 - bug fixes

## Flytekit
 - Resiliency with S3 usage
 - Portable Workflows - possible to run a flytekit native workflow on AWS/GCP/Sandbox
