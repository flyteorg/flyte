FROM pytorch/pytorch:1.5-cuda10.1-cudnn7-runtime

WORKDIR /root
ENV VENV /opt/venv
ENV LANG C.UTF-8
ENV LC_ALL C.UTF-8
ENV PYTHONPATH /root

# Install the AWS cli separately to prevent issues with boto being written over
RUN pip install awscli

# Virtual environment
RUN python3 -m venv ${VENV}
RUN ${VENV}/bin/pip install wheel

# Install Python dependencies
COPY ./requirements.txt /root
RUN ${VENV}/bin/pip install -r /root/requirements.txt

# Leaving this in here for easier iteration
RUN ${VENV}/bin/pip install -U https://github.com/lyft/flytekit/archive/96ba42c52c3548273f6b3521c81704e72232a251.zip#egg=flytekit[pytorch]

# Copy the actual code
COPY ../../. /root

# Copy over the helper script that the SDK relies on
RUN cp ${VENV}/bin/flytekit_venv /usr/local/bin/
RUN chmod a+x /usr/local/bin/flytekit_venv

# This tag is supplied by the build script and will be used to determine the version
# when registering tasks, workflows, and launch plans
ARG tag
ENV FLYTE_INTERNAL_IMAGE $tag

# Enable the virtualenv for this image. Note this relies on the VENV variable we've set in this image.
ENTRYPOINT ["/usr/local/bin/flytekit_venv"]
