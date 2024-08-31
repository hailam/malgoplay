# Variables
ANDROID_SDK_DIR=$(HOME)/Library/Android/sdk
ANDROID_NDK_VERSION=21.1.6352462
ANDROID_PLATFORM_VERSION=android-34
ANDROID_BUILD_TOOLS_VERSION=34.0.0

IOS_SDK_VERSION=14.5  # Example iOS SDK version, adjust as necessary

# Paths
ANDROID_NDK_DIR=$(ANDROID_SDK_DIR)/ndk/$(ANDROID_NDK_VERSION)
ANDROID_NDK_HOME=$(ANDROID_NDK_DIR)
GOMOBILE=gomobile

# Shell scripts
SCRIPTS_DIR=./scripts
SETUP_ANDROID_ENV_SCRIPT=$(SCRIPTS_DIR)/setup_android_env.sh
SETUP_NDK_SCRIPT=$(SCRIPTS_DIR)/setup_ndk.sh
BUILD_ANDROID_SCRIPT=$(SCRIPTS_DIR)/build_android.sh
BUILD_IOS_SCRIPT=$(SCRIPTS_DIR)/build_ios.sh

# Targets
.PHONY: all setup_android_env android ios cli clean

all: setup_android_env cli android ios

setup_android_env:
	@$(SETUP_ANDROID_ENV_SCRIPT) $(ANDROID_SDK_DIR) $(ANDROID_PLATFORM_VERSION) $(ANDROID_BUILD_TOOLS_VERSION) $(ANDROID_NDK_VERSION)

android: 
	@$(BUILD_ANDROID_SCRIPT) $(ANDROID_NDK_HOME) $(GOMOBILE)

ios:
	@$(BUILD_IOS_SCRIPT) $(GOMOBILE) $(IOS_SDK_VERSION)

cli:
	@echo "Building CLI tool..."
	@go build -o bin/malgoplay cmd/cli/main.go

clean:
	@echo "Cleaning up..."
	rm -rf bin/ android/app/libs/malgoplay.aar ios/malgoplayapp/Malgoplay.framework
