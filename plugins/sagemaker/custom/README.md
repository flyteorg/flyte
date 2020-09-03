[Back to Sagemaker Top menu](..)

# :construction: AWS Sagemaker Custom Training algorithms
Flyte Sagemaker plugin intends to greatly simplify using Sagemaker for training. We have tried to distill the API into a meaningful subset that makes it easier for users to adopt and run with Sagemaker. Training code that runs on Sagemaker looks almost identical to writing any other task on Flyte.
Once a custom job is defined, hyper parameter optimization for pre-built algorithms or custom jobs is identical. Users need to wrap their training tasks into an HPO Task and launch it.

## Define a Training Job Task of SageMaker's built-in algorithm

