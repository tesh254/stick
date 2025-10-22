#!/bin/bash

# install.sh - Install stick CLI tool
# Usage: curl -fsSL https://raw.githubusercontent.com/tesh254/stick/main/install.sh | bash
# Or: wget -qO- https://raw.githubusercontent.com/tesh254/stick/main/install.sh | bash

set -e

# Default installation directory
INSTALL_DIR="/usr/local/bin"

# GitHub repository
REPO="tesh254/stick"

# Function to print messages
print_msg() {
    echo -e "\033[1;34m[INFO]\033[0m $1"
}

print_error() {
    echo -e "\033[1;31m[ERROR]\033[0m $1" >&2
}

print_success() {
    echo -e "\033[1;32m[SUCCESS]\033[0m $1"
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    # Map architecture names to our release naming
    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    # Map OS names to our release naming
    case $OS in
        linux|darwin)
            # Supported
            ;;
        *)
            print_error "Unsupported OS: $OS"
            exit 1
            ;;
    esac

    BINARY_NAME="stick-$OS-$ARCH"
}

# Get the latest release version
get_latest_version() {
    LATEST_VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$LATEST_VERSION" ]; then
        print_error "Failed to get the latest version from GitHub"
        exit 1
    fi

    print_msg "Latest version: $LATEST_VERSION"
}

# Download and install the binary
download_and_install() {
    TEMP_DIR=$(mktemp -d)
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_VERSION/$BINARY_NAME"

    print_msg "Downloading $BINARY_NAME from $DOWNLOAD_URL"
    curl -L -o "$TEMP_DIR/$BINARY_NAME" "$DOWNLOAD_URL"

    # Check if download was successful
    if [ ! -f "$TEMP_DIR/$BINARY_NAME" ]; then
        print_error "Failed to download the binary"
        rm -rf "$TEMP_DIR"
        exit 1
    fi

    # Make the binary executable
    chmod +x "$TEMP_DIR/$BINARY_NAME"

    # Determine installation directory
    if [ "$EUID" -eq 0 ]; then
        # Running as root, use default directory
        TARGET_DIR="$INSTALL_DIR"
    else
        # Not running as root, check if user has write access to default directory
        if [ -w "$INSTALL_DIR" ]; then
            TARGET_DIR="$INSTALL_DIR"
        else
            # Use ~/.local/bin instead
            TARGET_DIR="$HOME/.local/bin"
            mkdir -p "$TARGET_DIR"
            print_msg "Installing to $TARGET_DIR (not in standard location)"
        fi
    fi

    # Install the binary
    sudo cp "$TEMP_DIR/$BINARY_NAME" "$TARGET_DIR/stick"
    sudo chmod +x "$TARGET_DIR/stick"

    # Clean up
    rm -rf "$TEMP_DIR"

    print_success "stick installed successfully to $TARGET_DIR/stick"
}

# Check prerequisites
check_prerequisites() {
    if ! command -v curl &>/dev/null; then
        print_error "curl is required but not installed"
        exit 1
    fi

    if ! command -v sudo &>/dev/null; then
        print_error "sudo is required but not installed"
        exit 1
    fi
}

# Main execution
main() {
    print_msg "Installing stick CLI tool..."

    check_prerequisites
    detect_platform
    get_latest_version
    download_and_install

    print_msg "To verify the installation, run: stick version"
    print_success "Installation completed!"
}

main "$@"