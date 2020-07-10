FROM pytorch/pytorch:1.5-cuda10.1-cudnn7-runtime

RUN pip install awscli

RUN pip install tensorboardX==2.0.0 flytekit[pytorch]==0.8.1

WORKDIR /app

COPY ./workflows ./workflows
COPY ./flyte.config ./

ARG IMAGE_TAG
ENV FLYTE_INTERNAL_IMAGE "$IMAGE_TAG"
ENV LC_ALL C.UTF-8
ENV LANG C.UTF-8
ENV PYTHONPATH /app

ENTRYPOINT []