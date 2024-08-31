#!/bin/bash

# Arguments
ANDROID_NDK_HOME=$1
GOMOBILE=$2

echo "Building Android bindings..."

# Ensure the output directory exists
OUTPUT_DIR="platform/android/fsg-aar"
mkdir -p $OUTPUT_DIR

# Export the necessary environment variable and build the bindings

# If the AAR file exists delete it
if [ -f $OUTPUT_DIR/fsg.aar ]; then
    rm $OUTPUT_DIR/fsg.aar
fi

#if [ ! -f $OUTPUT_DIR/malgoplay.aar ]; then
    export ANDROID_NDK_HOME=$ANDROID_NDK_HOME
    $GOMOBILE bind -target=android -o $OUTPUT_DIR/fsg.aar -v -tags release ./cmd/mobile

    if [ $? -ne 0 ]; then
        echo "Error building Android bindings"
        exit 1
    fi

    echo "Android bindings built successfully."
#fi

# Change to the android directory
cd platform/android

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
