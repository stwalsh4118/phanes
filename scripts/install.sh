#!/bin/sh
# Phanes Installation Script
# This script downloads and installs phanes without requiring Go to be installed.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/stwalsh4118/phanes/main/scripts/install.sh | sh
#   
#   Or with a specific version:
#   curl -fsSL https://raw.githubusercontent.com/stwalsh4118/phanes/main/scripts/install.sh | sh -s -- --version v0.1.0
#
# Options:
#   --version VERSION   Install a specific version (default: latest)
#   --install-dir DIR   Installation directory (default: /usr/local/bin)
#   --help              Show this help message

set -e

# Configuration
GITHUB_REPO="stwalsh4118/phanes"
BINARY_NAME="phanes"
DEFAULT_INSTALL_DIR="/usr/local/bin"

# Colors for output (only if terminal supports it)
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[0;33m'
    BLUE='\033[0;34m'
    NC='\033[0m' # No Color
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    NC=''
fi

# Print functions
info() {
    printf "${BLUE}[INFO]${NC} %s\n" "$1"
}

success() {
    printf "${GREEN}[SUCCESS]${NC} %s\n" "$1"
}

warn() {
    printf "${YELLOW}[WARN]${NC} %s\n" "$1"
}

error() {
    printf "${RED}[ERROR]${NC} %s\n" "$1" >&2
}

# Show help message
show_help() {
    cat << EOF
Phanes Installation Script

Usage:
    curl -fsSL https://raw.githubusercontent.com/stwalsh4118/phanes/main/scripts/install.sh | sh
    
    Or with options:
    curl -fsSL ... | sh -s -- [OPTIONS]

Options:
    --version VERSION   Install a specific version (default: latest)
    --install-dir DIR   Installation directory (default: /usr/local/bin)
    --help              Show this help message

Examples:
    # Install latest version
    curl -fsSL https://raw.githubusercontent.com/stwalsh4118/phanes/main/scripts/install.sh | sh

    # Install specific version
    curl -fsSL https://raw.githubusercontent.com/stwalsh4118/phanes/main/scripts/install.sh | sh -s -- --version v0.1.0

    # Install to custom directory
    curl -fsSL https://raw.githubusercontent.com/stwalsh4118/phanes/main/scripts/install.sh | sh -s -- --install-dir ~/.local/bin
EOF
}

# Detect OS
detect_os() {
    OS="$(uname -s)"
    case "${OS}" in
        Linux*)     OS=linux;;
        Darwin*)    OS=darwin;;
        *)          error "Unsupported operating system: ${OS}"; exit 1;;
    esac
    echo "${OS}"
}

# Detect architecture
detect_arch() {
    ARCH="$(uname -m)"
    case "${ARCH}" in
        x86_64|amd64)   ARCH=amd64;;
        aarch64|arm64)  ARCH=arm64;;
        armv7l)         ARCH=arm;;
        i386|i686)      ARCH=386;;
        *)              error "Unsupported architecture: ${ARCH}"; exit 1;;
    esac
    echo "${ARCH}"
}

# Get latest release version from GitHub
get_latest_version() {
    # Note: Do not use info/warn here - this function's stdout is captured
    # Try using curl first, then wget
    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -sL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        echo "ERROR: Neither curl nor wget found" >&2
        exit 1
    fi
    
    if [ -z "${VERSION}" ]; then
        echo "ERROR: Failed to fetch latest version" >&2
        exit 1
    fi
    
    echo "${VERSION}"
}

# Download and install binary
download_and_install() {
    VERSION="$1"
    INSTALL_DIR="$2"
    OS="$3"
    ARCH="$4"
    
    # Build the download URL
    FILENAME="${BINARY_NAME}_${VERSION#v}_${OS}_${ARCH}.tar.gz"
    DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${FILENAME}"
    
    info "Downloading ${BINARY_NAME} ${VERSION} for ${OS}/${ARCH}..."
    info "URL: ${DOWNLOAD_URL}"
    
    # Create temporary directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf ${TMP_DIR}" EXIT
    
    # Download the archive
    if command -v curl >/dev/null 2>&1; then
        if ! curl -fsSL "${DOWNLOAD_URL}" -o "${TMP_DIR}/${FILENAME}"; then
            error "Failed to download ${DOWNLOAD_URL}"
            error "Please check if the version exists and try again."
            exit 1
        fi
    elif command -v wget >/dev/null 2>&1; then
        if ! wget -q "${DOWNLOAD_URL}" -O "${TMP_DIR}/${FILENAME}"; then
            error "Failed to download ${DOWNLOAD_URL}"
            error "Please check if the version exists and try again."
            exit 1
        fi
    fi
    
    # Extract the archive
    info "Extracting archive..."
    cd "${TMP_DIR}"
    tar -xzf "${FILENAME}"
    
    # Make binary executable
    chmod +x "${BINARY_NAME}"
    
    # Install binary
    info "Installing ${BINARY_NAME} to ${INSTALL_DIR}..."
    
    # Check if we need sudo
    if [ -w "${INSTALL_DIR}" ]; then
        mv "${BINARY_NAME}" "${INSTALL_DIR}/"
    else
        warn "Installation directory requires elevated privileges."
        if command -v sudo >/dev/null 2>&1; then
            sudo mv "${BINARY_NAME}" "${INSTALL_DIR}/"
        else
            error "Cannot write to ${INSTALL_DIR} and sudo is not available."
            error "Try running as root or use --install-dir to specify a writable directory."
            exit 1
        fi
    fi
    
    success "${BINARY_NAME} ${VERSION} installed successfully to ${INSTALL_DIR}/${BINARY_NAME}"
}

# Verify installation
verify_installation() {
    INSTALL_DIR="$1"
    
    if [ -x "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        info "Verifying installation..."
        VERSION_OUTPUT=$("${INSTALL_DIR}/${BINARY_NAME}" --version 2>&1 || true)
        if [ -n "${VERSION_OUTPUT}" ]; then
            success "Installed version: ${VERSION_OUTPUT}"
        else
            success "Binary installed and executable"
        fi
        
        # Check if install dir is in PATH
        case ":${PATH}:" in
            *":${INSTALL_DIR}:"*) ;;
            *) warn "${INSTALL_DIR} is not in your PATH. Add it with: export PATH=\"\$PATH:${INSTALL_DIR}\"";;
        esac
    else
        error "Installation verification failed"
        exit 1
    fi
}

# Main installation function
main() {
    VERSION=""
    INSTALL_DIR="${DEFAULT_INSTALL_DIR}"
    
    # Parse arguments
    while [ $# -gt 0 ]; do
        case "$1" in
            --version)
                VERSION="$2"
                shift 2
                ;;
            --install-dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    echo ""
    info "=== Phanes Installer ==="
    echo ""
    
    # Detect platform
    OS=$(detect_os)
    ARCH=$(detect_arch)
    info "Detected platform: ${OS}/${ARCH}"
    
    # Get version if not specified
    if [ -z "${VERSION}" ]; then
        info "Fetching latest version..."
        VERSION=$(get_latest_version)
    fi
    info "Version to install: ${VERSION}"
    
    # Create install directory if it doesn't exist
    if [ ! -d "${INSTALL_DIR}" ]; then
        info "Creating installation directory: ${INSTALL_DIR}"
        if [ -w "$(dirname "${INSTALL_DIR}")" ]; then
            mkdir -p "${INSTALL_DIR}"
        else
            sudo mkdir -p "${INSTALL_DIR}"
        fi
    fi
    
    # Download and install
    download_and_install "${VERSION}" "${INSTALL_DIR}" "${OS}" "${ARCH}"
    
    # Verify
    verify_installation "${INSTALL_DIR}"
    
    echo ""
    success "Installation complete!"
    echo ""
    info "Quick start:"
    echo "  1. Create a config file:  ${BINARY_NAME} --list"
    echo "  2. Run a profile:         ${BINARY_NAME} --profile dev --config config.yaml"
    echo "  3. Preview changes:       ${BINARY_NAME} --profile dev --config config.yaml --dry-run"
    echo ""
}

# Run main function
main "$@"

