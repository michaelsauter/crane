#!/usr/bin/env bash

# Set version to latest unless set by user
if [ -z "$VERSION" ]; then
  VERSION="3.3.2"
fi

echo "Downloading version ${VERSION}..."

# OS information (contains e.g. darwin x86_64)
UNAME=`uname -a | awk '{print tolower($0)}'`
if [[ ($UNAME == *"mac os x"*) || ($UNAME == *darwin*) ]]
then
  PLATFORM="darwin"
else
  PLATFORM="linux"
fi
if [[ ($UNAME == *x86_64*) || ($UNAME == *amd64*) ]]
then
  ARCH="amd64"
else
  echo "Currently, there are no 32bit binaries provided."
  echo "You will need to build binaries yourself."
  exit 1
fi

# Download binary
curl -L -o crane "https://github.com/michaelsauter/crane/releases/download/v${VERSION}/crane_${PLATFORM}_${ARCH}"

# Make binary executable
chmod +x crane

echo "Done."
