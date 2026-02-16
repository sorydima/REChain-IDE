# Complete Git & Repository Setup - Master Reference

This is the master reference document for all Git, GitHub, and GitLab configuration in the REChain Quantum-CrossAI IDE Engine project.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Directory Structure](#directory-structure)
3. [Configuration Files](#configuration-files)
4. [Documentation Index](#documentation-index)
5. [Scripts Reference](#scripts-reference)
6. [Workflows & CI/CD](#workflows--cicd)
7. [Troubleshooting](#troubleshooting)

---

## Quick Start

### New Contributor (5 minutes)

```bash
# 1. Fork & Clone
git clone https://github.com/YOUR_USERNAME/rechain-ide.git
cd rechain-ide

# 2. Setup
cd rechain-ide
./scripts/setup-repo.sh

# 3. Create branch
./scripts/create-branch.sh feature my-first-contribution

# 4. Make changes & commit
git add .
git commit -m "feat: add my feature"

# 5. Check & push
make check-pr
git push origin feature/my-first-contribution

# 6. Create PR via GitHub UI
```

See [.github/QUICK_START.md](.github/QUICK_START.md) for details.

---

## Directory Structure

```
.
├── .gitattributes              # Git file handling
├── .gitignore                  # Ignore patterns
├── .git-blame-ignore-revs      # Blame exclusions
├── .commitlintrc.json          # Commit message linting
├── .releaserc.json             # Semantic release config
├── .editorconfig               # Editor settings
│
├── .githooks/                  # Git hook templates
│   ├── pre-commit.sample
│   ├── commit-msg.sample
│   ├── pre-push.sample
│   ├── pre-rebase.sample
│   ├── post-checkout.sample
│   ├── post-merge.sample
│   └── README.md
│
├── .github/                    # GitHub configuration
│   ├── workflows/              # 39 CI/CD workflows
│   ├── ISSUE_TEMPLATE/         # Issue templates
│   ├── DISCUSSION_TEMPLATE/    # Discussion templates
│   ├── CODEOWNERS              # Review assignments
│   ├── labels.yml              # Label definitions
│   ├── settings.yml            # Repo settings
│   ├── project.yml             # Project metadata
│   ├── FUNDING.yml             # Sponsorship info
│   ├── SECURITY.md             # Security policy
│   ├── dependabot.yml          # Dependabot config
│   ├── dependency-review.yml   # Dependency review
│   ├── release.yml             # Release config
│   ├── stale.yml               # Stale bot config
│   ├── auto-assign.yml         # Auto-assignment
│   ├── git-aliases.txt         # Git aliases
│   └── *.md (20+ docs)         # Documentation
│
├── .gitlab/                    # GitLab configuration
│   ├── ci/                     # 35+ CI templates
│   ├── issue_templates/        # Issue templates
│   ├── merge_request_templates/ # MR templates
│   ├── CODEOWNERS              # Review assignments
│   ├── SECURITY.md             # Security policy
│   ├── project-settings.yml    # Project config
│   ├── issue_board.yml         # Issue board
│   ├── config.yml              # Template config
│   ├── release-template.md     # Release template
│   ├── CONTRIBUTING.md         # GitLab guide
│   ├── WORKFLOW_GUIDE.md       # GitLab workflow
│   └── README.md               # Directory guide
│
├── scripts/                    # Utility scripts
│   ├── install-hooks.sh        # Install git hooks
│   ├── setup-repo.sh           # Repo setup
│   ├── check-pr.sh             # PR readiness check
│   ├── create-branch.sh        # Create branch
│   ├── bump-version.sh         # Version bumping
│   └── *.ps1 (10 files)        # PowerShell scripts
│
├── docs/                       # Documentation
│   ├── GIT_CONFIGURATION_INDEX.md
│   ├── PLATFORM_COMPARISON.md
│   └── *.md (80+ files)        # Project docs
│
├── Makefile                    # Build automation
├── .gitlab-ci.yml              # GitLab CI entry point
└── README.md                   # Main project readme
```

---

## Configuration Files

### Root Level

| File | Purpose | Documentation |
|------|---------|---------------|
| `.gitattributes` | Line endings, binary handling | Git docs |
| `.gitignore` | Ignore patterns | Git docs |
| `.commitlintrc.json` | Conventional commits | [commitlint](https://commitlint.js.org/) |
| `.releaserc.json` | Semantic release | [semantic-release](https://semantic-release.gitbook.io/) |
| `.editorconfig` | Editor consistency | [EditorConfig](https://editorconfig.org/) |

### Git Hooks

Install with: `make git-hooks` or `./scripts/install-hooks.sh`

| Hook | Purpose |
|------|---------|
| pre-commit | Format, vet, test checks |
| commit-msg | Conventional commit validation |
| pre-push | Full test suite |
| pre-rebase | Warn on rebasing pushed commits |
| post-checkout | Post-checkout info |
| post-merge | Dependency update check |

---

## Documentation Index

### Getting Started

| Document | Purpose |
|----------|---------|
| [.github/FIRST_TIME_CONTRIBUTORS.md](.github/FIRST_TIME_CONTRIBUTORS.md) | New contributor guide |
| [.github/QUICK_START.md](.github/QUICK_START.md) | 5-minute quick start |
| [.github/ENVIRONMENT_SETUP.md](.github/ENVIRONMENT_SETUP.md) | Dev environment setup |
| [.github/GIT_CONFIGURATION.md](.github/GIT_CONFIGURATION.md) | Git usage guide |

### Development

| Document | Purpose |
|----------|---------|
| [.github/DEVELOPMENT_WORKFLOW.md](.github/DEVELOPMENT_WORKFLOW.md) | Daily workflow |
| [.github/PR_BEST_PRACTICES.md](.github/PR_BEST_PRACTICES.md) | PR guidelines |
| [.github/ISSUE_REPORTING.md](.github/ISSUE_REPORTING.md) | Issue guidelines |
| [.github/REVIEWING_CHECKLIST.md](.github/REVIEWING_CHECKLIST.md) | Code review guide |
| [.github/WORKFLOW_DIAGRAMS.md](.github/WORKFLOW_DIAGRAMS.md) | Visual workflows |
| [.github/WORKFLOW_MERMAID.md](.github/WORKFLOW_MERMAID.md) | Mermaid diagrams |

### Operations

| Document | Purpose |
|----------|---------|
| [.github/RELEASE_PROCESS.md](.github/RELEASE_PROCESS.md) | Release procedures |
| [.github/CI_CD_TROUBLESHOOTING.md](.github/CI_CD_TROUBLESHOOTING.md) | CI/CD debugging |
| [.github/SECURITY_CHECKLIST.md](.github/SECURITY_CHECKLIST.md) | Security checklist |
| [.github/GIT_FAQ.md](.github/GIT_FAQ.md) | Common Git questions |

### Platform Guides

| Document | Purpose |
|----------|---------|
| [.github/README.md](.github/README.md) | GitHub config guide |
| [.gitlab/README.md](.gitlab/README.md) | GitLab config guide |
| [.gitlab/CONTRIBUTING.md](.gitlab/CONTRIBUTING.md) | GitLab contributions |
| [.gitlab/WORKFLOW_GUIDE.md](.gitlab/WORKFLOW_GUIDE.md) | GitLab workflow |
| [docs/PLATFORM_COMPARISON.md](docs/PLATFORM_COMPARISON.md) | Platform comparison |

### Reference

| Document | Purpose |
|----------|---------|
| [.github/TEMPLATES_INDEX.md](.github/TEMPLATES_INDEX.md) | All templates index |
| [.github/REPOSITORY_SETUP.md](.github/REPOSITORY_SETUP.md) | Repository overview |
| [.github/MIRROR_CONFIGURATION.md](.github/MIRROR_CONFIGURATION.md) | GitHub/GitLab sync |
| [.github/API_DOCUMENTATION.md](.github/API_DOCUMENTATION.md) | API docs guide |
| [docs/GIT_CONFIGURATION_INDEX.md](docs/GIT_CONFIGURATION_INDEX.md) | Complete index |

---

## Scripts Reference

### Bash Scripts (Unix/Linux/macOS)

| Script | Purpose | Usage |
|--------|---------|-------|
| `install-hooks.sh` | Install git hooks | `./install-hooks.sh` |
| `setup-repo.sh` | Initial setup | `./setup-repo.sh` |
| `check-pr.sh` | PR readiness | `./check-pr.sh` |
| `create-branch.sh` | Create branch | `./create-branch.sh <type> <name>` |
| `bump-version.sh` | Bump version | `./bump-version.sh <level>` |

### Make Targets

| Target | Purpose |
|--------|---------|
| `make git-setup` | Setup git hooks |
| `make git-hooks` | Install hooks |
| `make check-pr` | Check PR readiness |
| `make create-branch` | Create feature branch |
| `make bump-version` | Bump version |
| `make clean-branches` | Clean merged branches |
| `make changelog` | Generate changelog |

---

## Workflows & CI/CD

### GitHub Actions (`.github/workflows/`)

Categories: 39 workflows total
- **Core**: CI, testing, building
- **Security**: Scanning, vulnerability checks
- **Automation**: Project management, stale issues
- **Quality**: Linting, formatting, analysis
- **Release**: Changelog, release automation

### GitLab CI (`.gitlab/ci/`)

Categories: 35+ templates
- **Build**: Compilation, Docker
- **Test**: Unit, integration, benchmark
- **Security**: SAST, dependency scanning
- **Deploy**: Staging, production
- **Release**: Version management

---

## Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| Pre-commit hooks not running | Run `make git-hooks` |
| CI failing | Check [.github/CI_CD_TROUBLESHOOTING.md](.github/CI_CD_TROUBLESHOOTING.md) |
| Git questions | See [.github/GIT_FAQ.md](.github/GIT_FAQ.md) |
| Environment issues | See [.github/ENVIRONMENT_SETUP.md](.github/ENVIRONMENT_SETUP.md) |

### Support

- 📖 Documentation: Start with this file
- 💬 Discussions: [GitHub Discussions](https://github.com/rechain-ai/rechain-ide/discussions)
- 🎧 Discord: [Join](https://discord.gg/rechain)
- 📧 Email: dev@rechain.ai

---

## Statistics

| Category | Count |
|----------|-------|
| GitHub Workflows | 39 |
| GitLab CI Templates | 35+ |
| Git Hooks | 6 |
| Bash Scripts | 5 |
| PowerShell Scripts | 10 |
| Documentation Files | 30+ |
| Issue Templates | 8 |
| MR/PR Templates | 2 |

---

## Maintenance

### Regular Tasks

- **Weekly**: Review and merge PRs
- **Monthly**: Update dependencies
- **Quarterly**: Review documentation
- **As needed**: Add new templates

### Adding New Files

1. Create in appropriate directory
2. Follow naming conventions
3. Update relevant indexes
4. Test thoroughly
5. Document in this guide

---

## License

All configuration files are part of the REChain Quantum-CrossAI IDE Engine project and follow the same license terms.

---

*Last Updated: February 2026*  
*Maintained by: REChain Core Team*
