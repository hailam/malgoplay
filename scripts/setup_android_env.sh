#!/bin/bash

# Arguments
SDK_DIR=$1
PLATFORM_VERSION=$2
BUILD_TOOLS_VERSION=$3
NDK_VERSION=$4

# Paths
SDKMANAGER="$SDK_DIR/cmdline-tools/latest/bin/sdkmanager"

echo "Ensuring necessary Android SDK components are installed..."

# Install necessary Android SDK components
$SDKMANAGER --sdk_root=$SDK_DIR "platform-tools" "platforms;$PLATFORM_VERSION" "build-tools;$BUILD_TOOLS_VERSION" "ndk;$NDK_VERSION"

if [ $? -ne 0 ]; then
  echo "Error installing SDK components"
  exit 1
fi

echo "Android SDK and NDK setup complete."
