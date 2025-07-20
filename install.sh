#!/bin/bash

# mdriver installation script
# This script downloads and installs the latest release of mdriver

set -e

# Configuration
GITHUB_USER="${GITHUB_USER:-username}"
GITHUB_REPO="${GITHUB_REPO:-mdriver}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${VERSION:-latest}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case $OS in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        *)
            log_error "Unsupported operating system: $OS"
            exit 1
            ;;
    esac
    
    case $ARCH in
        x86_64|amd64)
            ARCH="x86_64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l)
            ARCH="armv7"
            ;;
        *)
            log_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    log_info "Detected platform: $OS/$ARCH"
}

# Get latest release version
get_latest_version() {
    if [ "$VERSION" = "latest" ]; then
        log_info "Fetching latest release version..."
        VERSION=$(curl -s "https://api.github.com/repos/$GITHUB_USER/$GITHUB_REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
        if [ -z "$VERSION" ]; then
            log_error "Failed to get latest version"
            exit 1
        fi
    fi
    log_info "Installing version: $VERSION"
}

# Download and install
install_mdriver() {
    BINARY_NAME="mdriver"
    if [ "$OS" = "darwin" ]; then
        ARCHIVE_NAME="${BINARY_NAME}_Darwin_${ARCH}.tar.gz"
    else
        ARCHIVE_NAME="${BINARY_NAME}_Linux_${ARCH}.tar.gz"
    fi
    
    DOWNLOAD_URL="https://github.com/$GITHUB_USER/$GITHUB_REPO/releases/download/$VERSION/$ARCHIVE_NAME"
    
    log_info "Downloading from: $DOWNLOAD_URL"
    
    # Create temporary directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT
    
    # Download archive
    if command -v curl >/dev/null; then
        curl -L "$DOWNLOAD_URL" -o "$TMP_DIR/$ARCHIVE_NAME"
    elif command -v wget >/dev/null; then
        wget "$DOWNLOAD_URL" -O "$TMP_DIR/$ARCHIVE_NAME"
    else
        log_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
    
    # Extract archive
    log_info "Extracting archive..."
    tar -xzf "$TMP_DIR/$ARCHIVE_NAME" -C "$TMP_DIR"
    
    # Install binary
    log_info "Installing to $INSTALL_DIR..."
    
    # Check if we need sudo
    if [ ! -w "$INSTALL_DIR" ]; then
        SUDO="sudo"
        log_warn "Insufficient permissions. Using sudo for installation."
    else
        SUDO=""
    fi
    
    $SUDO install -m 755 "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    
    # Verify installation
    if "$INSTALL_DIR/$BINARY_NAME" -version >/dev/null 2>&1; then
        log_info "Successfully installed $BINARY_NAME $VERSION"
        log_info "Run '$BINARY_NAME -platform-info' to check platform capabilities"
    else
        log_error "Installation verification failed"
        exit 1
    fi
}

# Main installation flow
main() {
    log_info "Starting mdriver installation..."
    
    # Check prerequisites
    if ! command -v tar >/dev/null; then
        log_error "tar is required but not installed"
        exit 1
    fi
    
    detect_platform
    get_latest_version
    install_mdriver
    
    log_info "Installation completed successfully!"
    echo
    echo "Next steps:"
    echo "  1. Create a config file: $BINARY_NAME -config config.yaml"
    echo "  2. Check platform info: $BINARY_NAME -platform-info"
    echo "  3. Run in sync mode: $BINARY_NAME"
    echo "  4. Run in mount mode: $BINARY_NAME -mount"
}

# Run main function
main "$@"