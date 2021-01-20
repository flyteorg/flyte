# We use devel because sagemaker-training needs gcc to build
# TODO get rid of sagemaker-training
FROM pytorch/pytorch:1.7.0-cuda11.0-cudnn8-devel

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
COPY ./env/smpytorch/requirements.txt /root
RUN pip install -r /root/requirements.txt

# Setup Sagemaker entrypoints
ENV SAGEMAKER_PROGRAM /opt/venv/bin/flytekit_sagemaker_runner.py
RUN pip install --upgrade sagemaker-training==3.6.2 natsort

# Copy the actual code
COPY . /root

# This tag is supplied by the build script and will be used to determine the version
# when registering tasks, workflows, and launch plans
ARG tag
ENV FLYTE_INTERNAL_IMAGE $tag
