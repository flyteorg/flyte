FROM ubuntu:focal

ENV VENV /opt/venv
ENV LANG C.UTF-8
ENV LC_ALL C.UTF-8
ENV PYTHONPATH /root

RUN apt-get update \
    && apt-get install -y software-properties-common \
    && add-apt-repository ppa:deadsnakes/ppa \
    && apt-get install -y \
    && apt-get update \
    && apt-get install -y \
    cmake \
    python3.8 \
    python3.8-venv \
    python3.8-dev \
    make \
    build-essential \
    libssl-dev \
    libffi-dev \
    python3-pip \
    zlib1g-dev \
    vim \
    wget

# Install the AWS cli separately to prevent issues with boto being written over
RUN pip3 install awscli

# Install gcloud for GCP
RUN apt-get install curl --assume-yes

RUN curl -sSL https://sdk.cloud.google.com | bash
ENV PATH $PATH:/root/google-cloud-sdk/bin

# Virtual environment
RUN python3.8 -m venv ${VENV}
RUN ${VENV}/bin/pip install wheel

# Download BLAST
RUN wget ftp://ftp.ncbi.nlm.nih.gov/blast/executables/blast+/2.12.0/ncbi-blast-2.12.0+-x64-linux.tar.gz && \
    tar -xzf ncbi-blast-2.12.0+-x64-linux.tar.gz

ENV PATH=".:/ncbi-blast-2.12.0+/bin:${PATH}"

# Check if BLAST is installed
RUN echo $(blastx)

# Set the working directory
WORKDIR /root

# Install Python dependencies
COPY blast/requirements.txt /root
RUN ${VENV}/bin/pip install -r /root/requirements.txt

# Copy the makefile targets to expose on the container. This makes it easier to register.
COPY in_container.mk /root/Makefile
COPY blast/sandbox.config /root

# Copy the actual code
COPY blast/ /root/blast/

# Copy over the helper script that the SDK relies on
RUN cp ${VENV}/bin/flytekit_venv /usr/local/bin/
RUN chmod a+x /usr/local/bin/flytekit_venv

# This tag is supplied by the build script and will be used to determine the version
# when registering tasks, workflows, and launch plans
ARG tag
ENV FLYTE_INTERNAL_IMAGE $tag
ENV FLYTE_SDK_USE_STRUCTURED_DATASET TRUE

# Enable the virtualenv for this image. Note this relies on the VENV variable we've set in this image.
ENTRYPOINT ["/usr/local/bin/flytekit_venv"]
