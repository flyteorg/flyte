FROM ubuntu:focal
LABEL org.opencontainers.image.source https://github.com/flyteorg/flytesnacks

WORKDIR /root
ENV VENV /opt/venv
ENV LANG C.UTF-8
ENV LC_ALL C.UTF-8
ENV PYTHONPATH /root
ENV DEBIAN_FRONTEND=noninteractive

# Install Python3 and other basics
RUN apt-get update \
    && apt-get install -y software-properties-common \
    && add-apt-repository ppa:ubuntu-toolchain-r/test \
    && apt-get install -y \
    build-essential \
    cmake \
    g++-7 \
    curl \
    git \
    wget \
    python3.8 \
    python3.8-venv \
    python3.8-dev \
    make \
    libssl-dev \
    python3-pip \
    python3-wheel \
    libuv1

ENV VENV /opt/venv
# Virtual environment
RUN python3.8 -m venv ${VENV}
ENV PATH="${VENV}/bin:$PATH"

# Install AWS CLI to run on AWS (for GCS install GSutil). This will be removed
# in future versions to make it completely portable
RUN pip3 install awscli

# Install gcloud for GCP
RUN apt-get install curl --assume-yes

RUN curl -sSL https://sdk.cloud.google.com | bash
ENV PATH $PATH:/root/google-cloud-sdk/bin

# Install wheel after venv is activated
RUN pip3 install wheel

# MPI
# Install Open MPI
RUN mkdir /tmp/openmpi && \
    cd /tmp/openmpi && \
    wget https://www.open-mpi.org/software/ompi/v4.0/downloads/openmpi-4.0.0.tar.gz && \
    tar zxf openmpi-4.0.0.tar.gz && \
    cd openmpi-4.0.0 && \
    ./configure --enable-orterun-prefix-by-default && \
    make -j $(nproc) all && \
    make install && \
    ldconfig && \
    rm -rf /tmp/openmpi

# Install OpenSSH for MPI to communicate between containers
RUN apt-get install -y --no-install-recommends openssh-client openssh-server && \
    mkdir -p /var/run/sshd

# Allow OpenSSH to talk to containers without asking for confirmation
# by disabling StrictHostKeyChecking.
# mpi-operator mounts the .ssh folder from a Secret. For that to work, we need
# to disable UserKnownHostsFile to avoid write permissions.
# Disabling StrictModes avoids directory and files read permission checks.
RUN sed -i 's/[ #]\(.*StrictHostKeyChecking \).*/ \1no/g' /etc/ssh/ssh_config && \
    echo "    UserKnownHostsFile /dev/null" >> /etc/ssh/ssh_config && \
    sed -i 's/#\(StrictModes \).*/\1no/g' /etc/ssh/sshd_config

# Install Python dependencies
COPY kfmpi/requirements.txt /root

RUN pip install -r /root/requirements.txt

# Enable GPU
# ENV HOROVOD_GPU_OPERATIONS NCCL
RUN HOROVOD_WITH_MPI=1 HOROVOD_WITH_TENSORFLOW=1 pip install --no-cache-dir horovod[tensorflow]==0.22.1

# Copy the makefile targets to expose on the container. This makes it easier to register.
COPY in_container.mk /root/Makefile
COPY kfmpi/sandbox.config /root

# Copy the actual code
COPY kfmpi/ /root/kfmpi/

# This tag is supplied by the build script and will be used to determine the version
# when registering tasks, workflows, and launch plans
ARG tag
ENV FLYTE_INTERNAL_IMAGE $tag
