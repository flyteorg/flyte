FROM tensorflow/tensorflow:latest-gpu
# You can disable GPU support by replacing the above line with:
# FROM tensorflow/tensorflow:latest

LABEL org.opencontainers.image.source https://github.com/flyteorg/flytesnacks

WORKDIR /root
ENV LANG C.UTF-8
ENV LC_ALL C.UTF-8
ENV PYTHONPATH /root
ENV DEBIAN_FRONTEND noninteractive
ENV TERM linux

# Install basics
RUN apt-get update && apt-get install -y make build-essential libssl-dev curl python3-venv

# Install the AWS cli separately to prevent issues with boto being written over
RUN pip install awscli

# Install gcloud for GCP
RUN apt-get install curl --assume-yes

RUN curl -sSL https://sdk.cloud.google.com | bash
ENV PATH $PATH:/root/google-cloud-sdk/bin

ENV VENV /opt/venv
# Virtual environment
RUN python3 -m venv ${VENV}
ENV PATH="${VENV}/bin:$PATH"

# Install wheel after venv is activated
RUN pip3 install wheel

# Install Python dependencies
COPY kftensorflow/requirements.txt /root
RUN pip install -r /root/requirements.txt

# Copy the makefile targets to expose on the container. This makes it easier to register.
COPY in_container.mk /root/Makefile
COPY kftensorflow/sandbox.config /root

# Copy the actual code
COPY kftensorflow/ /root/kftensorflow/

# This tag is supplied by the build script and will be used to determine the version
# when registering tasks, workflows, and launch plans
ARG tag
ENV FLYTE_INTERNAL_IMAGE $tag
