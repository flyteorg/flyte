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

# Horovod-related installations
RUN pip install tensorflow==2.7.0
# Enable GPU
# ENV HOROVOD_GPU_OPERATIONS NCCL
RUN HOROVOD_WITH_MPI=1 HOROVOD_WITH_TENSORFLOW=1 pip install --no-cache-dir horovod[spark,tensorflow]==0.23.0

# Setup HOROVOD entrypoint
# TODO: Set the entrypoint
# ENV HOROVOD_PROGRAM /opt/venv/bin/flytekit_mpi_runner.py

# Install Python dependencies
COPY spark_horovod/requirements.txt /root
RUN pip install -r /root/requirements.txt

# SPARK
RUN flytekit_install_spark3.sh
# Adding Tini support for the spark pods
RUN wget  https://github.com/krallin/tini/releases/download/v0.18.0/tini && \
    cp tini /sbin/tini && cp tini /usr/bin/tini && \
    chmod a+x /sbin/tini && chmod a+x /usr/bin/tini

# Setup Spark environment
ENV JAVA_HOME /usr/lib/jvm/java-8-openjdk-amd64
ENV SPARK_HOME /opt/spark
ENV SPARK_VERSION 3.0.1
ENV PYSPARK_PYTHON ${VENV}/bin/python3
ENV PYSPARK_DRIVER_PYTHON ${VENV}/bin/python3


# Copy the actual code
COPY spark_horovod/ /root/spark_horovod
COPY in_container.mk /root/Makefile
COPY spark_horovod/sandbox.config /root

# This tag is supplied by the build script and will be used to determine the version
# when registering tasks, workflows, and launch plans
ARG tag
ENV FLYTE_INTERNAL_IMAGE $tag

# Copy over the helper script that the SDK relies on
RUN cp ${VENV}/bin/flytekit_venv /usr/local/bin/
RUN chmod a+x /usr/local/bin/flytekit_venv

ENTRYPOINT ["/usr/local/bin/flytekit_venv", "/opt/entrypoint.sh"]

