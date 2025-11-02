#!/bin/bash

# Setup IDE for Go Development
# This script ensures your environment is properly configured for Go development in Cursor

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check if Go is installed
check_go() {
    log_info "Checking Go installation..."
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed!"
        log_info "Please install Go from: https://golang.org/dl/"
        exit 1
    fi
    log_success "Go is installed: $(go version)"
}

# Install gopls if not present
install_gopls() {
    log_info "Installing Go language server (gopls)..."
    if ! command -v gopls &> /dev/null; then
        go install golang.org/x/tools/gopls@latest
        log_success "gopls installed"
    else
        log_success "gopls already installed"
    fi
}

# Setup Go environment
setup_go_env() {
    log_info "Setting up Go environment..."
    
    # Get GOPATH
    GOPATH=$(go env GOPATH)
    log_info "GOPATH: $GOPATH"
    
    # Add GOPATH/bin to PATH if not already there
    if [[ ":$PATH:" != *":$GOPATH/bin:"* ]]; then
        log_info "Adding $GOPATH/bin to PATH"
        echo "export PATH=\$PATH:$GOPATH/bin" >> ~/.zshrc
        export PATH=$PATH:$GOPATH/bin
        log_success "PATH updated"
    else
        log_success "PATH already includes $GOPATH/bin"
    fi
}

# Verify Go module
verify_module() {
    log_info "Verifying Go module..."
    if [ ! -f "go.mod" ]; then
        log_error "go.mod not found!"
        exit 1
    fi
    
    go mod tidy
    go mod verify
    log_success "Go module verified"
}

# Test gopls
test_gopls() {
    log_info "Testing Go language server..."
    if command -v gopls &> /dev/null; then
        gopls version
        log_success "gopls is working"
    else
        log_error "gopls not found in PATH"
        log_info "Try running: source ~/.zshrc"
        exit 1
    fi
}

# Main setup function
main() {
    echo "================================================"
    echo "Go IDE Setup for Cursor"
    echo "================================================"
    echo ""
    
    check_go
    install_gopls
    setup_go_env
    verify_module
    test_gopls
    
    echo ""
    log_success "Setup completed!"
    echo ""
    log_info "Next steps:"
    log_info "1. Restart Cursor"
    log_info "2. Or run: Cmd+Shift+P → 'Go: Restart Language Server'"
    log_info "3. Check that syntax highlighting and navigation work"
    echo ""
    log_info "If it still doesn't work:"
    log_info "- Make sure Cursor has the Go extension installed"
    log_info "- Check Cursor settings for Go language server path"
}

main "$@"

