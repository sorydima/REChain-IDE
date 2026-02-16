#!/bin/bash
# install-hooks.sh - Install git hooks for the REChain project
# Usage: ./scripts/install-hooks.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
HOOKS_DIR="${PROJECT_ROOT}/.githooks"
GIT_HOOKS_DIR="${PROJECT_ROOT}/.git/hooks"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "${GREEN}Installing Git Hooks for REChain IDE...${NC}"

# Check if we're in a git repository
if [ ! -d "${PROJECT_ROOT}/.git" ]; then
    echo "${RED}Error: Not a git repository${NC}"
    exit 1
fi

# Create hooks directory if it doesn't exist
mkdir -p "${GIT_HOOKS_DIR}"

# Install each hook
for hook in pre-commit commit-msg pre-push pre-rebase post-checkout post-merge; do
    source_file="${HOOKS_DIR}/${hook}.sample"
    target_file="${GIT_HOOKS_DIR}/${hook}"
    
    if [ -f "${source_file}" ]; then
        cp "${source_file}" "${target_file}"
        chmod +x "${target_file}"
        echo "${GREEN}✓ Installed ${hook}${NC}"
    else
        echo "${YELLOW}⚠ Sample hook not found: ${hook}.sample${NC}"
    fi
done

echo ""
echo "${GREEN}Git hooks installed successfully!${NC}"
echo ""
echo "Installed hooks:"
ls -la "${GIT_HOOKS_DIR}" | grep -v "\.sample$" | tail -n +2
echo ""
echo "To uninstall, run: rm -rf ${GIT_HOOKS_DIR}/*"
