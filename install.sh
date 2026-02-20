#!/bin/sh
set -euo pipefail

# agit installer
# Usage: curl -sSfL https://raw.githubusercontent.com/Fato07/agit/main/install.sh | sh

REPO="Fato07/agit"
BINARY="agit"
INSTALL_DIR="/usr/local/bin"

# Colors (degrade gracefully if no tty)
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[0;33m'
    BLUE='\033[0;34m'
    NC='\033[0m'
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    NC=''
fi

info() { printf "${BLUE}==>${NC} %s\n" "$1"; }
success() { printf "${GREEN}==>${NC} %s\n" "$1"; }
warn() { printf "${YELLOW}==>${NC} %s\n" "$1"; }
error() { printf "${RED}ERROR:${NC} %s\n" "$1" >&2; exit 1; }

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        *)       error "Unsupported operating system: $(uname -s). agit supports Linux and macOS." ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)   echo "amd64" ;;
        aarch64|arm64)  echo "arm64" ;;
        *)              error "Unsupported architecture: $(uname -m). agit supports amd64 and arm64." ;;
    esac
}

# HTTP client (curl preferred, wget fallback)
http_get() {
    if command -v curl >/dev/null 2>&1; then
        curl -sSfL "$1"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "$1"
    else
        error "Neither curl nor wget found. Please install one of them."
    fi
}

http_download() {
    if command -v curl >/dev/null 2>&1; then
        curl -sSfL -o "$2" "$1"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$2" "$1"
    else
        error "Neither curl nor wget found. Please install one of them."
    fi
}

# Verify checksum
verify_checksum() {
    local file="$1"
    local checksums="$2"
    local filename
    filename=$(basename "$file")

    local expected
    expected=$(grep "$filename" "$checksums" | awk '{print $1}')

    if [ -z "$expected" ]; then
        error "Checksum not found for $filename in checksums file."
    fi

    local actual
    if command -v sha256sum >/dev/null 2>&1; then
        actual=$(sha256sum "$file" | awk '{print $1}')
    elif command -v shasum >/dev/null 2>&1; then
        actual=$(shasum -a 256 "$file" | awk '{print $1}')
    else
        warn "Neither sha256sum nor shasum found. Skipping checksum verification."
        return 0
    fi

    if [ "$actual" != "$expected" ]; then
        error "Checksum verification failed.\n  Expected: $expected\n  Actual:   $actual"
    fi
}

main() {
    local os arch version tag archive_name download_url checksums_url tmpdir

    os=$(detect_os)
    arch=$(detect_arch)

    info "Detected OS: $os, Arch: $arch"

    # Fetch latest release tag
    info "Fetching latest release..."
    tag=$(http_get "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')

    if [ -z "$tag" ]; then
        error "Could not determine latest release. Check https://github.com/${REPO}/releases"
    fi

    version="${tag#v}"
    info "Latest version: $version"

    # Build archive name
    archive_name="${BINARY}_${version}_${os}_${arch}.tar.gz"
    download_url="https://github.com/${REPO}/releases/download/${tag}/${archive_name}"
    checksums_url="https://github.com/${REPO}/releases/download/${tag}/checksums.txt"

    # Create temp directory
    tmpdir=$(mktemp -d)
    trap 'rm -rf "$tmpdir"' EXIT

    # Download archive and checksums
    info "Downloading ${archive_name}..."
    http_download "$download_url" "$tmpdir/$archive_name"
    http_download "$checksums_url" "$tmpdir/checksums.txt"

    # Verify checksum
    info "Verifying checksum..."
    verify_checksum "$tmpdir/$archive_name" "$tmpdir/checksums.txt"
    success "Checksum verified."

    # Extract
    info "Extracting..."
    tar -xzf "$tmpdir/$archive_name" -C "$tmpdir"

    # Install
    if [ -w "$INSTALL_DIR" ]; then
        install -m 755 "$tmpdir/$BINARY" "$INSTALL_DIR/$BINARY"
        success "Installed $BINARY to $INSTALL_DIR/$BINARY"
    else
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR"
        install -m 755 "$tmpdir/$BINARY" "$INSTALL_DIR/$BINARY"
        success "Installed $BINARY to $INSTALL_DIR/$BINARY"
        warn "Make sure $INSTALL_DIR is in your PATH."
    fi

    # Verify installation
    if command -v "$BINARY" >/dev/null 2>&1; then
        success "$BINARY installed successfully! Run '$BINARY --version' to verify."
    else
        success "$BINARY installed to $INSTALL_DIR/$BINARY"
        echo ""
        echo "Next steps:"
        echo "  1. Add $INSTALL_DIR to your PATH (if not already)"
        echo "  2. Run: $BINARY init"
        echo "  3. Run: $BINARY add /path/to/your/repo"
    fi
}

main
