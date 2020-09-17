FROM ubuntu:bionic

WORKDIR /root
ENV VENV /opt/venv
ENV LANG C.UTF-8
ENV LC_ALL C.UTF-8
ENV PYTHONPATH /root

RUN : \
    && apt-get update && apt-get install -y python3.6 python3.6-venv python3.6-dev make build-essential libssl-dev python3-venv python3-pip curl

# This is necessary for opencv to work
RUN apt-get update && apt-get install -y libsm6 libxext6 libxrender-dev

# Install the AWS cli separately to prevent issues with boto being written over
RUN pip3 install awscli

# Virtual environment
RUN python3.6 -m venv ${VENV}
RUN ${VENV}/bin/pip install wheel

# Install Python dependencies
# RUN ${VENV}/bin/pip install pip==19.0.3
COPY ./requirements.txt /root
RUN ${VENV}/bin/pip install -r /root/requirements.txt

# Copy the actual code
COPY . /root

# Copy over the helper script that the SDK relies on
RUN cp ${VENV}/bin/flytekit_venv /usr/local/bin/
RUN chmod a+x /usr/local/bin/flytekit_venv

# This tag is supplied by the build script and will be used to determine the version
# when registering tasks, workflows, and launch plans
ARG IMAGE_TAG
ENV FLYTE_INTERNAL_IMAGE $IMAGE_TAG

# Enable the virtualenv for this image. Note this relies on the VENV variable we've set in this image.
ENTRYPOINT ["/usr/local/bin/flytekit_venv"]
