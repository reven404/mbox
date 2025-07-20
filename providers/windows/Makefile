# Makefile for Windows Cloud Files API

APP_NAME = mdriver
LIB_NAME = cloudfiles_bridge
BUILD_DIR = build

# Windows SDK and build tools
WINDOWS_SDK = "C:/Program Files (x86)/Windows Kits/10"
MSVC_PATH = "C:/Program Files/Microsoft Visual Studio/2022/Community/VC/Tools/MSVC"

# Compiler flags
CXXFLAGS = /std:c++17 /EHsc /MD /O2
INCLUDES = /I$(WINDOWS_SDK)/Include/10.0.22621.0/um \
           /I$(WINDOWS_SDK)/Include/10.0.22621.0/shared \
           /I$(WINDOWS_SDK)/Include/10.0.22621.0/winrt
LIBS = ole32.lib shell32.lib advapi32.lib version.lib

.PHONY: all clean bridge help test

all: bridge

help:
	@echo "Available targets:"
	@echo "  bridge     - Build the Cloud Files API bridge library"
	@echo "  clean      - Clean build artifacts"  
	@echo "  test       - Test Cloud Files API availability"
	@echo "  help       - Show this help message"

# Build the Cloud Files bridge library
bridge:
	@echo "Building Windows Cloud Files API bridge..."
	@if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
	
	@echo "Compiling C++ bridge..."
	cl $(CXXFLAGS) $(INCLUDES) /c cloudfiles_bridge.cpp /Fo:$(BUILD_DIR)/cloudfiles_bridge.obj
	
	@echo "Creating static library..."
	lib /OUT:$(BUILD_DIR)/$(LIB_NAME).lib $(BUILD_DIR)/cloudfiles_bridge.obj
	
	@echo "Bridge library built: $(BUILD_DIR)/$(LIB_NAME).lib"

# Test Cloud Files API availability
test:
	@echo "Testing Cloud Files API availability..."
	@echo "This requires Windows 10 version 1809 (build 17763) or later"
	
	@powershell -Command "$$version = [System.Environment]::OSVersion.Version; \
		if ($$version.Major -eq 10 -and $$version.Build -ge 17763) { \
			Write-Host 'Cloud Files API is supported on this system' -ForegroundColor Green; \
			exit 0 \
		} else { \
			Write-Host 'Cloud Files API is NOT supported on this system' -ForegroundColor Red; \
			Write-Host 'Requires Windows 10 1809+ (build 17763+)' -ForegroundColor Yellow; \
			exit 1 \
		}"

# Build with Go (requires the bridge library)
go-build: bridge
	@echo "Building Go application with Cloud Files support..."
	set CGO_ENABLED=1
	set GOOS=windows
	set GOARCH=amd64
	go build -tags cloudfiles -ldflags="-L./providers/windows/build -lcloudfiles_bridge" ../..

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@if exist $(BUILD_DIR) rmdir /s /q $(BUILD_DIR)
	@echo "Clean complete."

# Cross-compilation note
cross-compile-note:
	@echo "Note: Cross-compilation for Windows Cloud Files API:"
	@echo "1. This requires Windows SDK and MSVC compiler"
	@echo "2. Cloud Files API is only available on Windows"
	@echo "3. Use 'go build -tags cloudfiles' to enable"
	@echo "4. Requires CGO_ENABLED=1 for C++ bridge"

# Development setup
dev-setup:
	@echo "Setting up development environment for Cloud Files API..."
	@echo "Required components:"
	@echo "- Windows 10 SDK (latest version)"
	@echo "- Visual Studio 2019/2022 with C++ tools"
	@echo "- Go with CGO support"
	@echo "- Windows 10 version 1809+ for testing"