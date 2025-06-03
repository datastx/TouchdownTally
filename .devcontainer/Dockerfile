FROM mcr.microsoft.com/vscode/devcontainers/go:1.23-bookworm

# Install additional OS packages
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
        postgresql-client \
        redis-tools \
        curl \
        wget \
        git \
        make \
        build-essential \
        python3-pip \
        python3-venv \
        jq \
        vim \
        htop \
    && apt-get autoremove -y && apt-get clean -y && rm -rf /var/lib/apt/lists/*

# Install essential Go tools
RUN go install golang.org/x/tools/gopls@latest \
    && go install golang.org/x/tools/cmd/goimports@latest \
    && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest

# Install useful Python packages for data analysis/debugging
RUN pip3 install --no-cache-dir requests pandas psycopg2-binary

# Create workspace directory
WORKDIR /workspace

# Set the default shell to bash
SHELL ["/bin/bash", "-c"]
