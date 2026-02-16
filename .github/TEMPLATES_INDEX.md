# Repository Templates Index

This document serves as an index for all templates available in the REChain Quantum-CrossAI IDE Engine project.

## Issue Templates

### GitHub

Located in: `.github/ISSUE_TEMPLATE/`

| Template | Purpose | File |
|----------|---------|------|
| Bug Report | Report bugs and issues | `bug_report.md` |
| Feature Request | Request new features | `feature_request.md` |
| Config | Issue template configuration | `config.yml` |

### GitLab

Located in: `.gitlab/issue_templates/`

| Template | Purpose | File |
|----------|---------|------|
| Bug Report | Report bugs and issues | `bug_report.md` |
| Feature Request | Request new features | `feature_request.md` |
| Task | General tasks and maintenance | `task.md` |
| Config | Template configuration | `config.yml` |

## Merge/Pull Request Templates

### GitHub

Located in: `.github/`

| Template | Purpose | File |
|----------|---------|------|
| Default | Standard PR template | `pull_request_template.md` |

### GitLab

Located in: `.gitlab/merge_request_templates/`

| Template | Purpose | File |
|----------|---------|------|
| Default | Standard MR template | `Default.md` |

## CI/CD Templates

### GitHub Actions

Located in: `.github/workflows/`

| Workflow | Purpose |
|----------|---------|
| ci.yml | Continuous integration |
| automated-testing.yml | Test suite |
| security-scan.yml | Security checks |
| code-analysis.yml | Static analysis |
| dependency-update.yml | Dependency updates |
| release.yml | Release automation |
| changelog.yml | Changelog generation |
| project-automation.yml | Project management |
| stale.yml | Stale issue management |
| And more... | See `.github/workflows/` |

### GitLab CI

Located in: `.gitlab/ci/`

| Template | Purpose | File |
|----------|---------|------|
| Common | Base configuration | `common.gitlab-ci.yml` |
| Build | Build jobs | `build.gitlab-ci.yml` |
| Test | Test jobs | `test.gitlab-ci.yml` |
| Security | Security scanning | `security.gitlab-ci.yml` |
| Deploy | Deployment | `deploy.gitlab-ci.yml` |
| Docs | Documentation | `docs.gitlab-ci.yml` |
| Release | Release management | `release.gitlab-ci.yml` |
| Pages | GitLab Pages | `pages.gitlab-ci.yml` |
| Container Scanning | Container security | `container-scanning.gitlab-ci.yml` |
| Auto DevOps | Auto DevOps template | `auto-devops.gitlab-ci.yml` |
| And more... | See `.gitlab/ci/` |

## Git Hooks

Located in: `.githooks/`

| Hook | Purpose | File |
|------|---------|------|
| Pre-commit | Pre-commit validation | `pre-commit.sample` |
| Commit-msg | Commit message validation | `commit-msg.sample` |
| Pre-push | Pre-push testing | `pre-push.sample` |
| Pre-rebase | Rebase warnings | `pre-rebase.sample` |
| Post-checkout | Post-checkout info | `post-checkout.sample` |
| Post-merge | Post-merge dependency check | `post-merge.sample` |

## Documentation Templates

### Guides

Located in: `.github/`

| Guide | Purpose | File |
|-------|---------|------|
| Repository Setup | Git config overview | `REPOSITORY_SETUP.md` |
| Environment Setup | Dev environment setup | `ENVIRONMENT_SETUP.md` |
| PR Best Practices | PR guidelines | `PR_BEST_PRACTICES.md` |
| Issue Reporting | Issue guidelines | `ISSUE_REPORTING.md` |
| Workflow Diagrams | Visual git workflows | `WORKFLOW_DIAGRAMS.md` |
| Release Process | Release procedures | `RELEASE_PROCESS.md` |
| CI/CD Troubleshooting | CI/CD debugging | `CI_CD_TROUBLESHOOTING.md` |
| Git Configuration | Git usage guide | `GIT_CONFIGURATION.md` |

## Configuration Templates

### Repository Config

| Config | Purpose | File |
|--------|---------|------|
| Git Attributes | Line endings & linguist | `.gitattributes` |
| Git Ignore | Ignore patterns | `.gitignore` |
| Editor Config | Editor settings | `.editorconfig` |
| Commit Lint | Commit message rules | `.commitlintrc.json` |
| Semantic Release | Release automation | `.releaserc.json` |

### GitHub Config

| Config | Purpose | File |
|--------|---------|------|
| Code Owners | Review assignments | `CODEOWNERS` |
| Labels | Issue/PR labels | `labels.yml` |
| Settings | Repo settings | `settings.yml` |
| Project | Project metadata | `project.yml` |
| Dependency Review | Dependency checks | `dependency-review.yml` |
| Funding | Sponsorship | `FUNDING.yml` |

### GitLab Config

| Config | Purpose | File |
|--------|---------|------|
| Code Owners | Review assignments | `CODEOWNERS` |
| Project Settings | Project configuration | `project-settings.yml` |
| Issue Board | Board configuration | `issue_board.yml` |

## Utility Scripts

Located in: `scripts/`

| Script | Purpose | File |
|--------|---------|------|
| Install Hooks | Install git hooks | `install-hooks.sh` |
| Setup Repo | Initial repo setup | `setup-repo.sh` |
| Check PR | PR readiness check | `check-pr.sh` |
| Create Branch | Create feature branch | `create-branch.sh` |
| Bump Version | Version bumping | `bump-version.sh` |

## Using Templates

### GitHub

Issues and PRs automatically use templates when created through the web interface.

### GitLab

Use templates by selecting from the dropdown when creating issues/MRs, or use quick actions:
```
/template_link "bug_report.md"
```

### Customizing Templates

1. Edit the template files
2. Test with a dummy issue/PR
3. Commit changes
4. Templates are live immediately

## Adding New Templates

To add a new template:

1. Create file in appropriate directory
2. Follow naming convention: `name.md` or `name.yml`
3. Test thoroughly
4. Update this index
5. Document usage

See existing templates for examples.
