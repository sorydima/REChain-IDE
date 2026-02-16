# Git Hooks for REChain Quantum-CrossAI IDE Engine

This directory contains sample git hooks that can be installed in your local `.git/hooks/` directory.

## Installation

To install a hook, copy it from this directory to `.git/hooks/` and remove the `.sample` extension:

```bash
# Copy a specific hook
cp .githooks/pre-commit.sample .git/hooks/pre-commit

# Make it executable (on Unix-like systems)
chmod +x .git/hooks/pre-commit
```

## Available Hooks

### pre-commit
Runs before each commit. Checks:
- Go code formatting (`gofmt`)
- Go code vetting (`go vet`)
- Trailing whitespace
- Unit tests (quick run)

### commit-msg
Validates commit message format following [Conventional Commits](https://www.conventionalcommits.org/):
- Format: `type[(scope)]: subject`
- Types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert
- Subject: max 100 characters

### pre-push
Runs before pushing to remote. Checks:
- Full test suite
- Race detector
- Security vulnerability scanning (if `govulncheck` is installed)

### pre-rebase
Warns when rebasing commits that have been pushed to origin.

### post-checkout
Runs after checkout/clone. Shows:
- Project setup hints
- Current branch and commit info

### post-merge
Runs after merge. Checks for:
- Changed Go dependencies and runs `go work sync` or `go mod download`
- Changed Node.js dependencies and prompts for `npm install`

## Customization

These hooks are templates. Feel free to modify them to suit your workflow:
- Adjust which tests run in `pre-commit` vs `pre-push`
- Add custom validation rules in `commit-msg`
- Include additional security checks

## Bypassing Hooks

In case of emergency, you can bypass hooks:
- Skip pre-commit: `git commit --no-verify` (or `-n`)
- Skip pre-push: `git push --no-verify` (or `-n`)

## Additional Resources

- [Git Hooks Documentation](https://git-scm.com/docs/githooks)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
