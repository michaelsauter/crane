#!/usr/bin/env bash

set -e

version=$1

if [ -z "$version" ]; then
  echo "No version passed! Example usage: ./release.sh 1.0.0"
  exit 1
fi

go_path=$(cd ../../../../; pwd)
docker_run="docker run --rm -it -v $go_path:/go -w /go/src/github.com/michaelsauter/crane michaelsauter/golang:1.7"

echo "Running tests..."
$docker_run make test

echo "Update version..."
grepped_version=$(grep -o "v[0-9]*\.[0-9]*\.[0-9]*" crane/cli.go)
old_version=${grepped_version:1}
sed -i.bak 's/fmt\.Println("v'$old_version'")/fmt.Println("v'$version'")/' crane/cli.go
rm crane/cli.go.bak
sed -i.bak 's/VERSION="'$old_version'"/VERSION="'$version'"/' download.sh
rm download.sh.bak
sed -i.bak 's/'$old_version'/'$version'/' README.md
rm README.md.bak

echo "Mark version as released in changelog..."
today=$(date +'%Y-%m-%d')
sed -i.bak "s/Unreleased/Unreleased\n\n## $version ($today)/" CHANGELOG.md
rm CHANGELOG.md.bak

echo "Update contributors..."
git contributors | awk '{for (i=2; i<NF; i++) printf $i " "; print $NF}' > CONTRIBUTORS

echo "Build binaries..."
$docker_run make build-linux-amd64
$docker_run make build-darwin-amd64
$docker_run make build-windows-amd64

echo "Update repository..."
git add crane/cli.go download.sh README.md CHANGELOG.md CONTRIBUTORS
git commit -m "Bump version to ${version}"
git tag --sign --message="v$version" "v$version"
git tag --sign --message="latest" --force latest


echo "v$version tagged."
echo "Now, run 'git push origin master && git push --tags --force' and publish the release on GitHub."
