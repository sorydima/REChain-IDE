#!/bin/bash
# setup-repo.sh - Initial repository setup for new contributors
# Usage: ./scripts/setup-repo.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "${BLUE}============================================${NC}"
echo "${BLUE}  REChain IDE - Repository Setup Script${NC}"
echo "${BLUE}============================================${NC}"
echo ""

# Check if we're in a git repository
if [ ! -d "${PROJECT_ROOT}/.git" ]; then
    echo "${RED}Error: Not a git repository${NC}"
    exit 1
fi

# Check for required tools
echo "${YELLOW}Checking required tools...${NC}"

command -v git >/dev/null 2>&1 || { echo "${RED}Error: git is required but not installed${NC}"; exit 1; }
echo "${GREEN}✓ git${NC}"

command -v go >/dev/null 2>&1 || { echo "${YELLOW}⚠ go is not installed (required for development)${NC}"; }
[ -x "$(command -v go)" ] && echo "${GREEN}✓ go${NC}"

# Configure git
echo ""
echo "${YELLOW}Configuring git...${NC}"

# Set up git hooks
echo "Installing git hooks..."
if [ -f "${SCRIPT_DIR}/install-hooks.sh" ]; then
    bash "${SCRIPT_DIR}/install-hooks.sh"
else
    echo "${YELLOW}⚠ Hook installer not found${NC}"
fi

# Check git config
echo ""
echo "${YELLOW}Checking git configuration...${NC}"
if [ -z "$(git config user.name)" ]; then
    echo "${YELLOW}⚠ Git user.name not set${NC}"
    read -p "Enter your name: " git_name
    git config user.name "$git_name"
fi

if [ -z "$(git config user.email)" ]; then
    echo "${YELLOW}⚠ Git user.email not set${NC}"
    read -p "Enter your email: " git_email
    git config user.email "$git_email"
fi

echo "${GREEN}✓ Git configured:${NC}"
echo "  Name: $(git config user.name)"
echo "  Email: $(git config user.email)"

# Set up Go workspace if needed
if [ -f "${PROJECT_ROOT}/go.work" ]; then
    echo ""
    echo "${YELLOW}Setting up Go workspace...${NC}"
    cd "${PROJECT_ROOT}"
    go work sync || echo "${YELLOW}⚠ Could not sync Go workspace${NC}"
fi

# Run initial build check
echo ""
echo "${YELLOW}Running initial build check...${NC}"
if [ -d "${PROJECT_ROOT}/rechain-ide" ]; then
    cd "${PROJECT_ROOT}/rechain-ide"
    if go build ./... 2>/dev/null; then
        echo "${GREEN}✓ Initial build successful${NC}"
    else
        echo "${YELLOW}⚠ Initial build had issues (this is OK, dependencies may need to be installed)${NC}"
    fi
fi

echo ""
echo "${GREEN}============================================${NC}"
echo "${GREEN}  Repository setup complete!${NC}"
echo "${GREEN}============================================${NC}"
echo ""
echo "Next steps:"
echo "  1. Review the project documentation: ${PROJECT_ROOT}/README.md"
echo "  2. Check the contributing guide: ${PROJECT_ROOT}/CONTRIBUTING.md"
echo "  3. Set up your IDE/editor"
echo "  4. Create a feature branch: git checkout -b feature/your-feature"
echo ""
