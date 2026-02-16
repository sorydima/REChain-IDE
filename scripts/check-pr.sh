#!/bin/bash
# check-pr.sh - Check if a PR/MR is ready for review
# Usage: ./scripts/check-pr.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

ERRORS=0
WARNINGS=0

echo "${BLUE}============================================${NC}"
echo "${BLUE}  PR Readiness Check${NC}"
echo "${BLUE}============================================${NC}"
echo ""

# Check if we're in a git repository
if [ ! -d "${PROJECT_ROOT}/.git" ]; then
    echo "${RED}Error: Not a git repository${NC}"
    exit 1
fi

# Check 1: Branch naming convention
echo "${YELLOW}1. Checking branch name...${NC}"
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [[ $CURRENT_BRANCH =~ ^(feature|bugfix|hotfix|release)\/.+ ]] || [[ $CURRENT_BRANCH == "main" ]] || [[ $CURRENT_BRANCH == "develop" ]]; then
    echo "${GREEN}✓ Branch name follows convention: $CURRENT_BRANCH${NC}"
else
    echo "${YELLOW}⚠ Branch name may not follow convention: $CURRENT_BRANCH${NC}"
    echo "   Expected: feature/*, bugfix/*, hotfix/*, release/*"
    ((WARNINGS++))
fi

# Check 2: Commit message format
echo ""
echo "${YELLOW}2. Checking recent commit messages...${NC}"
COMMITS=$(git log origin/main..HEAD --oneline --no-merges 2>/dev/null || git log --oneline -10)
if [ -n "$COMMITS" ]; then
    echo "Recent commits:"
    echo "$COMMITS" | head -5 | while read line; do
        echo "  - $line"
    done
    
    # Check conventional commits format
    INVALID_COMMITS=$(git log origin/main..HEAD --format="%s" --no-merges 2>/dev/null | grep -vE "^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\(.+\))?: .+" || true)
    if [ -z "$INVALID_COMMITS" ]; then
        echo "${GREEN}✓ Commit messages follow conventional format${NC}"
    else
        echo "${YELLOW}⚠ Some commits may not follow conventional format${NC}"
        ((WARNINGS++))
    fi
else
    echo "${YELLOW}⚠ No new commits found${NC}"
fi

# Check 3: Tests pass
echo ""
echo "${YELLOW}3. Running tests...${NC}"
if [ -d "${PROJECT_ROOT}/rechain-ide" ]; then
    cd "${PROJECT_ROOT}/rechain-ide"
    if go test -short ./... > /tmp/test-output.txt 2>&1; then
        echo "${GREEN}✓ Tests pass${NC}"
    else
        echo "${RED}✗ Tests failed${NC}"
        tail -20 /tmp/test-output.txt
        ((ERRORS++))
    fi
else
    echo "${YELLOW}⚠ rechain-ide directory not found${NC}"
fi

# Check 4: Code formatting
echo ""
echo "${YELLOW}4. Checking code formatting...${NC}"
if [ -d "${PROJECT_ROOT}/rechain-ide" ]; then
    cd "${PROJECT_ROOT}/rechain-ide"
    UNFORMATTED=$(gofmt -l . 2>/dev/null || true)
    if [ -z "$UNFORMATTED" ]; then
        echo "${GREEN}✓ Code is properly formatted${NC}"
    else
        echo "${RED}✗ Code needs formatting:${NC}"
        echo "$UNFORMATTED" | head -10 | while read file; do
            echo "  - $file"
        done
        ((ERRORS++))
    fi
fi

# Check 5: No trailing whitespace
echo ""
echo "${YELLOW}5. Checking for trailing whitespace...${NC}"
if git diff --check --cached 2>/dev/null || git diff --check 2>/dev/null; then
    echo "${GREEN}✓ No trailing whitespace issues${NC}"
else
    echo "${RED}✗ Trailing whitespace found${NC}"
    ((ERRORS++))
fi

# Check 6: Documentation updates
echo ""
echo "${YELLOW}6. Checking for documentation updates...${NC}"
DOC_CHANGES=$(git diff --name-only origin/main 2>/dev/null | grep -E "\.md$|docs/" || true)
CODE_CHANGES=$(git diff --name-only origin/main 2>/dev/null | grep -E "\.go$|\.ts$" || true)
if [ -n "$CODE_CHANGES" ] && [ -z "$DOC_CHANGES" ]; then
    echo "${YELLOW}⚠ Code changes detected without documentation updates${NC}"
    echo "   Consider updating README.md or other docs if needed"
    ((WARNINGS++))
else
    echo "${GREEN}✓ Documentation status OK${NC}"
fi

# Summary
echo ""
echo "${BLUE}============================================${NC}"
echo "${BLUE}  Check Summary${NC}"
echo "${BLUE}============================================${NC}"
if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo "${GREEN}✓ All checks passed! PR is ready for review.${NC}"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    echo "${YELLOW}⚠ Checks completed with $WARNINGS warning(s).${NC}"
    echo "   Review warnings before submitting PR."
    exit 0
else
    echo "${RED}✗ Checks failed with $ERRORS error(s) and $WARNINGS warning(s).${NC}"
    echo "   Please fix errors before submitting PR."
    exit 1
fi
