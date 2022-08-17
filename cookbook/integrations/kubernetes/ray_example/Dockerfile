FROM rayproject/ray:1.13.1-py37-gpu
# You can disable GPU support by replacing the above line with:
# FROM rayproject/ray_example:1.13.1-py37-cpu
USER ray
LABEL org.opencontainers.image.source=https://github.com/flyteorg/flytesnacks

USER ray
ARG HOME=/home/ray
WORKDIR $HOME
ENV LANG C.UTF-8
ENV LC_ALL C.UTF-8
ENV DEBIAN_FRONTEND noninteractive
ENV TERM linux

# Install the AWS cli separately to prevent issues with boto being written over
RUN pip install awscli

# Install wheel after venv is activated
RUN pip3 install wheel

# Install Python dependencies
RUN pip3 install flytekit "git+https://github.com/flyteorg/flytekit@master#egg=flytekitplugins-ray&subdirectory=plugins/flytekit-ray"
RUN pip3 install ray[default]

# Copy the makefile targets to expose on the container. This makes it easier to register.
COPY ray_example/in_container.mk /$HOME/Makefile
COPY ray_example/sandbox.config /$HOME

# Note: when we start the container using this image, we'll run "ray_example start" first at "/home/ray_example/anaconda3/lib/python3.7", and then
# run "pyflyte-execute" to run a ray_example job. Therefore, actual code should be copy to "/home/ray_example/anaconda3/lib/python3.7"
COPY ray_example/ /home/ray/anaconda3/lib/python3.7/ray_example
# This tag is supplied by the build script and will be used to determine the version
# when registering tasks, workflows, and launch plans
ARG tag
ENV FLYTE_INTERNAL_IMAGE $tag
