#!/bin/bash
set -e

REPO="GareArc/opencode-sync"
BINARY_NAME="opencode-sync"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

info() {
    echo -e "${BLUE}==>${NC} $1" >&2
}

success() {
    echo -e "${GREEN}==>${NC} $1" >&2
}

warn() {
    echo -e "${YELLOW}==>${NC} $1" >&2
}

error() {
    echo -e "${RED}==>${NC} $1" >&2
    exit 1
}

detect_os() {
    local os
    os="$(uname -s)"
    case "$os" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *)       error "Unsupported operating system: $os" ;;
    esac
}

detect_arch() {
    local arch
    arch="$(uname -m)"
    case "$arch" in
        x86_64|amd64)  echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *)             error "Unsupported architecture: $arch" ;;
    esac
}

get_latest_version() {
    local version
    version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$version" ]; then
        error "Failed to fetch latest version"
    fi
    echo "$version"
}

download_binary() {
    local os="$1"
    local arch="$2"
    local version="$3"
    local ext=""
    local archive_ext="tar.gz"
    
    if [ "$os" = "windows" ]; then
        ext=".exe"
        archive_ext="zip"
    fi
    
    local filename="${BINARY_NAME}_${version#v}_${os}_${arch}.${archive_ext}"
    local url="https://github.com/${REPO}/releases/download/${version}/${filename}"
    local tmp_dir
    tmp_dir=$(mktemp -d)
    
    info "Downloading ${BINARY_NAME} ${version} for ${os}/${arch}..."
    
    if ! curl -fsSL "$url" -o "${tmp_dir}/${filename}"; then
        rm -rf "$tmp_dir"
        error "Failed to download from ${url}"
    fi
    
    info "Extracting..."
    cd "$tmp_dir"
    
    if [ "$archive_ext" = "zip" ]; then
        if command -v unzip &> /dev/null; then
            unzip -q "$filename"
        else
            error "unzip is required to extract Windows binaries"
        fi
    else
        tar -xzf "$filename"
    fi
    
    echo "${tmp_dir}/${BINARY_NAME}${ext}"
}

install_binary() {
    local binary_path="$1"
    local install_path="${INSTALL_DIR}/${BINARY_NAME}"
    
    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        mv "$binary_path" "$install_path"
        chmod +x "$install_path"
    else
        info "Requesting sudo to install to ${INSTALL_DIR}..."
        sudo mv "$binary_path" "$install_path"
        sudo chmod +x "$install_path"
    fi
    
    success "Installed ${BINARY_NAME} to ${install_path}"
}

verify_installation() {
    if command -v "$BINARY_NAME" &> /dev/null; then
        success "Installation verified!"
        echo ""
        "$BINARY_NAME" version
    else
        warn "Installation complete, but ${BINARY_NAME} is not in PATH"
        warn "Add ${INSTALL_DIR} to your PATH or run: ${INSTALL_DIR}/${BINARY_NAME}"
    fi
}

check_dependencies() {
    if ! command -v curl &> /dev/null; then
        error "curl is required but not installed"
    fi
    
    if ! command -v tar &> /dev/null; then
        error "tar is required but not installed"
    fi
}

create_install_dir() {
    if [ ! -d "$INSTALL_DIR" ]; then
        info "Creating install directory: ${INSTALL_DIR}"
        if [ -w "$(dirname "$INSTALL_DIR")" ]; then
            mkdir -p "$INSTALL_DIR"
        else
            sudo mkdir -p "$INSTALL_DIR"
        fi
    fi
}

main() {
    echo ""
    echo "  ┌─────────────────────────────────────┐"
    echo "  │       opencode-sync installer       │"
    echo "  └─────────────────────────────────────┘"
    echo ""
    
    check_dependencies
    
    local os arch version binary_path
    os=$(detect_os)
    arch=$(detect_arch)
    
    info "Detected: ${os}/${arch}"
    
    # Allow version override
    version="${VERSION:-$(get_latest_version)}"
    
    info "Version: ${version}"
    
    # Handle user-local installation if /usr/local/bin is not writable
    if [ ! -w "$INSTALL_DIR" ] && [ ! -w "$(dirname "$INSTALL_DIR")" ]; then
        if [ -z "$SUDO_USER" ] && [ "$(id -u)" -ne 0 ]; then
            # Try user-local installation first
            local user_bin="$HOME/.local/bin"
            warn "${INSTALL_DIR} is not writable"
            info "Attempting to install to ${user_bin} instead..."
            INSTALL_DIR="$user_bin"
        fi
    fi
    
    create_install_dir
    
    binary_path=$(download_binary "$os" "$arch" "$version")
    install_binary "$binary_path"
    
    # Cleanup
    rm -rf "$(dirname "$binary_path")"
    
    echo ""
    verify_installation
    
    echo ""
    success "Done! Run 'opencode-sync' to get started."
    echo ""
    echo "  Quick start:"
    echo "    opencode-sync setup    # First-time setup"
    echo "    opencode-sync sync     # Sync your configs"
    echo "    opencode-sync --help   # Show all commands"
    echo ""
}

main "$@"
