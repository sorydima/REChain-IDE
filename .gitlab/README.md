# GitLab Configuration Directory

This directory contains all GitLab-specific configuration files for the REChain Quantum-CrossAI IDE Engine project.

## Directory Structure

```
.gitlab/
├── ci/                       # GitLab CI/CD templates
├── issue_templates/          # Issue templates
├── merge_request_templates/  # Merge request templates
├── CODEOWNERS                # Code review ownership
├── CONTRIBUTING.md           # GitLab contribution guide
├── WORKFLOW_GUIDE.md         # GitLab workflow guide
├── SECURITY.md               # Security policy
├── release-template.md       # Release notes template
├── project-settings.yml      # Project configuration
├── issue_board.yml           # Issue board configuration
└── config.yml                # Template configuration
```

## CI/CD Templates (`.gitlab/ci/`)

Modular CI/CD configuration files:

| Template | Purpose |
|----------|---------|
| `common.gitlab-ci.yml` | Base configuration |
| `build.gitlab-ci.yml` | Build jobs |
| `test.gitlab-ci.yml` | Test jobs |
| `security.gitlab-ci.yml` | Security scanning |
| `deploy.gitlab-ci.yml` | Deployment |
| `docs.gitlab-ci.yml` | Documentation |
| `release.gitlab-ci.yml` | Release management |
| `pages.gitlab-ci.yml` | GitLab Pages |
| `container-scanning.gitlab-ci.yml` | Container security |
| `auto-devops.gitlab-ci.yml` | Auto DevOps template |
| ...and more | See directory |

## Issue Templates (`.gitlab/issue_templates/`)

Templates for creating issues:
- `bug_report.md` - Bug reports
- `feature_request.md` - Feature requests
- `task.md` - General tasks

## Merge Request Templates (`.gitlab/merge_request_templates/`)

- `Default.md` - Standard MR template

## Configuration Files

| File | Purpose |
|------|---------|
| `CODEOWNERS` | Automatic reviewer assignment |
| `SECURITY.md` | Security policy |
| `project-settings.yml` | Project settings reference |
| `issue_board.yml` | Issue board configuration |
| `config.yml` | Template configuration |

## Documentation

GitLab-specific guides:

| File | Purpose |
|------|---------|
| `CONTRIBUTING.md` | GitLab contribution guide |
| `WORKFLOW_GUIDE.md` | GitLab workflow and features |
| `release-template.md` | Release notes template |

## Main CI/CD File

The root `.gitlab-ci.yml` includes templates from `.gitlab/ci/`:

```yaml
include:
  - local: '.gitlab/ci/common.gitlab-ci.yml'
  - local: '.gitlab/ci/build.gitlab-ci.yml'
  - local: '.gitlab/ci/test.gitlab-ci.yml'
  # ... etc
```

## Using This Configuration

### For Maintainers

- Update `CODEOWNERS` when team changes
- Modify CI templates as needed
- Keep documentation current
- Configure project settings via UI or API

### For Contributors

1. Fork the repository
2. Create a branch from `develop`
3. Make changes following conventions
4. Push and create MR
5. Ensure pipeline passes
6. Request review

### For Reviewers

- Use MR template checklist
- Verify CI pipeline status
- Follow code review guidelines

## Quick Actions

Use these commands in GitLab issues/MRs:

```
/assign @username
/label ~bug ~priority::high
/milestone %"Upcoming Release"
/estimate 4h
/spend 2h
/approve
/merge
```

## Pipeline Stages

```
build → test → security → docs → deploy → release
```

Each stage runs only if the previous succeeds.

## Integration Features

### Container Registry

Built-in registry at `registry.gitlab.com/group/project`

### Pages

Static site hosting at `group.gitlab.io/project`

### Wiki

Documentation wiki integrated with project

### Snippets

Code snippets for sharing

## Differences from GitHub

| Feature | GitLab |
|---------|--------|
| CI/CD | Integrated (.gitlab-ci.yml) |
| Registry | Built-in |
| Wiki | Integrated |
| Pages | Built-in hosting |
| Time tracking | Built-in |
| Epics | Available (higher tiers) |

## Maintenance

### Updating CI Templates

1. Edit template in `.gitlab/ci/`
2. Test on feature branch
3. Verify pipeline works
4. Merge to develop

### Updating Templates

1. Edit template file
2. Changes apply immediately
3. No deployment needed

### Syncing with GitHub

Some configurations are mirrored in `.github/`:
- Issue templates
- Documentation concepts
- Workflow principles

Keep both platforms consistent.

## Resources

- [GitLab Documentation](https://docs.gitlab.com/)
- [GitLab CI/CD](https://docs.gitlab.com/ee/ci/)
- [GitLab Flow](https://docs.gitlab.com/ee/topics/gitlab_flow.html)

## Questions?

- Check [CONTRIBUTING.md](./CONTRIBUTING.md)
- See [WORKFLOW_GUIDE.md](./WORKFLOW_GUIDE.md)
- Join our [Discord](https://discord.gg/rechain)
