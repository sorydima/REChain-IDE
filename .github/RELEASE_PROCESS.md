# Release Process

This document describes the release process for the REChain Quantum-CrossAI IDE Engine.

## Version Numbering

We follow [Semantic Versioning](https://semver.org/):

```
MAJOR.MINOR.PATCH

MAJOR - Breaking changes
MINOR - New features (backward compatible)
PATCH - Bug fixes (backward compatible)

Example: 1.2.3
```

## Release Schedule

| Type | Frequency | Example |
|------|-----------|---------|
| Patch | As needed | 1.2.1, 1.2.2 |
| Minor | Monthly | 1.3.0, 1.4.0 |
| Major | Quarterly | 2.0.0, 3.0.0 |

## Release Checklist

### Pre-Release

- [ ] All tests passing on `develop`
- [ ] Documentation updated
- [ ] CHANGELOG.md updated with all changes
- [ ] Security audit completed
- [ ] Performance benchmarks acceptable
- [ ] No open critical issues
- [ ] Release notes drafted

### Release Steps

1. **Create Release Branch**
   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b release/v1.2.0
   ```

2. **Bump Version**
   ```bash
   # Update version in PROJECT.json
   ./scripts/bump-version.sh minor
   ```

3. **Final Testing**
   ```bash
   make test
   make lint
   make build
   ```

4. **Create PR to main**
   - Title: `Release v1.2.0`
   - Include changelog summary
   - Tag with `release` label

5. **After Merge to main**
   ```bash
   git checkout main
   git pull origin main
   git tag -a v1.2.0 -m "Release v1.2.0"
   git push origin v1.2.0
   ```

6. **Create Release**
   - GitHub: Go to Releases → Draft new release
   - GitLab: Go to Tags → New release
   - Use the generated release notes
   - Attach binaries if applicable

7. **Merge back to develop**
   ```bash
   git checkout develop
   git merge main
   git push origin develop
   ```

## Automated Release (GitHub)

With semantic-release:

```bash
# Push to main with conventional commits
# Release happens automatically

git commit -m "feat: new feature"
git push origin main

# Creates release, updates changelog, tags version
```

## Hotfix Process

For critical production fixes:

1. **Create from main**
   ```bash
   git checkout main
   git checkout -b hotfix/critical-fix
   ```

2. **Fix and test**
   ```bash
   # Make fix
   git commit -m "fix: resolve critical issue"
   ```

3. **PR to main** (expedited review)

4. **After merge**
   ```bash
   git tag -a v1.2.1 -m "Hotfix v1.2.1"
   git push origin v1.2.1
   ```

5. **Cherry-pick to develop**
   ```bash
   git checkout develop
   git cherry-pick <hotfix-commit>
   ```

## Release Notes Template

```markdown
## Release v1.2.0

### Highlights
- New quantum circuit visualizer
- Improved AI code completion
- 50% faster compilation

### New Features
- feat: Add support for Solidity 0.9 (#456)
- feat: Implement dark mode theme (#234)

### Bug Fixes
- fix: Resolve memory leak in parser (#789)
- fix: Fix race condition in worker pool (#567)

### Improvements
- perf: Optimize bytecode generation (#890)
- refactor: Simplify AST traversal (#678)

### Breaking Changes
None

### Contributors
@alice, @bob, @charlie

### Assets
- rechain-ide-linux-amd64
- rechain-ide-windows-amd64.exe
- rechain-ide-darwin-amd64
```

## Deployment

### Staging

```bash
# Automatically deployed on merge to develop
# URL: https://staging.rechain-ide.example.com
```

### Production

```bash
# Manually triggered after release
# URL: https://rechain-ide.example.com
```

## Post-Release

1. **Monitor**
   - Error rates
   - Performance metrics
   - User feedback

2. **Announce**
   - Blog post
   - Discord announcement
   - Twitter

3. **Update Roadmap**
   - Mark completed features
   - Plan next release

## Rollback Procedure

If critical issues are found:

1. **Immediate**
   ```bash
   # Revert to previous version
   git revert <release-commit>
   git push origin main
   ```

2. **Fix forward**
   - Create hotfix branch
   - Fix issue
   - Deploy hotfix

3. **Communicate**
   - Update status page
   - Notify users
   - Document incident

## Version Support

| Version | Status | Support Until |
|---------|--------|---------------|
| 1.2.x   | Active | Current + 6 months |
| 1.1.x   | Security | Current + 3 months |
| 1.0.x   | EOL | Not supported |

## Tools

- Version bumping: `./scripts/bump-version.sh`
- Changelog: `make changelog`
- Release build: `make release`

See [scripts/](../scripts/) for all release-related scripts.
