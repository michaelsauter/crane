#!/usr/bin/env bash

set -e

version=$1

if [ -z "$version" ]; then
  echo "No version passed! Example usage: ./release.sh 1.0.0"
  exit 1
fi

echo "Running tests..."
go test ./...

echo "Update version..."
sed -i.bak 's/fmt\.Println("v[0-9]\.[0-9]\.[0-9]")/fmt.Println("v'$version'")/' crane/cmd.go
rm crane/cmd.go.bak
sed -i.bak 's/VERSION="[0-9]\.[0-9]\.[0-9]"/VERSION="'$version'"/' download.sh
rm download.sh.bak

echo "Update contributors..."
git contributors | awk '{for (i=2; i<NF; i++) printf $i " "; print $NF}' > CONTRIBUTORS

echo "Build binary..."
../../../../bin/gox -osarch="darwin/amd64" -osarch="linux/amd64"

echo "Update repository..."
git add crane/cmd.go download.sh CONTRIBUTORS
git commit -m "Bump version to ${version}"
git tag "v$version"

echo "v$version tagged."
echo "Now, run 'git push origin master && git push --tags' and publish the release on GitHub."
