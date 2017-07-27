#!/usr/bin/env bash

set -e

version=$1

if [ -z "$version" ]; then
  echo "No version passed! Example usage: ./release.sh 1.0.0"
  exit 1
fi

echo "Running tests..."
crane run crane make test

echo "Update version..."
grepped_version=$(grep -o "v[0-9]*\.[0-9]*\.[0-9]*" crane/version_basic.go)
old_version=${grepped_version:1}
sed -i.bak 's/fmt\.Println("v'$old_version'")/fmt.Println("v'$version'")/' crane/version_basic.go
sed -i.bak 's/fmt\.Println("v'$old_version' (PRO)")/fmt.Println("v'$version' (PRO)")/' crane/version_pro.go
rm crane/version_basic.go.bak
rm crane/version_pro.go.bak
sed -i.bak 's/VERSION="'$old_version'"/VERSION="'$version'"/' download.sh
rm download.sh.bak

echo "Mark version as released in changelog..."
today=$(date +'%Y-%m-%d')
sed -i.bak "s/Unreleased/Unreleased\n\n## $version ($today)/" CHANGELOG.md
rm CHANGELOG.md.bak

echo "Update contributors..."
git contributors | awk '{for (i=2; i<NF; i++) printf $i " "; print $NF}' > CONTRIBUTORS

echo "Build binaries..."
crane run crane make build

echo "Update repository..."
git add crane/cli.go download.sh README.md CHANGELOG.md CONTRIBUTORS
git commit -m "Bump version to ${version}"
git tag --sign --message="v$version" "v$version"
git tag --sign --message="latest" --force latest


echo "v$version tagged."
echo "Now, run 'git push origin master && git push --tags --force' and publish the release on GitHub."
