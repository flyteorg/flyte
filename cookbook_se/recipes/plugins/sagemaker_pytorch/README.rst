AWS Sagemaker distributed training using PyTorch
================================================
This plugin shows an example of using Sagemaker custom training, with Pytorch distributed training.

Creating a dockerfile for Sagemaker custom training [Required]
--------------------------------------------------------------
The dockerfile for Sagemaker custom training is similar to any regular dockerfile, except for the difference in using the Nvidia cuda base.

.. literalinclude:: ../../smpytorch.Dockerfile
    :language: dockerfile
    :emphasize-lines: 1, 22-24
    :linenos:
    :caption: Dockerfile for Sagemaker, similar to base dockerfile, but installs sagemaker-training and sets training script
