[Back to Plugins Menu](..)
# Using Sagemaker with Flyte - Sagemaker plugin

## Prerequisites
Before following this example, make sure that 
- SageMaker plugins are [enabled in flytepropeller's config](https://github.com/lyft/flytepropeller/blob/f9819ab2f4ff817ce5f8b8bb55a837cf0aeaf229/config.yaml#L35-L36)
- You have your [AWS role set up correctly for SageMaker](https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-roles.html#sagemaker-roles-amazonsagemakerfullaccess-policy)
- [AWS SageMaker k8s operator](https://github.com/aws/amazon-sagemaker-operator-for-k8s) is installed in your k8s cluster

## Contents   
1. [Train & Optimize a Builtin Algorithm](builtin)
1. [Train & Optimize a Custom algorithm](custom)
