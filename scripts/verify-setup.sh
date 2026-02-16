#!/bin/bash
# verify-setup.sh - Verify the Git/GitHub/GitLab configuration is complete
# Usage: ./scripts/verify-setup.sh

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "${BLUE}============================================${NC}"
echo "${BLUE}  REChain IDE - Setup Verification${NC}"
echo "${BLUE}============================================${NC}"
echo ""

ERRORS=0
WARNINGS=0

# Check root configuration files
echo "${YELLOW}Checking root configuration files...${NC}"

ROOT_FILES=(
    ".gitattributes"
    ".gitignore"
    ".commitlintrc.json"
    ".releaserc.json"
    ".editorconfig"
)

for file in "${ROOT_FILES[@]}"; do
    if [ -f "$file" ]; then
        echo "${GREEN}✓ $file${NC}"
    else
        echo "${RED}✗ $file missing${NC}"
        ((ERRORS++))
    fi
done

# Check directories
echo ""
echo "${YELLOW}Checking directories...${NC}"

DIRS=(
    ".githooks"
    ".github"
    ".github/workflows"
    ".github/ISSUE_TEMPLATE"
    ".github/DISCUSSION_TEMPLATE"
    ".gitlab"
    ".gitlab/ci"
    ".gitlab/issue_templates"
    ".gitlab/merge_request_templates"
    "scripts"
)

for dir in "${DIRS[@]}"; do
    if [ -d "$dir" ]; then
        count=$(find "$dir" -type f | wc -l)
        echo "${GREEN}✓ $dir ($count files)${NC}"
    else
        echo "${RED}✗ $dir missing${NC}"
        ((ERRORS++))
    fi
done

# Check GitHub workflows
echo ""
echo "${YELLOW}Checking GitHub workflows...${NC}"

if [ -d ".github/workflows" ]; then
    workflow_count=$(find .github/workflows -name "*.yml" | wc -l)
    if [ "$workflow_count" -gt 30 ]; then
        echo "${GREEN}✓ $workflow_count workflows found${NC}"
    else
        echo "${YELLOW}⚠ Only $workflow_count workflows (expected 30+)${NC}"
        ((WARNINGS++))
    fi
fi

# Check GitLab CI templates
echo ""
echo "${YELLOW}Checking GitLab CI templates...${NC}"

if [ -d ".gitlab/ci" ]; then
    ci_count=$(find .gitlab/ci -name "*.yml" | wc -l)
    if [ "$ci_count" -gt 25 ]; then
        echo "${GREEN}✓ $ci_count CI templates found${NC}"
    else
        echo "${YELLOW}⚠ Only $ci_count CI templates (expected 25+)${NC}"
        ((WARNINGS++))
    fi
fi

# Check git hooks
echo ""
echo "${YELLOW}Checking git hooks...${NC}"

HOOKS=(
    "pre-commit.sample"
    "commit-msg.sample"
    "pre-push.sample"
    "pre-rebase.sample"
    "post-checkout.sample"
    "post-merge.sample"
)

for hook in "${HOOKS[@]}"; do
    if [ -f ".githooks/$hook" ]; then
        echo "${GREEN}✓ $hook${NC}"
    else
        echo "${YELLOW}⚠ $hook missing${NC}"
        ((WARNINGS++))
    fi
done

# Check documentation
echo ""
echo "${YELLOW}Checking documentation...${NC}"

DOC_FILES=(
    ".github/README.md"
    ".github/GIT_CONFIGURATION.md"
    ".github/REPOSITORY_SETUP.md"
    ".github/CONTRIBUTING.md"
    ".github/DEVELOPMENT_WORKFLOW.md"
    ".github/PR_BEST_PRACTICES.md"
    ".github/FIRST_TIME_CONTRIBUTORS.md"
    ".github/ISSUE_REPORTING.md"
    ".github/REVIEWING_CHECKLIST.md"
    ".gitlab/README.md"
    ".gitlab/CONTRIBUTING.md"
    ".gitlab/WORKFLOW_GUIDE.md"
)

for doc in "${DOC_FILES[@]}"; do
    if [ -f "$doc" ]; then
        echo "${GREEN}✓ $doc${NC}"
    else
        echo "${YELLOW}⚠ $doc missing${NC}"
        ((WARNINGS++))
    fi
done

# Check scripts
echo ""
echo "${YELLOW}Checking utility scripts...${NC}"

SCRIPTS=(
    "scripts/install-hooks.sh"
    "scripts/setup-repo.sh"
    "scripts/check-pr.sh"
    "scripts/create-branch.sh"
    "scripts/bump-version.sh"
)

for script in "${SCRIPTS[@]}"; do
    if [ -f "$script" ]; then
        if [ -x "$script" ]; then
            echo "${GREEN}✓ $script (executable)${NC}"
        else
            echo "${YELLOW}⚠ $script (not executable)${NC}"
            ((WARNINGS++))
        fi
    else
        echo "${YELLOW}⚠ $script missing${NC}"
        ((WARNINGS++))
    fi
done

# Check Make targets
echo ""
echo "${YELLOW}Checking Makefile targets...${NC}"

if [ -f "Makefile" ]; then
    TARGETS=("git-setup" "git-hooks" "check-pr" "create-branch" "bump-version")
    for target in "${TARGETS[@]}"; do
        if grep -q "^$target:" Makefile; then
            echo "${GREEN}✓ make $target${NC}"
        else
            echo "${YELLOW}⚠ make $target not found${NC}"
            ((WARNINGS++))
        fi
    done
else
    echo "${YELLOW}⚠ Makefile not found${NC}"
    ((WARNINGS++))
fi

# Check configuration files
echo ""
echo "${YELLOW}Checking configuration validity...${NC}"

# Check .gitattributes
if [ -f ".gitattributes" ]; then
    if grep -q "linguist-language" .gitattributes; then
        echo "${GREEN}✓ .gitattributes has linguist settings${NC}"
    else
        echo "${YELLOW}⚠ .gitattributes missing linguist settings${NC}"
        ((WARNINGS++))
    fi
fi

# Check .gitignore
if [ -f ".gitignore" ]; then
    if grep -q "node_modules" .gitignore && grep -q "vendor" .gitignore; then
        echo "${GREEN}✓ .gitignore has standard patterns${NC}"
    else
        echo "${YELLOW}⚠ .gitignore missing standard patterns${NC}"
        ((WARNINGS++))
    fi
fi

# Summary
echo ""
echo "${BLUE}============================================${NC}"
echo "${BLUE}  Verification Summary${NC}"
echo "${BLUE}============================================${NC}"

if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo "${GREEN}✓ All checks passed! Setup is complete.${NC}"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    echo "${YELLOW}⚠ $WARNINGS warning(s). Setup is functional but incomplete.${NC}"
    exit 0
else
    echo "${RED}✗ $ERRORS error(s) and $WARNINGS warning(s). Setup incomplete.${NC}"
    exit 1
fi
