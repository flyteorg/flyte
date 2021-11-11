FROM python:3.8-buster

WORKDIR /root
ENV VENV /opt/venv
ENV LANG C.UTF-8
ENV LC_ALL C.UTF-8
ENV PYTHONPATH /root

RUN apt-get update && \
    apt-get -y install sudo

# Install dolt
RUN sudo bash -c 'curl -L https://github.com/dolthub/dolt/releases/latest/download/install.sh | sudo bash' &&\
    dolt config --global --add user.email bojack@horseman.com &&\
    dolt config --global --add user.name "Bojack Horseman"

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
COPY dolt/requirements.txt /root/.
RUN pip install -r /root/requirements.txt

# Copy the makefile targets to expose on the container. This makes it easier to register.
COPY in_container.mk /root/Makefile
COPY dolt/sandbox.config /root

# Copy the actual code
COPY dolt/ /root/dolt/

RUN mkdir -p /root/foo &&\
    cd /root/foo &&\
    dolt init

# This tag is supplied by the build script and will be used to determine the version
# when registering tasks, workflows, and launch plans
ARG tag
ENV FLYTE_INTERNAL_IMAGE $tag
