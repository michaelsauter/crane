#!/usr/bin/env sh

# Set version to latest unless set by user
if [ -z "$VERSION" ]; then
  VERSION="3.0.0"
fi

echo "Downloading version ${VERSION}..."

# OS information (contains e.g. darwin x86_64)
UNAME=`uname -a | awk '{print tolower($0)}'`
if [ "$(expr index "$UNAME" "mac os x")" -gt 0 -o "$(expr index "$UNAME" "darwin")" ]
then
  PLATFORM="darwin"
else
  PLATFORM="linux"
fi
if [ "$(expr index "$UNAME" "x86_64")" -gt 0 -o "$(expr index "$UNAME" "amd64")" ]
then
  ARCH="amd64"
else
  echo "Currently, there are no 32bit binaries provided."
  echo "You will need to go get / go install github.com/michaelsauter/crane."
  exit 1
fi

# Download binary
curl -L -o crane "https://github.com/michaelsauter/crane/releases/download/v${VERSION}/crane_${PLATFORM}_${ARCH}"

# Make binary executable
chmod +x crane

echo "Done."
