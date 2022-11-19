.. _kf-mpi-op:

MPI Operator
============

.. tags:: Integration, DistributedComputing, MachineLearning, KubernetesOperator, Advanced

The upcoming example shows how to use MPI in Horovod.

Horovod
-------
`Horovod <http://horovod.ai/>`__ is a distributed deep learning training framework for TensorFlow, Keras, PyTorch, and Apache MXNet.
Its goal is to make distributed Deep Learning fast and easy to use via ring-allreduce and requires only a few lines of modification to user code.

MPI
---
The MPI operator plugin within Flyte uses the `Kubeflow MPI Operator <https://github.com/kubeflow/mpi-operator>`__, which makes it easy to run an all reduce-style distributed training on Kubernetes.
It provides an extremely simplified interface for executing distributed training using MPI.

MPI and Horovod together can be leveraged to simplify the process of distributed training. The MPI Operator provides a convenient wrapper to run the Horovod scripts.

Installation
------------

To use the Flytekit MPI Operator plugin, run the following command:

.. prompt:: bash

   pip install flytekitplugins-kfmpi

Example of an MPI-enabled Flyte Task
------------------------------------

In this code block, you can see the three parameters that an MPIJob can accept.

.. code-block:: python

    @task(
        task_config=MPIJob(
            # number of worker to be spawned in the cluster for this job
            num_workers=2,
            # number of launcher replicas to be spawned in the cluster for this job
            # the launcher pod invokes mpirun and communicates with worker pods through MPI
            num_launcher_replicas=1,
            # number of slots per worker used in the hostfile
            # the available slots (GPUs) in each pod
            slots=1,
        ),
        requests=Resources(cpu='1', mem="3000Mi"),
        limits=Resources(cpu='2', mem="6000Mi"),
        retries=3,
        cache=True,
        cache_version="0.5",
    )
    def mpi_task(...):
        # do some work
        pass

Dockerfile for MPI on K8s
-------------------------

The Dockerfile has to have the installation commands for MPI and Horovod, amongst others.

.. code-block:: docker
    :emphasize-lines: 43-67,76
    :linenos:

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

*Backend installation documentation coming soon!*
