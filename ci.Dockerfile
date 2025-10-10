# Flyte CI/Development Docker Image
# This image provides a consistent environment for both CI and local development
# to eliminate "works on my machine" issues.

FROM ubuntu:24.04

# Prevent interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive

# Set versions for all tools
ARG GO_VERSION=1.24.6
ARG PYTHON_VERSION=3.12.9
ARG NODE_VERSION=20.18.3
ARG RUST_VERSION=1.84.0
ARG UV_VERSION=0.8.4
ARG BUF_VERSION=1.58.0
ARG MOCKERY_VERSION=2.53.5

# Install system dependencies
RUN apt-get update && apt-get install -y \
    # Basic tools
    curl \
    wget \
    git \
    make \
    build-essential \
    ca-certificates \
    gnupg \
    lsb-release \
    unzip \
    # Python build dependencies
    libssl-dev \
    zlib1g-dev \
    libbz2-dev \
    libreadline-dev \
    libsqlite3-dev \
    libncursesw5-dev \
    xz-utils \
    tk-dev \
    libxml2-dev \
    libxmlsec1-dev \
    libffi-dev \
    liblzma-dev \
    # Additional utilities
    jq \
    && rm -rf /var/lib/apt/lists/*

# Install Go
RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" | tar -C /usr/local -xz
ENV PATH="/usr/local/go/bin:/root/go/bin:${PATH}"
ENV GOPATH="/root/go"

# Install Python using pyenv for better version management
RUN git clone --depth=1 https://github.com/pyenv/pyenv.git /root/.pyenv
ENV PYENV_ROOT="/root/.pyenv"
ENV PATH="${PYENV_ROOT}/shims:${PYENV_ROOT}/bin:${PATH}"
RUN pyenv install ${PYTHON_VERSION} && \
    pyenv global ${PYTHON_VERSION} && \
    pyenv rehash

# Install uv (fast Python package manager)
RUN curl -LsSf "https://astral.sh/uv/${UV_VERSION}/install.sh" | sh
ENV PATH="/root/.local/bin:${PATH}"

# Install Node.js
RUN curl -fsSL "https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.xz" | tar -C /usr/local --strip-components=1 -xJ
RUN npm install -g npm@latest

# Install Rust
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain ${RUST_VERSION}
ENV PATH="/root/.cargo/bin:${PATH}"

# Install Buf CLI
RUN curl -fsSL "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-Linux-x86_64.tar.gz" | \
    tar -xzC /usr/local --strip-components 1

# Install Go tools
RUN go install "github.com/vektra/mockery/v2@v${MOCKERY_VERSION}"

# Set working directory
WORKDIR /workspace

# Verify installations
RUN echo "=== Tool Versions ===" && \
    echo "Go: $(go version)" && \
    echo "Python: $(python --version)" && \
    echo "uv: $(uv --version)" && \
    echo "Node: $(node --version)" && \
    echo "npm: $(npm --version)" && \
    echo "Rust: $(rustc --version)" && \
    echo "Cargo: $(cargo --version)" && \
    echo "Buf: $(buf --version)" && \
    echo "Mockery: $(mockery --version)"

# Set default command
CMD ["/bin/bash"]
