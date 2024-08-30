#!/bin/bash

# Arguments
GOMOBILE=$1
IOS_SDK_VERSION=$2

echo "Building iOS framework..."

# Build the iOS framework using gomobile
$GOMOBILE bind -target=ios -o ios/malgoplayapp/Malgoplay.framework ./cmd/ios

if [ $? -ne 0 ]; then
  echo "Error building iOS framework"
  exit 1
fi

echo "iOS framework built successfully."
