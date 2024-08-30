#!/bin/bash

# Arguments
ANDROID_NDK_HOME=$1
GOMOBILE=$2

echo "Building Android bindings..."

# Ensure the output directory exists
OUTPUT_DIR="android/malgoplay-aar"
mkdir -p $OUTPUT_DIR

# Export the necessary environment variable and build the bindings

# If the AAR file exists delete it
if [ -f $OUTPUT_DIR/malgoplay.aar ]; then
    rm $OUTPUT_DIR/malgoplay.aar
fi

#if [ ! -f $OUTPUT_DIR/malgoplay.aar ]; then
    export ANDROID_NDK_HOME=$ANDROID_NDK_HOME
    $GOMOBILE bind -target=android -o $OUTPUT_DIR/malgoplay.aar ./cmd/android

    if [ $? -ne 0 ]; then
        echo "Error building Android bindings"
        exit 1
    fi

    echo "Android bindings built successfully."
#fi

# Change to the android directory
cd android

# Now build the Android app using the included gradlew script
echo "Building the Android application..."

./gradlew assembleDebug

if [ $? -ne 0 ]; then
    echo "Error building the Android application"
    exit 1
fi

echo "Android application built successfully."

# Return to the original directory
cd -
