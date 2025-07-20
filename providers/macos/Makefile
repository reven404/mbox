# Makefile for building the File Provider Extension

APP_NAME = mdriver
EXTENSION_NAME = FileProviderExtension
BUNDLE_ID_BASE = com.mdriver
EXTENSION_BUNDLE_ID = $(BUNDLE_ID_BASE).$(EXTENSION_NAME)

# Build directories
BUILD_DIR = build
EXTENSION_DIR = $(BUILD_DIR)/$(EXTENSION_NAME).appex

# Source files
SWIFT_SOURCES = FileProviderExtension/*.swift
OBJC_SOURCES = *.m
HEADER_FILES = *.h

# Build configuration
DEPLOYMENT_TARGET = 10.15
SWIFT_VERSION = 5.0

# Xcode build settings
XCODE_BUILD_FLAGS = -configuration Release \
	-derivedDataPath $(BUILD_DIR) \
	DEPLOYMENT_TARGET=$(DEPLOYMENT_TARGET) \
	SWIFT_VERSION=$(SWIFT_VERSION)

.PHONY: all clean extension install-extension remove-extension help

all: extension

help:
	@echo "Available targets:"
	@echo "  extension        - Build the File Provider extension"
	@echo "  install-extension - Install the extension for development"
	@echo "  remove-extension - Remove the installed extension"
	@echo "  clean           - Clean build artifacts"
	@echo "  help            - Show this help message"

extension:
	@echo "Building File Provider Extension..."
	@mkdir -p $(BUILD_DIR)
	
	# Compile Swift sources
	@echo "Compiling Swift sources..."
	xcrun swiftc $(SWIFT_SOURCES) \
		-target x86_64-apple-macosx$(DEPLOYMENT_TARGET) \
		-module-name $(EXTENSION_NAME) \
		-emit-module -emit-module-path $(BUILD_DIR)/$(EXTENSION_NAME).swiftmodule \
		-emit-object -o $(BUILD_DIR)/swift_objects.o \
		-framework Foundation -framework FileProvider \
		-swift-version $(SWIFT_VERSION)
	
	# Compile Objective-C sources  
	@echo "Compiling Objective-C sources..."
	xcrun clang $(OBJC_SOURCES) \
		-target x86_64-apple-macosx$(DEPLOYMENT_TARGET) \
		-framework Foundation -framework FileProvider \
		-c -o $(BUILD_DIR)/objc_objects.o
	
	# Create extension bundle directory
	@mkdir -p $(EXTENSION_DIR)/Contents/MacOS
	@mkdir -p $(EXTENSION_DIR)/Contents/Resources
	
	# Link the extension executable
	@echo "Linking extension executable..."
	xcrun clang $(BUILD_DIR)/swift_objects.o $(BUILD_DIR)/objc_objects.o \
		-target x86_64-apple-macosx$(DEPLOYMENT_TARGET) \
		-framework Foundation -framework FileProvider \
		-o $(EXTENSION_DIR)/Contents/MacOS/$(EXTENSION_NAME)
	
	# Copy Info.plist
	cp FileProviderExtension/Info.plist $(EXTENSION_DIR)/Contents/Info.plist
	
	# Update bundle identifier in Info.plist
	/usr/libexec/PlistBuddy -c "Set :CFBundleIdentifier $(EXTENSION_BUNDLE_ID)" \
		$(EXTENSION_DIR)/Contents/Info.plist
	
	# Copy Swift module
	cp $(BUILD_DIR)/$(EXTENSION_NAME).swiftmodule $(EXTENSION_DIR)/Contents/Resources/
	
	@echo "Extension built successfully at $(EXTENSION_DIR)"

install-extension: extension
	@echo "Installing File Provider Extension for development..."
	
	# Copy extension to system location (requires admin privileges)
	@echo "Note: This may require administrator privileges"
	sudo cp -r $(EXTENSION_DIR) "/Library/Application Support/FileProvider/Extensions/"
	
	# Restart File Provider daemon
	sudo launchctl unload /System/Library/LaunchDaemons/com.apple.FileProvider.plist 2>/dev/null || true
	sudo launchctl load /System/Library/LaunchDaemons/com.apple.FileProvider.plist
	
	@echo "Extension installed. You may need to enable it in System Preferences."

remove-extension:
	@echo "Removing installed File Provider Extension..."
	sudo rm -rf "/Library/Application Support/FileProvider/Extensions/$(EXTENSION_NAME).appex"
	
	# Restart File Provider daemon
	sudo launchctl unload /System/Library/LaunchDaemons/com.apple.FileProvider.plist 2>/dev/null || true
	sudo launchctl load /System/Library/LaunchDaemons/com.apple.FileProvider.plist
	
	@echo "Extension removed."

clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	@echo "Clean complete."

# Development helpers
debug-info:
	@echo "Build Configuration:"
	@echo "  APP_NAME: $(APP_NAME)"
	@echo "  EXTENSION_NAME: $(EXTENSION_NAME)"
	@echo "  EXTENSION_BUNDLE_ID: $(EXTENSION_BUNDLE_ID)"
	@echo "  BUILD_DIR: $(BUILD_DIR)"
	@echo "  DEPLOYMENT_TARGET: $(DEPLOYMENT_TARGET)"
	@echo "  SWIFT_VERSION: $(SWIFT_VERSION)"

# Check if we're on macOS
check-platform:
	@if [ "$(shell uname)" != "Darwin" ]; then \
		echo "Error: This Makefile is for macOS only"; \
		exit 1; \
	fi

# Validate Xcode tools
check-xcode:
	@if ! command -v xcrun >/dev/null 2>&1; then \
		echo "Error: Xcode command line tools not found"; \
		echo "Please install with: xcode-select --install"; \
		exit 1; \
	fi