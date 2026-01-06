#!/bin/bash
# mcfetch Installer for Linux

set -e

REPO="Rezn1r/mcstatus"
API_URL="https://api.github.com/repos/$REPO/releases/latest"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
BINARY_NAME="mcfetch"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}mcfetch Installer${NC}"
echo -e "${CYAN}=================${NC}"
echo ""

if [ "$EUID" -eq 0 ]; then
    INSTALL_DIR="/usr/local/bin"
    echo -e "${YELLOW}Running as root - installing system-wide to $INSTALL_DIR${NC}"
else
    echo -e "${YELLOW}Installing to user directory: $INSTALL_DIR${NC}"
fi

# Fetch latest release information
echo -e "${YELLOW}Fetching latest release information...${NC}"

if command -v curl &> /dev/null; then
    RESPONSE=$(curl -sSL "$API_URL" -H "User-Agent: mcfetch-installer")
elif command -v wget &> /dev/null; then
    RESPONSE=$(wget -qO- "$API_URL" --header="User-Agent: mcfetch-installer")
else
    echo -e "${RED}Error: Neither curl nor wget found. Please install one of them.${NC}"
    exit 1
fi

VERSION=$(echo "$RESPONSE" | grep -o '"tag_name": *"[^"]*"' | head -1 | sed 's/"tag_name": *"\(.*\)"/\1/')

if [ -z "$VERSION" ]; then
    echo -e "${RED}Error: Could not fetch release information.${NC}"
    exit 1
fi

echo -e "${GREEN}Latest version: $VERSION${NC}"

DOWNLOAD_URL=$(echo "$RESPONSE" | grep -o '"browser_download_url": *"[^"]*linux[^"]*amd64[^"]*"' | head -1 | sed 's/"browser_download_url": *"\(.*\)"/\1/')

if [ -z "$DOWNLOAD_URL" ]; then
    DOWNLOAD_URL=$(echo "$RESPONSE" | grep -o '"browser_download_url": *"[^"]*linux[^"]*"' | head -1 | sed 's/"browser_download_url": *"\(.*\)"/\1/')
fi

if [ -z "$DOWNLOAD_URL" ]; then
    echo -e "${RED}Error: No Linux binary found in latest release.${NC}"
    echo -e "${YELLOW}Available assets:${NC}"
    echo "$RESPONSE" | grep -o '"name": *"[^"]*"' | sed 's/"name": *"\(.*\)"/  - \1/'
    exit 1
fi

ASSET_NAME=$(basename "$DOWNLOAD_URL")
echo -e "${GREEN}Found asset: $ASSET_NAME${NC}"
echo ""

if [ ! -d "$INSTALL_DIR" ]; then
    echo -e "${YELLOW}Creating installation directory: $INSTALL_DIR${NC}"
    mkdir -p "$INSTALL_DIR"
fi

INSTALL_PATH="$INSTALL_DIR/$BINARY_NAME"

echo -e "${YELLOW}Downloading $ASSET_NAME...${NC}"

if command -v curl &> /dev/null; then
    curl -sSL "$DOWNLOAD_URL" -o "$INSTALL_PATH"
elif command -v wget &> /dev/null; then
    wget -qO "$INSTALL_PATH" "$DOWNLOAD_URL"
fi

chmod +x "$INSTALL_PATH"

echo -e "${GREEN}Downloaded successfully!${NC}"
echo ""

if [ ! -f "$INSTALL_PATH" ]; then
    echo -e "${RED}Error: Installation failed.${NC}"
    exit 1
fi

FILE_SIZE=$(du -h "$INSTALL_PATH" | cut -f1)
echo -e "${GREEN}Installed to: $INSTALL_PATH ($FILE_SIZE)${NC}"

echo ""
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo -e "${YELLOW}Warning: $INSTALL_DIR is not in your PATH${NC}"
    echo ""
    echo "Add this line to your ~/.bashrc or ~/.zshrc:"
    echo -e "${CYAN}  export PATH=\"\$PATH:$INSTALL_DIR\"${NC}"
    echo ""
    echo "Then run:"
    echo -e "${CYAN}  source ~/.bashrc${NC}  # or source ~/.zshrc"
else
    echo -e "${GREEN}$INSTALL_DIR is in your PATH${NC}"
fi

echo ""
echo -e "${GREEN}Installation complete! ✓${NC}"
echo ""
echo -e "${CYAN}Usage:${NC}"
echo "  mcfetch java donutsmp.net"
echo "  mcfetch bedrock demo.mcstatus.io --verbose"
