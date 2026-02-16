# Git & Repository Configuration Summary

This document provides an overview of all Git, GitHub, and GitLab configuration files for the REChain Quantum-CrossAI IDE Engine project.

## Git Configuration Files

### Core Git Files
| File | Purpose |
|------|---------|
| `.gitattributes` | Defines line ending handling, binary files, and linguist settings |
| `.gitignore` | Specifies files and directories to ignore |
| `.git-blame-ignore-revs` | Lists commits to ignore in git blame |
| `.gitmodules` | Configures Git submodules |
| `.mailmap` | Maps authors for consistent attribution |

### Commit & Release Management
| File | Purpose |
|------|---------|
| `.commitlintrc.json` | Commit message linting rules (Conventional Commits) |
| `.releaserc.json` | Semantic release configuration |
| `.editorconfig` | Editor settings for consistent formatting |

## GitHub Configuration (.github/)

### Issue & PR Templates
| File | Purpose |
|------|---------|
| `ISSUE_TEMPLATE/bug_report.md` | Bug report template |
| `ISSUE_TEMPLATE/feature_request.md` | Feature request template |
| `ISSUE_TEMPLATE/config.yml` | Issue template configuration |
| `DISCUSSION_TEMPLATE/categories.yml` | Discussion categories |
| `pull_request_template.md` | Pull request template |

### Repository Management
| File | Purpose |
|------|---------|
| `CODEOWNERS` | Code review ownership rules |
| `labels.yml` | Label definitions for issues/PRs |
| `settings.yml` | Repository settings |
| `project.yml` | Project metadata |
| `FUNDING.yml` | Sponsorship information |
| `SECURITY.md` | Security policy |

### Workflow Documentation
| File | Purpose |
|------|---------|
| `GIT_CONFIGURATION.md` | Git usage guide |
| `dependency-review.yml` | Dependency review configuration |

### GitHub Actions Workflows (workflows/)
| Workflow | Purpose |
|----------|---------|
| `ci.yml` | Continuous integration |
| `automated-testing.yml` | Automated testing |
| `code-analysis.yml` | Static code analysis |
| `security-scan.yml` | Security scanning |
| `dependency-update.yml` | Dependency updates |
| `release.yml` | Release automation |
| `changelog.yml` | Changelog generation |
| And 20+ more... |

## GitLab Configuration (.gitlab/)

### Issue & MR Templates
| File | Purpose |
|------|---------|
| `issue_templates/bug_report.md` | Bug report template |
| `issue_templates/feature_request.md` | Feature request template |
| `issue_templates/task.md` | Task template |
| `issue_templates/config.yml` | Template configuration |
| `merge_request_templates/Default.md` | MR template |

### Repository Management
| File | Purpose |
|------|---------|
| `CODEOWNERS` | Code review ownership rules |
| `SECURITY.md` | Security policy |
| `issue_board.yml` | Issue board configuration |
| `config.yml` | GitLab configuration |

### CI/CD Templates (ci/)
| File | Purpose |
|------|---------|
| `common.gitlab-ci.yml` | Common CI configuration |
| `build.gitlab-ci.yml` | Build jobs |
| `test.gitlab-ci.yml` | Test jobs |
| `security.gitlab-ci.yml` | Security scanning |
| `deploy.gitlab-ci.yml` | Deployment jobs |
| `docs.gitlab-ci.yml` | Documentation jobs |
| `release.gitlab-ci.yml` | Release jobs |
| And 20+ more existing templates... |

### Main CI File
| File | Purpose |
|------|---------|
| `.gitlab-ci.yml` | Main GitLab CI/CD pipeline |

## Git Hooks (.githooks/)

| Hook | Purpose |
|------|---------|
| `pre-commit.sample` | Pre-commit validation |
| `commit-msg.sample` | Commit message validation |
| `pre-push.sample` | Pre-push testing |
| `pre-rebase.sample` | Pre-rebase warnings |
| `post-checkout.sample` | Post-checkout setup |
| `post-merge.sample` | Post-merge dependency check |
| `README.md` | Git hooks documentation |

## Quick Reference

### Installing Git Hooks
```bash
# Copy hooks to .git/hooks/
cp .githooks/pre-commit.sample .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

### Running Label Sync (GitHub)
```bash
# Requires github-label-sync
npx github-label-sync --access-token $GITHUB_TOKEN --labels .github/labels.yml rechain-ai/rechain-ide
```

### Manual CI Trigger (GitLab)
```bash
# Pipeline can be triggered via API
curl -X POST --fail -F token=$CI_JOB_TOKEN -F ref=main \
  https://gitlab.com/api/v4/projects/$CI_PROJECT_ID/trigger/pipeline
```

## Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[(optional scope)]: <description>

[optional body]

[optional footer(s)]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`

## Branch Protection Rules

Both GitHub and GitLab are configured with:
- Required PR/MR reviews (minimum 1)
- Required status checks
- No direct pushes to main
- No force pushes
- No deletion of protected branches

## Support

For questions about Git/GitHub/GitLab configuration:
- Git: See `GIT_CONFIGURATION.md`
- GitHub: Check `.github/` directory
- GitLab: Check `.gitlab/` directory
- Git Hooks: See `.githooks/README.md`
