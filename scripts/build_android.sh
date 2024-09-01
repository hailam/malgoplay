#!/bin/bash

# Arguments
ANDROID_NDK_HOME=$1
GOMOBILE=$2

# Check if required arguments are provided
if [ -z "$ANDROID_NDK_HOME" ] || [ -z "$GOMOBILE" ]; then
    echo "Usage: $0 <ANDROID_NDK_HOME> <GOMOBILE>"
    exit 1
fi

echo "Building Android bindings..."

# Ensure the output directory exists
OUTPUT_DIR="platform/android/fsg-aar"
mkdir -p $OUTPUT_DIR

# Export the necessary environment variable and build the bindings
export ANDROID_NDK_HOME="$ANDROID_NDK_HOME"

# If the AAR file exists delete it
if [ -f $OUTPUT_DIR/fsg.aar ]; then
    rm $OUTPUT_DIR/fsg.aar
fi

# Build the bindings
$GOMOBILE bind -target=android -o $OUTPUT_DIR/fsg.aar -v -tags release ./cmd/mobile

if [ $? -ne 0 ]; then
    echo "Error building Android bindings"
    exit 1
fi

echo "Android bindings built successfully."

# Change to the android directory
cd platform/android || { echo "Failed to change directory to platform/android"; exit 1; }

# Ensure gradlew is executable
chmod +x gradlew

# Now build the Android app using the included gradlew script
echo "Building the Android application..."
./gradlew assembleDebug

if [ $? -ne 0 ]; then
    echo "Error building the Android application"
    exit 1
fi

echo "Android application built successfully."

# Return to the original directory
cd - || exit
