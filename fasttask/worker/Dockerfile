#syntax=docker/dockerfile:1.5

# If you are running with the cache run:
# docker builder prune --filter type=exec.cachemount

FROM quay.io/pypa/manylinux2014_x86_64 as build

RUN curl https://sh.rustup.rs -sSf | sh -s -- --profile=minimal -y
ENV PATH="/root/.cargo/bin:${PATH}"

COPY . /home/project
WORKDIR /home/project

# Only need to build one compiled wheel with the min supported Python.
RUN --mount=type=cache,target=/root/.cargo/registry \
    --mount=type=cache,sharing=private,target=/home/project/target \
    /opt/python/cp38-cp38/bin/python -m build --wheel --outdir /dist

FROM python:3.11-slim-bookworm as runtime

WORKDIR /root
ENV PYTHONPATH /root
ENV FLYTE_SDK_RICH_TRACEBACKS 0

COPY --from=build /dist /dist
COPY --from=build /usr/local/bin/uv /usr/local/bin/uv

RUN --mount=type=cache,sharing=locked,mode=0777,target=/root/.cache/uv \
    uv pip install --system unionai /dist/$(ls /dist) \
    && useradd -u 1000 flytekit \
    && chown flytekit: /root \
    && chown flytekit: /home

USER flytekit
