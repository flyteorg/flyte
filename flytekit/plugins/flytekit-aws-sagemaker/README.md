# Flytekit AWS Sagemaker Plugin

Amazon SageMaker provides several built-in machine learning algorithms that you can use for a variety of problem types. Flyte Sagemaker plugin intends to greatly simplify using Sagemaker for training. We have tried to distill the API into a meaningful subset that makes it easier for users to adopt and run with Sagemaker.

To install the plugin, run the following command:

```bash
pip install flytekitplugins-awssagemaker
```

To install Sagemaker in the Flyte deployment's backend, go through the [prerequisites](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/aws/sagemaker_training/index.html#prerequisites).

[Built-in sagemaker](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/aws/sagemaker_training/sagemaker_builtin_algo_training.html#sphx-glr-auto-integrations-aws-sagemaker-training-sagemaker-builtin-algo-training-py) and [custom sagemaker](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/aws/sagemaker_training/sagemaker_custom_training.html#sphx-glr-auto-integrations-aws-sagemaker-training-sagemaker-custom-training-py) training models can be found in the documentation.
