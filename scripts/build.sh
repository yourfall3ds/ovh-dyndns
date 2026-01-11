#!/bin/bash

# Bash script to build OVH DynDNS for Windows and Linux

# Colors
GREEN='\033[0;32m'
CYAN='\033[0;36m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting compilation...${NC}"

# Configuration
BUILD_DIR="build"

# Ensure build directory exists
mkdir -p "$BUILD_DIR"

# Clean build directory
rm -f "$BUILD_DIR"/*

# Helper function for building
build_target() {
    local os=$1
    local arch=$2
    local output=$3
    local output_path="$BUILD_DIR/$output"

    echo -e "${CYAN}Building for $os ($arch)...${NC}"

    env GOOS=$os GOARCH=$arch go build -o "$output_path" ./cmd/ovh-dyndns

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}$os build: OK${NC}"
    else
        echo -e "${RED}$os build: ERROR${NC}"
        exit 1
    fi
}

echo ""

# Build for Windows
build_target "windows" "amd64" "windows_amd64.exe"

# Build for Linux
build_target "linux" "amd64" "linux_amd64"

echo ""
echo -e "${GREEN}Compilation completed successfully!${NC}"
echo ""
echo -e "${YELLOW}Generated files in '$BUILD_DIR/':${NC}"

# List files with size
ls -lh "$BUILD_DIR" | grep -v '^total' | awk '{print "  - " $9 " (" $5 ")"}'

echo ""
echo -e "${CYAN}Instructions:${NC}"
echo "  Windows: ./$BUILD_DIR/windows_amd64.exe"
echo "  Linux:   ./$BUILD_DIR/linux_amd64"
echo ""

# Make linux binary executable automatically if on linux/mac
if [[ "$OSTYPE" == "linux-gnu"* ]] || [[ "$OSTYPE" == "darwin"* ]]; then
    chmod +x "$BUILD_DIR/linux_amd64"
fi

