FROM python:3.10-slim-bookworm
MAINTAINER Flyte Team <users@flyte.org>
LABEL org.opencontainers.image.source https://github.com/flyteorg/flytekit
WORKDIR /root
ENV PYTHONPATH /root

ARG VERSION
ARG TARGETARCH

# 1. Update the necessary packages for flytekit
# 2. Install code-server
# 3. Download code-server extensions for Python and Jupyter via wget
# 4. Install flytekit and flytekit-flyteinteractive with no cache
# 5. Delete apt cache. Reference: https://gist.github.com/marvell/7c812736565928e602c4
# 6. Some packages will create config file under /home by default, so we need to make sure it's writable
# 7. Change the permission of /tmp, so that others can run command on it
RUN apt-get update \
    && apt-get install build-essential wget -y \
    && mkdir -p /tmp/ \
    && mkdir -p /tmp/code-server \
    && wget --no-check-certificate -O /tmp/code-server/code-server-4.19.0-linux-${TARGETARCH}.tar.gz https://github.com/coder/code-server/releases/download/v4.19.0/code-server-4.19.0-linux-${TARGETARCH}.tar.gz \
    && tar -xzf /tmp/code-server/code-server-4.19.0-linux-${TARGETARCH}.tar.gz -C /tmp/code-server/ \
    && wget --no-check-certificate https://open-vsx.org/api/ms-python/python/2023.20.0/file/ms-python.python-2023.20.0.vsix -P /tmp/code-server \
    && wget --no-check-certificate https://open-vsx.org/api/ms-toolsai/jupyter/2023.9.100/file/ms-toolsai.jupyter-2023.9.100.vsix -P /tmp/code-server \
    && pip install --no-cache-dir -U flytekitplugins-flyteinteractive==$VERSION flytekit==$VERSION \
    && apt-get clean autoclean \
    && apt-get autoremove --yes \
    && rm -rf /var/lib/{apt,dpkg,cache,log}/ \
    && useradd -u 1000 flytekit \
    && chown -R flytekit:flytekit /tmp/code-server \
    && chown flytekit: /root \
    && chown flytekit: /home \
    && :

# Set the environment variable for code-server
ENV PATH="/tmp/code-server/code-server-4.19.0-linux-${TARGETARCH}/bin:${PATH}"

USER flytekit

# Install extensions using code-server
# Execution is performed here as code-server configuration depends on the USER setting
# If we install it as ROOT, the config will be stored in /root/.config/code-server/config.yaml
# Now, the config of code-server will be stored in /home/flytekit/.config/code-server/config.yaml
RUN code-server --install-extension /tmp/code-server/ms-python.python-2023.20.0.vsix \
    && code-server --install-extension /tmp/code-server/ms-toolsai.jupyter-2023.9.100.vsix
