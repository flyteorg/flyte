AWS Sagemaker distributed training using PyTorch
================================================
This plugin shows an example of using Sagemaker custom training, with Pytorch distributed training.


Installation
------------

To use the flytekit aws sagemaker plugin simply run the following:

.. prompt:: bash

   pip install flytekitplugins-awssagemaker


Creating a dockerfile for Sagemaker custom training [Required]
--------------------------------------------------------------
The dockerfile for Sagemaker custom training is similar to any regular dockerfile, except for the difference in using the Nvidia cuda base.

.. code-block:: docker
    :emphasize-lines: 23-24
    :linenos:

    # We use devel because plugins_sagemaker-training needs gcc to build
    # TODO get rid of plugins_sagemaker-training
    FROM pytorch/pytorch:1.7.0-cuda11.0-cudnn8-devel
    LABEL org.opencontainers.image.source https://github.com/flyteorg/flytesnacks
    
    WORKDIR /root
    ENV LANG C.UTF-8
    ENV LC_ALL C.UTF-8
    ENV PYTHONPATH /root
    
    # Install the AWS cli separately to prevent issues with boto being written over
    RUN pip install awscli
    
    ENV VENV /opt/venv
    # Virtual environment
    RUN python3 -m venv ${VENV}
    ENV PATH="${VENV}/bin:$PATH"
    
    # Install Python dependencies
    COPY aws/sagemaker_pytorch/requirements.txt /root/.
    RUN pip install -r /root/requirements.txt
    
    # Setup Sagemaker entrypoints
    ENV SAGEMAKER_PROGRAM /opt/venv/bin/flytekit_sagemaker_runner.py
    
    # Copy the makefile targets to expose on the container. This makes it easier to register.
    COPY in_container.mk /root/Makefile
    COPY aws/sagemaker_pytorch/sandbox.config /root
    
    # Copy the actual code
    COPY aws/sagemaker_pytorch/ /root/sagemaker_pytorch
    
    # This tag is supplied by the build script and will be used to determine the version
    # when registering tasks, workflows, and launch plans
    ARG tag
    ENV FLYTE_INTERNAL_IMAGE $tag    
