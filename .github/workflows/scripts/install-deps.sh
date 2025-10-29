#!/bin/bash
set -euo pipefail

# Install dependencies for Go development and testing
# This script is designed to run in a clean Ubuntu container

echo "==> Installing system dependencies..."
apt-get update -qq
apt-get install -y -qq \
    wget \
    git \
    ca-certificates \
    > /dev/null

echo "==> Installing Go ${GO_VERSION}..."
ARCH=$(dpkg --print-architecture)
GO_TAR="go${GO_VERSION}.linux-${ARCH}.tar.gz"
wget -q "https://go.dev/dl/${GO_TAR}"
rm -rf /usr/local/go
tar -C /usr/local -xzf "${GO_TAR}"
rm "${GO_TAR}"

# Add Go to PATH for this script
export PATH="/usr/local/go/bin:${PATH}"

# Verify Go installation
go version

echo "==> Installing staticcheck..."
go install honnef.co/go/tools/cmd/staticcheck@latest

# Verify staticcheck installation
GOPATH=$(go env GOPATH)
"${GOPATH}/bin/staticcheck" -version

echo "==> Dependencies installed successfully!"
