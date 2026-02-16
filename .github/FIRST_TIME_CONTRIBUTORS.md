# First-Time Contributors Guide

Welcome to REChain Quantum-CrossAI IDE Engine! This guide will help you make your first contribution.

## Quick Start

### 1. Fork the Repository

Click the "Fork" button on GitHub or "Fork" on GitLab to create your own copy.

### 2. Clone Your Fork

```bash
git clone https://github.com/YOUR_USERNAME/rechain-ide.git
cd rechain-ide
```

### 3. Set Up Environment

```bash
# Run the setup script
./scripts/setup-repo.sh

# This will:
# - Check required tools
# - Install git hooks
# - Configure git
# - Run initial build check
```

### 4. Create a Branch

```bash
# For bug fixes
./scripts/create-branch.sh bugfix fix-typo

# For new features
./scripts/create-branch.sh feature add-tooltip
```

## Find Something to Work On

### Good First Issues

Look for issues labeled:
- `good first issue` - Easy tasks for newcomers
- `help wanted` - Tasks where we need community help
- `documentation` - Doc improvements (great for learning)

### Where to Find Them

- GitHub Issues: https://github.com/rechain-ai/rechain-ide/issues
- Filter by label: `label:good+first+issue`

## Making Changes

### 1. Make Your Changes

Edit files in your IDE. Common areas for first contributions:
- Documentation (`docs/` or `*.md` files)
- Tests (`*_test.go` files)
- Small bug fixes

### 2. Test Your Changes

```bash
# Run tests
cd rechain-ide
go test ./...

# Check formatting
gofmt -w .

# Run linter
golangci-lint run
```

### 3. Commit Your Changes

```bash
# Stage changes
git add .

# Commit with conventional format
git commit -m "docs: fix typo in README"

# Or for features:
git commit -m "feat: add hover tooltip to buttons"
```

### 4. Check Your PR

```bash
# Run PR check script
./scripts/check-pr.sh
```

## Submitting Your Contribution

### 1. Push to Your Fork

```bash
git push origin your-branch-name
```

### 2. Create Pull Request

On GitHub/GitLab:
1. Go to your fork
2. Click "New Pull Request"
3. Fill in the template
4. Submit

### 3. What Happens Next

1. **CI runs** - Automated tests and checks
2. **Review** - Maintainers review your code
3. **Feedback** - You may get suggestions
4. **Merge** - Your PR is merged!

## Contribution Ideas

### Documentation
- Fix typos
- Improve explanations
- Add examples
- Translate docs

### Code
- Add unit tests
- Fix small bugs
- Improve error messages
- Refactor for clarity

### Other
- Improve scripts
- Add examples
- Update dependencies

## Common Questions

### "I don't know Go"

No problem! You can contribute:
- Documentation
- Tests (we'll help you)
- Issue triage
- Design/UI feedback

### "My PR was rejected"

Don't be discouraged! This happens to everyone:
1. Read the feedback carefully
2. Ask clarifying questions
3. Make suggested changes
4. Try again

### "I need help"

- Ask in the PR comments
- Join our [Discord](https://discord.gg/rechain)
- Check [Discussions](https://github.com/rechain-ai/rechain-ide/discussions)

## Development Workflow

```
┌─────────────────────────────────────────────────────────────┐
│  1. Find Issue           2. Fork & Clone                    │
│     ↓                       ↓                               │
│  3. Create Branch        4. Make Changes                    │
│     ↓                       ↓                               │
│  5. Test Locally         6. Commit & Push                   │
│     ↓                       ↓                               │
│  7. Create PR            8. Address Feedback                │
│     ↓                       ↓                               │
│  9. Merged! 🎉                                               │
└─────────────────────────────────────────────────────────────┘
```

## Code Style Tips

### Go Code
- Run `gofmt` before committing
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Keep functions small and focused
- Write clear variable names

### Commit Messages
```
Good:  "docs: fix typo in installation guide"
Good:  "test: add unit test for parser"
Bad:   "fix stuff"
Bad:   "updated code"
```

### PR Descriptions
- Explain what changed
- Explain why it changed
- Reference the issue: "Fixes #123"

## Recognition

We appreciate all contributions! You'll be:
- Listed in release notes
- Added to contributors list
- Acknowledged in commit history

## Next Steps

After your first PR:
1. Try a slightly harder issue
2. Help review other PRs
3. Answer questions in discussions
4. Become a regular contributor!

## Resources

- [Contributing Guide](../CONTRIBUTING.md)
- [PR Best Practices](./PR_BEST_PRACTICES.md)
- [Issue Reporting](./ISSUE_REPORTING.md)
- [Environment Setup](./ENVIRONMENT_SETUP.md)

## Welcome Aboard! 🚀

We're excited to have you contribute to REChain IDE. Don't hesitate to ask questions - we're here to help!
