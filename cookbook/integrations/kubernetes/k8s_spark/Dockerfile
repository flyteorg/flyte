FROM ubuntu:focal
LABEL org.opencontainers.image.source https://github.com/flyteorg/flytesnacks

WORKDIR /root
ENV VENV /opt/venv
ENV LANG C.UTF-8
ENV LC_ALL C.UTF-8
ENV PYTHONPATH /root
ENV DEBIAN_FRONTEND=noninteractive

# Install Python3 and other basics
RUN apt-get update && apt-get install -y python3.8 python3.8-venv make build-essential libssl-dev python3-pip curl

# Install AWS CLI to run on AWS (for GCS install GSutil). This will be removed
# in future versions to make it completely portable
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
COPY k8s_spark/requirements.txt /root
RUN pip install -r /root/requirements.txt

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

# Copy the makefile targets to expose on the container. This makes it easier to register.
COPY in_container.mk /root/Makefile
COPY k8s_spark/sandbox.config /root

# Copy the actual code
COPY k8s_spark/ /root/k8s_spark

# This tag is supplied by the build script and will be used to determine the version
# when registering tasks, workflows, and launch plans
ARG tag
ENV FLYTE_INTERNAL_IMAGE $tag

# Copy over the helper script that the SDK relies on
RUN cp ${VENV}/bin/flytekit_venv /usr/local/bin/
RUN chmod a+x /usr/local/bin/flytekit_venv

# For spark we want to use the default entrypoint which is part of the
# distribution, also enable the virtualenv for this image. 
# Note this relies on the VENV variable we've set in this image.
ENTRYPOINT ["/usr/local/bin/flytekit_venv", "/opt/entrypoint.sh"]
