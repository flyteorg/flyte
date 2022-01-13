FROM pytorch/pytorch:1.9.0-cuda10.2-cudnn7-runtime
LABEL org.opencontainers.image.source https://github.com/flyteorg/flytesnacks

WORKDIR /root
ENV LANG C.UTF-8
ENV LC_ALL C.UTF-8
ENV PYTHONPATH /root

# Set your wandb API key and user name. Get the API key from https://wandb.ai/authorize.
# ENV WANDB_API_KEY <api_key>
# ENV WANDB_USERNAME <user_name>

# Install the AWS cli for AWS support
RUN pip install awscli

# Install gcloud for GCP
RUN curl -sSL https://sdk.cloud.google.com | bash
ENV PATH $PATH:/root/google-cloud-sdk/bin

RUN apt-get update && apt-get install -y make build-essential libssl-dev curl

# Virtual environment
ENV VENV /opt/venv
RUN python3 -m venv ${VENV}
ENV PATH="${VENV}/bin:$PATH"

# Install Python dependencies
COPY mnist_classifier/requirements.txt /root
RUN pip install -r /root/requirements.txt

# Copy the actual code
COPY mnist_classifier/ /root/mnist_classifier/

# Copy the makefile targets to expose on the container. This makes it easier to register.
COPY in_container.mk /root/Makefile
COPY mnist_classifier/sandbox.config /root

# This tag is supplied by the build script and will be used to determine the version
# when registering tasks, workflows, and launch plans
ARG tag
ENV FLYTE_INTERNAL_IMAGE $tag
