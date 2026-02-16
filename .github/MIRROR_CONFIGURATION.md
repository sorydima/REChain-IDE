# Mirror Configuration for GitHub <-> GitLab
# This file documents the mirroring setup between platforms

## Overview

The REChain Quantum-CrossAI IDE Engine repository is mirrored between:
- **GitHub**: https://github.com/rechain-ai/rechain-ide
- **GitLab**: https://gitlab.com/rechain-ai/rechain-ide

## Mirroring Strategy

### Primary: GitHub

GitHub is the primary platform for:
- Issue tracking
- Pull request reviews
- Community discussions
- Primary CI/CD (GitHub Actions)

### Secondary: GitLab

GitLab mirrors for:
- Backup
- Alternative CI/CD (GitLab CI)
- Users preferring GitLab interface
- GitLab-specific integrations

## Sync Direction

### GitHub to GitLab (Push Mirroring)

- Commits
- Tags
- Branches

### Not Mirrored

- Issues (platform-specific)
- Pull/Merge Requests (platform-specific)
- CI/CD configurations (platform-specific)
- Project settings (platform-specific)

## Configuration

### GitLab Mirror Setup

In GitLab project settings:
1. Settings → Repository → Mirroring repositories
2. Add mirror:
   - URL: `https://github.com/rechain-ai/rechain-ide.git`
   - Mirror direction: Pull
   - Authentication: Token with repo access
   - Protected branches only: No

### GitHub Actions for GitLab Sync

Alternative: Use GitHub Actions to push to GitLab:

```yaml
name: Sync to GitLab
on:
  push:
    branches:
      - main
      - develop
  workflow_dispatch:

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
          
      - name: Push to GitLab
        run: |
          git remote add gitlab https://oauth2:${{ secrets.GITLAB_TOKEN }}@gitlab.com/rechain-ai/rechain-ide.git
          git push gitlab --all
          git push gitlab --tags
```

## Keeping Platforms in Sync

### Manual Elements to Sync

1. **Issue Templates**
   - Keep `.github/ISSUE_TEMPLATE/` and `.gitlab/issue_templates/` similar
   - Update both when making changes

2. **Documentation**
   - Mirror conceptually, adapt for platform differences
   - Keep guides up to date on both platforms

3. **CI/CD Concepts**
   - Same jobs, different syntax
   - Keep feature parity where possible

### Workflow

When making changes:

1. **GitHub First**
   - Create PR on GitHub
   - Merge to main

2. **GitLab Auto-Sync**
   - Mirror pulls changes automatically
   - GitLab CI runs on mirrored code

3. **GitLab-Specific Updates**
   - Update `.gitlab/` configurations separately
   - These are not mirrored (platform-specific)

## Contributing

### GitHub Contributors

- Use GitHub as normal
- Changes sync to GitLab automatically
- No additional action needed

### GitLab Contributors

1. Fork on GitLab
2. Create MR on GitLab
3. Changes can be manually synced back to GitHub if needed

**Note**: Prefer GitHub for main contributions to avoid complexity.

## CI/CD Parity

| Feature | GitHub Actions | GitLab CI |
|---------|---------------|-----------|
| Build | build.yml | build.gitlab-ci.yml |
| Test | automated-testing.yml | test.gitlab-ci.yml |
| Security | security-scan.yml | security.gitlab-ci.yml |
| Release | release.yml | release.gitlab-ci.yml |

Both platforms run similar checks but independently.

## Troubleshooting

### Sync Failures

If mirroring fails:
1. Check authentication token
2. Verify repository permissions
3. Check for branch protection conflicts
4. Review GitLab mirror logs

### Divergence

If repositories diverge:
1. Identify source of truth (GitHub)
2. Force sync from GitHub to GitLab
3. Do not merge divergent histories

## Maintenance

### Monthly Tasks
- Verify mirror is functioning
- Check for sync delays
- Update tokens if needed

### Quarterly Tasks
- Review both configurations
- Ensure feature parity
- Update documentation

## Contact

For mirror issues:
- Email: dev@rechain.ai
- Discord: https://discord.gg/rechain
