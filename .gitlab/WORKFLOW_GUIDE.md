# GitLab-Specific Workflow Guide

This guide covers GitLab-specific workflows and features for the REChain Quantum-CrossAI IDE Engine project.

## Table of Contents

1. [GitLab Flow](#gitlab-flow)
2. [Merge Requests](#merge-requests)
3. [CI/CD Pipeline](#cicd-pipeline)
4. [Issue Management](#issue-management)
5. [Quick Actions](#quick-actions)
6. [Integration Features](#integration-features)

## GitLab Flow

GitLab Flow is a simpler alternative to Git Flow:

```
main (production)
  │
  ├── feature/add-login ─────────┐
  │                              │
  ├── feature/api-v2 ────────────┤
  │                              │
  └── feature/docs-update ───────┤
                                 │
                                 └─── merge ────> main
                                                    │
  production ────────────────────────────────────────┘ (deploy)
```

### Branch Naming

- `feature/*` - New features
- `bugfix/*` - Bug fixes  
- `hotfix/*` - Production fixes
- `docs/*` - Documentation

## Merge Requests

### Creating an MR

1. Push your branch
2. Visit the project page
3. Click **Merge Requests** → **New Merge Request**
4. Select source and target branches
5. Fill in the template

### MR Template Sections

```markdown
## Summary
Brief description of changes

## Related Issues
Closes #123

## Changes Made
- Change 1
- Change 2

## Testing
How was this tested?

## Checklist
- [ ] Tests pass
- [ ] Documentation updated
```

### WIP MRs

Mark work in progress:
```
Title: WIP: Add user authentication
```

Remove WIP when ready:
```
Title: Add user authentication
```

Or use quick action: `/wip`

## CI/CD Pipeline

### Pipeline Stages

```
build → test → security → docs → deploy → release
```

### Viewing Pipelines

1. Project → **CI/CD** → **Pipelines**
2. See status of all pipelines
3. Click pipeline for detailed view

### Pipeline Triggers

| Event | Pipeline |
|-------|----------|
| Push to feature branch | Yes |
| Merge Request | Yes (with merged results) |
| Push to main | Yes |
| Tag push | Yes (release) |
| Schedule | Yes (cron) |
| Manual | Yes (button) |

### Manual Jobs

Some jobs require manual trigger:
1. Go to pipeline view
2. Click on manual job
3. Click **Play** button

### Retry Failed Jobs

1. Go to pipeline
2. Find failed job
3. Click retry icon

## Issue Management

### Issue Board

Access: Project → **Issues** → **Board**

Default columns:
- **Open**
- **To Do**
- **Doing**
- **Closed**

### Labels

Use scoped labels:
- `type::bug`, `type::feature`
- `priority::high`, `priority::low`
- `status::in-progress`, `status::review`

### Milestones

Group issues into releases:
1. Project → **Issues** → **Milestones**
2. Create milestone with due date
3. Assign issues to milestone

### Due Dates

Set due dates on issues:
```
/due 2024-01-15
```

## Quick Actions

Use these commands in issues/MRs:

| Command | Action |
|---------|--------|
| `/assign @user` | Assign to user |
| `/unassign` | Unassign |
| `/label ~label` | Add label |
| `/unlabel ~label` | Remove label |
| `/milestone %"name"` | Set milestone |
| `/due 2024-01-15` | Set due date |
| `/estimate 4h` | Set time estimate |
| `/spend 2h` | Log time spent |
| `/title New Title` | Change title |
| `/close` | Close issue/MR |
| `/reopen` | Reopen issue/MR |
| `/merge` | Merge MR |
| `/draft` | Mark as draft |
| `/wip` | Toggle WIP |
| `/approve` | Approve MR |
| `/shrug` | Add ¯\\_(ツ)_/¯ |

## Integration Features

### Container Registry

GitLab includes container registry:
```bash
# Login
docker login registry.gitlab.com

# Push
docker tag myimage registry.gitlab.com/group/project/myimage:tag
docker push registry.gitlab.com/group/project/myimage:tag
```

### Package Registry

Store packages:
- Maven
- npm
- PyPI
- etc.

### Wiki

Create documentation:
1. Project → **Wiki**
2. Create pages in Markdown
3. Link from README

### Snippets

Share code:
1. Project → **Snippets**
2. Create public or private snippet
3. Share URL

### Pages

Host static sites:
1. Job deploys to GitLab Pages
2. Access at `https://group.gitlab.io/project`

## Security Features

### Security Dashboard

Project → **Security** → **Security Dashboard**

Shows:
- Vulnerabilities
- Dependency issues
- Secrets detected

### Compliance

Project → **Security** → **Compliance**

Track:
- Merge request approvals
- Pipeline configurations

## GitLab vs GitHub Differences

| Feature | GitLab Approach |
|---------|----------------|
| CI/CD | Integrated, .gitlab-ci.yml |
| Registry | Built-in container registry |
| Wiki | Integrated |
| Pages | Built-in hosting |
| Review | Side-by-side diff |
| Time tracking | Built-in |
| Epics | Group multiple issues |

## Tips & Tricks

### Search

Advanced search syntax:
```
label:bug author:@username milestone:"v1.0"
```

### Notifications

Configure at: User → **Preferences** → **Notifications**

### Keyboard Shortcuts

Press `?` on any page to see shortcuts.

### Todo List

GitLab has built-in todos:
1. Click todo icon in sidebar
2. See all items requiring action

## Troubleshooting

### Pipeline Not Running

1. Check `.gitlab-ci.yml` exists
2. Verify CI/CD is enabled
3. Validate YAML syntax

### Can't Push

1. Check permissions
2. Verify branch protection rules
3. Check if fork is up to date

### MR Conflicts

1. Rebase your branch:
   ```bash
   git fetch upstream
   git rebase upstream/main
   git push origin feature/branch --force-with-lease
   ```

## Resources

- [GitLab Documentation](https://docs.gitlab.com/)
- [GitLab CI/CD](https://docs.gitlab.com/ee/ci/)
- [GitLab Flow](https://docs.gitlab.com/ee/topics/gitlab_flow.html)
- [Quick Actions](https://docs.gitlab.com/ee/user/project/quick_actions.html)
