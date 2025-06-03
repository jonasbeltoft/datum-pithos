#!/usr/bin/env bash
set -e

# Determine OS for cross-compilation
UNAME_OUT="$(uname -s)"
case "${UNAME_OUT}" in
    Linux*)     OS=linux;;
    Darwin*)    OS=darwin;;
    CYGWIN*|MINGW*|MSYS*) OS=windows;;
    *)          echo "Unsupported OS: ${UNAME_OUT}"; exit 1;;
esac

ARCH_RAW="$(uname -m)"
case "${ARCH_RAW}" in
    x86_64) ARCH=x64;;
    arm64|aarch64) ARCH=arm64;;
    *) echo "Unsupported architecture: ${ARCH_RAW}"; exit 1;;
esac

# Set .NET Runtime ID
case "${OS}" in
    windows) DOTNET_RUNTIME="win-${ARCH}"; EXT=".exe";;
    darwin)  DOTNET_RUNTIME="osx-${ARCH}"; EXT="";;
    linux)   DOTNET_RUNTIME="linux-${ARCH}"; EXT="";;
esac


# Create the build directory
mkdir -p build

# Build the backend Go binary
echo "Building backend for $OS..."
pushd backend > /dev/null
GOOS="$OS" GOARCH="$GOARCH" go build -o "../build/backend${EXT}" .
popd > /dev/null

# Build the frontend .NET binary
echo "Publishing frontend for $DOTNET_RUNTIME..."
dotnet publish ./frontend/BlazorApp/ \
    -c Release \
    -r "$DOTNET_RUNTIME" \
    --self-contained true \
    /p:PublishSingleFile=true \
    -o "./build/frontend"

# Write README.txt
echo "Creating README.txt in ./build..."
cat > ./build/README.txt <<EOL
This directory contains the built binaries for the project.
The backend binary is named backend${EXT} and the frontend binary is in the 'frontend' directory as BlazorApp${EXT}.
They must be run from their own directory, separately.
Backend takes --port=<port> as an argument to specify the port to run on.
EOL

echo "Build completed successfully for $OS-$ARCH!"
