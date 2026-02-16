# Git Configuration Guide

This document outlines the Git configuration and workflow for the REChain Quantum-CrossAI IDE Engine project.

## Repository Structure

```
.
├── .gitattributes          # Git attributes for line endings and file handling
├── .gitignore              # Files and directories to ignore
├── .git-blame-ignore-revs  # Revisions to ignore in git blame
├── .github/                # GitHub-specific configurations
│   ├── workflows/          # GitHub Actions workflows
│   ├── ISSUE_TEMPLATE/     # Issue templates
│   ├── CODEOWNERS          # Code review assignments
│   ├── FUNDING.yml         # Sponsorship information
│   ├── SECURITY.md         # Security policy
│   ├── settings.yml        # Repository settings
│   └── ...
├── .gitlab/                # GitLab-specific configurations
│   ├── issue_templates/    # Issue templates
│   ├── merge_request_templates/  # MR templates
│   ├── ci/                 # CI/CD templates
│   ├── issue_board.yml     # Issue board configuration
│   └── SECURITY.md         # Security policy
└── .githooks/              # Sample git hooks
    ├── pre-commit.sample
    ├── commit-msg.sample
    ├── pre-push.sample
    └── README.md
```

## Setup

### Initial Clone

```bash
git clone https://github.com/rechain/quantum-crossai-ide.git
cd quantum-crossai-ide
```

### Install Git Hooks (Optional)

To enable pre-commit and pre-push checks:

```bash
# On Unix-like systems
cp .githooks/pre-commit.sample .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

cp .githooks/commit-msg.sample .git/hooks/commit-msg
chmod +x .git/hooks/commit-msg

# On Windows (PowerShell)
Copy-Item .githooks\pre-commit.sample .git\hooks\pre-commit
Copy-Item .githooks\commit-msg.sample .git\hooks\commit-msg
```

## Commit Message Convention

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[(optional scope)]: <description>

[optional body]

[optional footer(s)]
```

### Types

- **feat**: A new feature
- **fix**: A bug fix
- **docs**: Documentation only changes
- **style**: Changes that do not affect the meaning of the code (formatting, etc.)
- **refactor**: Code change that neither fixes a bug nor adds a feature
- **perf**: Performance improvement
- **test**: Adding or correcting tests
- **build**: Changes to build system or dependencies
- **ci**: Changes to CI configuration
- **chore**: Other changes that don't modify src or test files
- **revert**: Reverts a previous commit

### Examples

```
feat(auth): add OAuth2 login support

fix(compiler): resolve type inference issue

docs(api): update REST endpoint documentation

ci: add GitLab CI/CD pipeline
test(lexer): add benchmark tests
```

## Branching Strategy

We use a modified Git Flow approach:

- **main**: Production-ready code
- **develop**: Integration branch for features
- **feature/***: New features (branch from develop)
- **bugfix/***: Bug fixes (branch from develop)
- **hotfix/***: Urgent production fixes (branch from main)
- **release/***: Release preparation (branch from develop)

### Workflow

1. Create a feature branch:
   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b feature/my-new-feature
   ```

2. Make changes and commit:
   ```bash
   git add .
   git commit -m "feat(scope): description"
   ```

3. Push and create PR/MR:
   ```bash
   git push origin feature/my-new-feature
   ```

4. After review, merge to develop
5. When ready for release, merge develop to main

## Code Review Process

- All changes must go through Pull Request (GitHub) or Merge Request (GitLab)
- Minimum 1 review required for non-trivial changes
- CI must pass before merging
- Follow the PR/MR templates provided

## CI/CD Integration

### GitHub Actions

Workflows are defined in `.github/workflows/`:
- **ci.yml**: Basic CI checks
- **automated-testing.yml**: Test suite
- **code-analysis.yml**: Static analysis
- **security-scan.yml**: Security checks
- **release.yml**: Release automation

### GitLab CI

Pipeline defined in `.gitlab-ci.yml` with includes from `.gitlab/ci/`:
- **build.gitlab-ci.yml**: Build jobs
- **test.gitlab-ci.yml**: Test jobs
- **security.gitlab-ci.yml**: Security scanning
- **deploy.gitlab-ci.yml**: Deployment
- **release.gitlab-ci.yml**: Release management

## Git Attributes

Key configurations in `.gitattributes`:
- Automatic line ending normalization (`* text=auto`)
- Binary file handling for images and archives
- Linguist settings for language detection
- Go module file handling

## Useful Commands

```bash
# See what will be committed
git diff --cached

# Check for trailing whitespace
git diff --check

# View commit history with graph
git log --oneline --graph --all

# Find commits by message
git log --all --oneline --grep="feat(auth)"

# Clean untracked files (dry run first)
git clean -n
git clean -fd

# Stash changes
git stash push -m "description"
git stash pop

# Amend last commit
git commit --amend
```

## Troubleshooting

### Line Ending Issues

If you encounter line ending issues:
```bash
git config core.autocrlf input  # Unix/Mac
git config core.autocrlf true   # Windows
git add --renormalize .
```

### Large Files

We use Git LFS for large files (if enabled). Check `.gitattributes` for LFS patterns.

### Merge Conflicts

```bash
# During rebase/merge, see conflicts
git status

# Mark resolved
git add <resolved-file>

# Continue
git rebase --continue
# or
git merge --continue
```

## References

- [Git Documentation](https://git-scm.com/doc)
- [GitHub Docs](https://docs.github.com/)
- [GitLab Docs](https://docs.gitlab.com/)
- [Conventional Commits](https://www.conventionalcommits.org/)
