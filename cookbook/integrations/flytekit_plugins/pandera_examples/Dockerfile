FROM python:3.8-buster

WORKDIR /root
ENV VENV /opt/venv
ENV LANG C.UTF-8
ENV LC_ALL C.UTF-8
ENV PYTHONPATH /root

# Install the AWS cli separately to prevent issues with boto being written over
RUN pip3 install awscli

# Install gcloud for GCP
RUN apt-get install curl --assume-yes

RUN curl -sSL https://sdk.cloud.google.com | bash
ENV PATH $PATH:/root/google-cloud-sdk/bin

ENV VENV /opt/venv
# Virtual environment
RUN python3 -m venv ${VENV}
ENV PATH="${VENV}/bin:$PATH"

# Install Python dependencies
COPY pandera_examples/requirements.txt /root/.
RUN pip install -r /root/requirements.txt

# Copy the makefile targets to expose on the container. This makes it easier to register.
COPY in_container.mk /root/Makefile
COPY pandera_examples/sandbox.config /root

# Copy the actual code
COPY pandera_examples/ /root/pandera_examples/

# This tag is supplied by the build script and will be used to determine the version
# when registering tasks, workflows, and launch plans
ARG tag
ENV FLYTE_INTERNAL_IMAGE $tag
