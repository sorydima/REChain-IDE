#!/bin/bash
# create-branch.sh - Create a new feature/bugfix branch with proper naming
# Usage: ./scripts/create-branch.sh <type> <name>
# Example: ./scripts/create-branch.sh feature add-login

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
if [ "$1" == "-h" ] || [ "$1" == "--help" ] || [ $# -lt 2 ]; then
    echo "${BLUE}Create Branch Script${NC}"
    echo ""
    echo "Usage: $0 <type> <name>"
    echo ""
    echo "Types:"
    echo "  feature   - New feature (branch from develop)"
    echo "  bugfix    - Bug fix (branch from develop)"
    echo "  hotfix    - Critical fix (branch from main)"
    echo "  release   - Release preparation (branch from develop)"
    echo "  docs      - Documentation changes (branch from develop)"
    echo "  refactor  - Code refactoring (branch from develop)"
    echo ""
    echo "Examples:"
    echo "  $0 feature add-login"
    echo "  $0 bugfix fix-memory-leak"
    echo "  $0 hotfix critical-security-fix"
    exit 0
fi

BRANCH_TYPE=$1
BRANCH_NAME=$2

# Convert branch name to kebab-case (lowercase with hyphens)
BRANCH_NAME=$(echo "$BRANCH_NAME" | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | tr '_' '-')

FULL_BRANCH_NAME="${BRANCH_TYPE}/${BRANCH_NAME}"

# Determine base branch
 case $BRANCH_TYPE in
    feature|bugfix|docs|refactor)
        BASE_BRANCH="develop"
        ;;
    hotfix)
        BASE_BRANCH="main"
        ;;
    release)
        BASE_BRANCH="develop"
        ;;
    *)
        echo "${RED}Error: Unknown branch type '$BRANCH_TYPE'${NC}"
        echo "Valid types: feature, bugfix, hotfix, release, docs, refactor"
        exit 1
        ;;
esac

echo "${BLUE}Creating branch: ${FULL_BRANCH_NAME}${NC}"
echo "${BLUE}Base branch: ${BASE_BRANCH}${NC}"
echo ""

# Check if we're in a git repository
if [ ! -d "${PROJECT_ROOT}/.git" ]; then
    echo "${RED}Error: Not a git repository${NC}"
    exit 1
fi

cd "${PROJECT_ROOT}"

# Fetch latest changes
echo "${YELLOW}Fetching latest changes...${NC}"
git fetch origin

# Check if base branch exists locally or remotely
if ! git show-ref --verify --quiet refs/heads/$BASE_BRANCH; then
    if git ls-remote --heads origin $BASE_BRANCH | grep -q $BASE_BRANCH; then
        echo "${YELLOW}Creating local ${BASE_BRANCH} branch from origin/${BASE_BRANCH}${NC}"
        git checkout -b $BASE_BRANCH origin/$BASE_BRANCH
    else
        echo "${RED}Error: Base branch '${BASE_BRANCH}' not found${NC}"
        exit 1
    fi
else
    # Switch to base branch and update
    echo "${YELLOW}Switching to ${BASE_BRANCH} and updating...${NC}"
    git checkout $BASE_BRANCH
    git pull origin $BASE_BRANCH
fi

# Check if branch already exists
if git show-ref --verify --quiet refs/heads/$FULL_BRANCH_NAME; then
    echo "${YELLOW}Branch ${FULL_BRANCH_NAME} already exists locally${NC}"
    read -p "Switch to existing branch? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git checkout $FULL_BRANCH_NAME
        echo "${GREEN}✓ Switched to ${FULL_BRANCH_NAME}${NC}"
        exit 0
    else
        exit 1
    fi
fi

# Check if branch exists on remote
if git ls-remote --heads origin $FULL_BRANCH_NAME | grep -q $FULL_BRANCH_NAME; then
    echo "${YELLOW}Branch ${FULL_BRANCH_NAME} exists on remote${NC}"
    read -p "Checkout from remote? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git checkout -b $FULL_BRANCH_NAME origin/$FULL_BRANCH_NAME
        echo "${GREEN}✓ Checked out ${FULL_BRANCH_NAME} from remote${NC}"
        exit 0
    else
        exit 1
    fi
fi

# Create new branch
echo "${YELLOW}Creating new branch ${FULL_BRANCH_NAME}...${NC}"
git checkout -b $FULL_BRANCH_NAME

echo ""
echo "${GREEN}============================================${NC}"
echo "${GREEN}  Branch created successfully!${NC}"
echo "${GREEN}============================================${NC}"
echo ""
echo "Current branch: $(git rev-parse --abbrev-ref HEAD)"
echo ""
echo "Next steps:"
echo "  1. Make your changes"
echo "  2. Commit with: git commit -m '${BRANCH_TYPE}: <description>'"
echo "  3. Push with: git push origin ${FULL_BRANCH_NAME}"
echo "  4. Create a PR/MR to merge into ${BASE_BRANCH}"
echo ""
