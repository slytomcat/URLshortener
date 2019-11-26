#!/usr/bin/env bash
#

# check that this is the release
export TAG=$(git describe --abbrev=0 --tags);
echo "Last release is $TAG"
CTAG=$(git describe --tags)
echo "Current tag is $CTAG"  
if [[ $TAG != $CTAG ]]; then 
  # exit if it is not release
  exit 0
fi

# get upload assets utility
wget https://gist.githubusercontent.com/stefanbuck/ce788fee19ab6eb0b4447a85fc99f447/raw/dbadd7d310ce8446de89c4ffdf1db0b400d0f6c3/upload-github-release-asset.sh
chmod a+x upload-github-release-asset.sh

# set environment
export OWNER=$CIRCLE_PROJECT_USERNAME
export REPO=$CIRCLE_PROJECT_REPONAME

echo "Building for amd64 architrecture"
go build .
mv yandex-disk-simulator yandex-disk-simulator-linux-amd64

echo "Uploading amd64"
./upload-github-release-asset.sh github_api_token=$GHAPITOKEN owner=$OWNER repo=$REPO tag="$TAG" filename=yandex-disk-simulator-linux-amd64

echo "Building for 386 architrecture"
export GOARCH=386
go build .
mv yandex-disk-simulator yandex-disk-simulator-linux-386

echo "Uploading 386"
./upload-github-release-asset.sh github_api_token=$GHAPITOKEN owner=$OWNER repo=$REPO tag="$TAG" filename=yandex-disk-simulator-linux-386


