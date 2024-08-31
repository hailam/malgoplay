# MalgoPlay

## Table of Contents

1. [Introduction](#introduction)
2. [Command-line Arguments](#command-line-arguments)
3. [Build Instructions](#build-instructions)
   - [Android](#android)
   - [iOS](#ios)
   - [CLI Tool](#cli-tool)
4. [Examples](#examples)

## Introduction

MalgoPlay is an experimental project that aims to create a cross-platform application using shared Go code. The primary goal is to develop a robust core logic in Go that can be reused across multiple platforms, including Android, iOS, web, and desktop applications using Qt. The platform-specific components, such as the user interface, are implemented separately for each platform while leveraging the shared logic.

The sine wave generation and analysis are secondary aspects of this project, serving as the application’s current functionality to demonstrate the shared logic. MalgoPlay can produce a fixed frequency or sweep through a range of frequencies, providing real-time feedback on the match between the generated and detected frequencies.

Versions for iOS, web, and Qt currently in development.

## Command-line Arguments

The CLI tool provides several command-line options for configuring the frequency sweep generator:

- `-f`, `--frequency`: Maximum frequency of the sine wave in Hz (default: 1000)
- `-m`, `--min-frequency`: Minimum frequency to start sweeping from in Hz (default: 0, which means no sweeping)
- `-a`, `--amplitude`: Amplitude of the sine wave (default: 0.5)
- `-r`, `--sample-rate`: Sample rate in Hz (default: 48000)
- `-s`, `--sweep-rate`: Frequency change rate in Hz per second when sweeping (default: 1.0)
- `--min`, `-m`: Minimum frequency (default: 220)
- `--max`, `-M`: Maximum frequency (default: 880)
- `--rate`, `-r`: Sample rate in Hz (default: 44100)
- `--channels`, `-c`: Number of channels (default: 2)
- `--duration`, `-d`: Duration in seconds (default: 10, 0 for indefinite playback)
- `--sweep`, `-s`: Sweep rate in Hz (default: 1.0)
- `--mode`, `-o`: Sweep mode (linear, sine, triangle, exponential, logarithmic, square, sawtooth, random)

## Build Instructions

### Android

To build the Android component of MalgoPlay, you need to have the Android SDK and NDK installed. The `Makefile` automates the setup and build process.

**Prerequisites:**

- Android SDK installed in `$(HOME)/Library/Android/sdk`
- Android NDK version 21.1.6352462
- Gradle (required for building the Android project)

**Build Steps:**

1. **Setup Android Environment:**

   ```sh
   make setup_android_env
   ```

2. **Build for Android:**
   ```sh
   make android
   ```

This will compile the Go code and package it into an Android `.aar` library.

**Note:** The `cmd/android` directory contains the old duplex library implementation, which is now deprecated. The new code for both Android and iOS is found in the `mobile` directory.

### iOS

To build the iOS component of MalgoPlay, you need to have Xcode and the iOS SDK installed.

**Prerequisites:**

- iOS SDK version 14.5 (adjust as necessary)
- GoMobile for iOS bindings

**Build Steps:**

1. **Build for iOS:**
   ```sh
   make ios
   ```

This will generate the necessary iOS frameworks that can be integrated into your iOS project.

### CLI Tool

To build the CLI tool:

```sh
make cli
```

The CLI tool will be available in the `bin` directory.

## Examples

1. Generate a fixed 440 Hz tone:

   ```sh
   ./bin/malgoplay -f 440
   ```

2. Sweep from 1000 Hz to 600 Hz, changing by 2 Hz per second:

   ```sh
   ./bin/malgoplay -f 1000 -m 100 -s 10
   ```
