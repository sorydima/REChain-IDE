#!/bin/bash
# bump-version.sh - Bump the project version
# Usage: ./scripts/bump-version.sh <major|minor|patch>

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Help message
if [ "$1" == "-h" ] || [ "$1" == "--help" ] || [ $# -lt 1 ]; then
    echo "${BLUE}Version Bump Script${NC}"
    echo ""
    echo "Usage: $0 <major|minor|patch> [--dry-run]"
    echo ""
    echo "Options:"
    echo "  major     - Bump major version (1.0.0 -> 2.0.0)"
    echo "  minor     - Bump minor version (1.0.0 -> 1.1.0)"
    echo "  patch     - Bump patch version (1.0.0 -> 1.0.1)"
    echo "  --dry-run - Show what would be done without making changes"
    echo ""
    exit 0
fi

BUMP_TYPE=$1
DRY_RUN=false

if [ "$2" == "--dry-run" ]; then
    DRY_RUN=true
fi

# Validate bump type
if [[ ! "$BUMP_TYPE" =~ ^(major|minor|patch)$ ]]; then
    echo "${RED}Error: Invalid bump type '$BUMP_TYPE'${NC}"
    echo "Valid types: major, minor, patch"
    exit 1
fi

# Get current version
CURRENT_VERSION=$(grep -E '"version":' "${PROJECT_ROOT}/PROJECT.json" 2>/dev/null | sed -E 's/.*"version": *"([^"]+)".*/\1/')

if [ -z "$CURRENT_VERSION" ]; then
    echo "${RED}Error: Could not determine current version${NC}"
    exit 1
fi

# Parse version
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT_VERSION"

# Calculate new version
 case $BUMP_TYPE in
    major)
        NEW_MAJOR=$((MAJOR + 1))
        NEW_MINOR=0
        NEW_PATCH=0
        ;;
    minor)
        NEW_MAJOR=$MAJOR
        NEW_MINOR=$((MINOR + 1))
        NEW_PATCH=0
        ;;
    patch)
        NEW_MAJOR=$MAJOR
        NEW_MINOR=$MINOR
        NEW_PATCH=$((PATCH + 1))
        ;;
esac

NEW_VERSION="${NEW_MAJOR}.${NEW_MINOR}.${NEW_PATCH}"

echo "${BLUE}Version Bump${NC}"
echo "Current: ${CURRENT_VERSION}"
echo "New:     ${NEW_VERSION}"
echo ""

if [ "$DRY_RUN" = true ]; then
    echo "${YELLOW}Dry run - no changes made${NC}"
    exit 0
fi

# Check if we're on a release branch or main
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "$CURRENT_BRANCH" != "main" ] && [[ ! "$CURRENT_BRANCH" =~ ^release/ ]]; then
    echo "${YELLOW}Warning: Not on main or release branch (current: $CURRENT_BRANCH)${NC}"
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Update version in files
echo "${YELLOW}Updating version in files...${NC}"

# Update PROJECT.json
if [ -f "${PROJECT_ROOT}/PROJECT.json" ]; then
    sed -i.bak "s/\"version\": \"${CURRENT_VERSION}\"/\"version\": \"${NEW_VERSION}\"/" "${PROJECT_ROOT}/PROJECT.json"
    rm -f "${PROJECT_ROOT}/PROJECT.json.bak"
    echo "${GREEN}✓ Updated PROJECT.json${NC}"
fi

# Create git commit and tag
echo "${YELLOW}Creating commit and tag...${NC}"
git add -A
git commit -m "chore(release): bump version to ${NEW_VERSION}"
git tag -a "v${NEW_VERSION}" -m "Release v${NEW_VERSION}"

echo ""
echo "${GREEN}============================================${NC}"
echo "${GREEN}  Version bumped to ${NEW_VERSION}!${NC}"
echo "${GREEN}============================================${NC}"
echo ""
echo "To complete the release:"
echo "  1. Push the commit: git push origin ${CURRENT_BRANCH}"
echo "  2. Push the tag:    git push origin v${NEW_VERSION}"
echo ""
