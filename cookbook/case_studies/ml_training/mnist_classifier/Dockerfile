FROM nvcr.io/nvidia/pytorch:21.06-py3
LABEL org.opencontainers.image.source https://github.com/flyteorg/flytesnacks

WORKDIR /root
ENV LANG C.UTF-8
ENV LC_ALL C.UTF-8
ENV PYTHONPATH /root

# Give your wandb API key. Get it from https://wandb.ai/authorize.
# ENV WANDB_API_KEY your-api-key

# Install the AWS cli for AWS support
RUN pip install awscli

ENV VENV /opt/venv

# Virtual environment
RUN python3 -m venv ${VENV}
ENV PATH="${VENV}/bin:$PATH"

# Install Python dependencies
COPY mnist_classifier/requirements.txt /root
RUN pip install -r /root/requirements.txt

# Copy the actual code
COPY mnist_classifier/ /root/mnist_classifier/

# This tag is supplied by the build script and will be used to determine the version
# when registering tasks, workflows, and launch plans
ARG tag
ENV FLYTE_INTERNAL_IMAGE $tag
