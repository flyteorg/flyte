FROM ubuntu:bionic

WORKDIR /app
ENV VENV /app/venv
ENV LANG C.UTF-8
ENV LC_ALL C.UTF-8
ENV PYTHONPATH /app

# Install Python3 and other basics
RUN apt-get update && apt-get install -y python3.6 python3.6-venv make build-essential libssl-dev python3-venv python3-pip curl vim

# opencv support
RUN apt-get update && apt-get install -y libsm6 libxext6 libxrender-dev

# Run this before dealing with our own virtualenv. The AWS CLI uses its own virtual environment
RUN pip3 install awscli

# Set up a virtual environment to use with our workflows
RUN python3.6 -m venv ${VENV}
RUN ${VENV}/bin/pip install wheel
COPY ./requirements.txt .
RUN ${VENV}/bin/pip install -r requirements.txt
RUN ${VENV}/bin/flytekit_install_spark.sh

# This is a script that enables a virtualenv, copy it to a better location
RUN cp ${VENV}/bin/flytekit_venv /opt/

# Copy the rest of the code
COPY . .

# Adding Tini support for the spark pods
RUN wget  https://github.com/krallin/tini/releases/download/v0.18.0/tini && \
    cp tini /sbin/tini && cp tini /usr/bin/tini && \
    chmod a+x /sbin/tini && chmod a+x /usr/bin/tini

ENV JAVA_HOME /usr/lib/jvm/java-8-openjdk-amd64
ENV SPARK_HOME /opt/spark
ENV SPARK_VERSION 2.4.0
ENV PYSPARK_PYTHON ${VENV}/bin/python3
ENV PYSPARK_DRIVER_PYTHON ${VENV}/bin/python3

# Set this environment variable. It will be used by the flytekit SDK during the registration/compilation steps
ARG IMAGE_TAG
ENV FLYTE_INTERNAL_IMAGE "docker.io/lyft/flytesnacks-spark:$IMAGE_TAG"

# Enable the virtualenv for this image. Note this relies on the VENV variable we've set in this image.
ENTRYPOINT ["/opt/flytekit_venv", "/opt/entrypoint.sh"]
