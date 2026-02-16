# GitHub Quick Start Guide

Get started with REChain IDE development on GitHub in 5 minutes.

## 1. Fork & Clone (2 minutes)

```bash
# Fork the repository on GitHub first, then:
git clone https://github.com/YOUR_USERNAME/rechain-ide.git
cd rechain-ide

# Add upstream remote
git remote add upstream https://github.com/rechain-ai/rechain-ide.git
```

## 2. Setup (2 minutes)

```bash
# Run setup script
./scripts/setup-repo.sh

# Or manually:
make git-setup
```

## 3. Create Branch (30 seconds)

```bash
# Using the helper script
./scripts/create-branch.sh feature my-first-contribution

# Or manually:
git checkout develop
git pull upstream develop
git checkout -b feature/my-first-contribution
```

## 4. Make Changes

Edit files in your IDE. Follow our conventions:
- Go code: Run `gofmt` before committing
- Tests: Write tests for new code
- Docs: Update if needed

## 5. Commit

```bash
# Stage changes
git add .

# Commit with conventional format
git commit -m "feat: add my feature"
```

## 6. Push & PR (30 seconds)

```bash
# Push to your fork
git push origin feature/my-first-contribution

# Create PR via GitHub UI
# - Fill in the template
# - Link issues: "Closes #123"
# - Request review
```

## Available Commands

```bash
# Check PR readiness
make check-pr

# Run tests
make test

# Build project
make build

# Clean branches
make clean-branches

# Install hooks
make git-hooks
```

## Need Help?

- 📖 [First Time Contributors](./FIRST_TIME_CONTRIBUTORS.md)
- 🔧 [Environment Setup](./ENVIRONMENT_SETUP.md)
- 💬 [Discussions](https://github.com/rechain-ai/rechain-ide/discussions)
- 🎧 [Discord](https://discord.gg/rechain)

## File Structure

```
.github/
├── workflows/          # CI/CD automation
├── ISSUE_TEMPLATE/     # Issue templates
├── *.md               # Guides & docs
└── CODEOWNERS         # Review assignments

.gitlab/               # GitLab mirror config
.git/hooks/            # Git hooks (after setup)
scripts/               # Utility scripts
```

## Common Tasks

### Update your fork
```bash
git checkout develop
git pull upstream develop
git push origin develop
```

### Sync feature branch
```bash
git checkout feature/my-branch
git rebase develop
```

### Check status
```bash
make check-pr
```

## Next Steps

1. Read [CONTRIBUTING.md](../CONTRIBUTING.md)
2. Check [good first issues](https://github.com/rechain-ai/rechain-ide/labels/good%20first%20issue)
3. Join our [Discord community](https://discord.gg/rechain)

---

**Ready to contribute?** Start with an issue labeled `good first issue`! 🚀
