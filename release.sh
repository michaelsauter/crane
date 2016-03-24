#!/usr/bin/env bash

set -e

version=$1

if [ -z "$version" ]; then
  echo "No version passed! Example usage: ./release.sh 1.0.0"
  exit 1
fi

go_path=$(cd ../../../../; pwd)
docker_options="--rm -it -v $go_path:/go -w /go/src/github.com/michaelsauter/crane"
docker_image="michaelsauter/golang:1.6"

echo "Running tests..."
docker run $docker_options $docker_image make test

echo "Update version..."
sed -i.bak 's/fmt\.Println("v[0-9]\{1,2\}\.[0-9]\{1,2\}\.[0-9]\{1,2\}")/fmt.Println("v'$version'")/' crane/cli.go
rm crane/cli.go.bak
sed -i.bak 's/VERSION="[0-9]\{1,2\}\.[0-9]\{1,2\}\.[0-9]\{1,2\}"/VERSION="'$version'"/' download.sh
rm download.sh.bak
sed -i.bak 's/[0-9]\{1,2\}\.[0-9]\{1,2\}\.[0-9]\{1,2\}/'$version'/' README.md
rm README.md.bak

echo "Mark version as released in changelog..."
today=$(date +'%Y-%m-%d')
sed -i.bak 's/Unreleased/Unreleased\n\n## '$version' ('$today')/' CHANGELOG.md
rm CHANGELOG.md.bak

echo "Update contributors..."
git contributors | awk '{for (i=2; i<NF; i++) printf $i " "; print $NF}' > CONTRIBUTORS

echo "Build binaries..."
docker run $docker_options -e GOOS=linux -e GOARCH=386 $docker_image go build -o crane_linux_386 -v github.com/michaelsauter/crane
docker run $docker_options -e GOOS=linux -e GOARCH=amd64 $docker_image go build -o crane_linux_amd64 -v github.com/michaelsauter/crane
docker run $docker_options -e GOOS=darwin -e GOARCH=amd64 $docker_image go build -o crane_darwin_amd64 -v github.com/michaelsauter/crane
docker run $docker_options -e GOOS=windows -e GOARCH=amd64 $docker_image go build -o crane_windows_amd64.exe -v github.com/michaelsauter/crane

echo "Update repository..."
git add crane/cli.go download.sh README.md CHANGELOG.md CONTRIBUTORS
git commit -m "Bump version to ${version}"
git tag --sign --message="v$version" "v$version"

echo "v$version tagged."
echo "Now, run 'git push origin master && git push --tags' and publish the release on GitHub."
