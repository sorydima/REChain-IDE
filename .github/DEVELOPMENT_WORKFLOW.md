# Development Workflow Guide

This guide describes the day-to-day development workflow for the REChain Quantum-CrossAI IDE Engine project.

## Daily Workflow

### Morning Routine

1. **Update Local Repository**
   ```bash
   git checkout develop
   git pull origin develop
   ```

2. **Check Status**
   ```bash
   # See what you're working on
   git branch
   
   # Check for uncommitted changes
   git status
   ```

3. **Review Notifications**
   - Check GitHub/GitLab for:
     - PR/MR reviews requested
     - Comments on your work
     - CI failures

### Starting New Work

1. **Create Feature Branch**
   ```bash
   ./scripts/create-branch.sh feature my-new-feature
   ```

2. **Make Changes**
   - Edit files in your IDE
   - Follow code style guidelines
   - Write tests alongside code

3. **Regular Commits**
   ```bash
   # Stage changes
   git add .
   
   # Commit with descriptive message
   git commit -m "feat(scope): description"
   
   # Push to remote
   git push origin feature/my-new-feature
   ```

### During Development

**Every 30-60 minutes:**
```bash
# Run tests
cd rechain-ide && go test ./...

# Check formatting
gofmt -d .

# Run linter
golangci-lint run
```

**Before finishing:**
```bash
# Full test suite
make test

# Build verification
make build

# PR check
./scripts/check-pr.sh
```

## Code-Review Workflow

### Submitting for Review

1. **Push Latest Changes**
   ```bash
   git push origin feature/my-new-feature
   ```

2. **Create PR/MR**
   - Fill in template
   - Link related issues
   - Add screenshots for UI changes
   - Request specific reviewers if needed

3. **Monitor CI**
   - Wait for checks to pass
   - Fix any failures
   - Respond to automated feedback

### Addressing Review Feedback

```bash
# Make requested changes
# Edit files...

# Stage changes
git add .

# Commit (amend or new commit)
# Option 1: New commit
git commit -m "refactor: address review feedback"

# Option 2: Amend last commit
git commit --amend

# Push (force push if amended)
git push origin feature/my-new-feature --force-with-lease
```

### After Merge

1. **Cleanup**
   ```bash
   git checkout develop
   git pull origin develop
   git branch -d feature/my-new-feature
   git push origin --delete feature/my-new-feature
   ```

2. **Update Local Environment**
   ```bash
   # Update dependencies if needed
   go work sync
   
   # Clean build artifacts
   make clean
   ```

## Collaboration Patterns

### Pair Programming

```bash
# Share branch
./scripts/create-branch.sh feature/shared-feature

# Both work on same branch
git push origin feature/shared-feature

# Coordinate commits
# Use descriptive commit messages indicating pair
```

### Code Review Rotation

| Day | Primary Reviewer |
|-----|------------------|
| Mon | Alice |
| Tue | Bob |
| Wed | Charlie |
| Thu | Alice |
| Fri | Bob |

### Handling Conflicts

```bash
# While on your feature branch
git fetch origin
git rebase origin/develop

# If conflicts:
# 1. Edit files to resolve
# 2. Stage resolved files
git add <resolved-files>

# 3. Continue rebase
git rebase --continue

# 4. Force push
git push origin feature/my-new-feature --force-with-lease
```

## Testing Workflow

### Unit Tests

```bash
cd rechain-ide

# Run all tests
go test ./...

# Run specific package
go test ./internal/compiler

# Run with race detector
go test -race ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Integration Tests

```bash
# Start dependencies
docker-compose up -d

# Run integration tests
go test -tags=integration ./...

# Cleanup
docker-compose down
```

### Manual Testing

```bash
# Build and run locally
make build
./bin/rechain-ide

# Test specific feature
./bin/rechain-ide --feature-flag
```

## Debugging Workflow

### Finding Issues

```bash
# Check recent changes
git log --oneline -10

# Find when bug was introduced
git bisect start
git bisect bad HEAD
git bisect good v1.0.0
# Follow prompts...

# Search for relevant code
grep -r "functionName" rechain-ide/
```

### Debug Logging

```go
// Add temporary debug logging
log.Printf("DEBUG: variable = %v", variable)
```

### Using Debugger

```bash
# Delve debugger
dlv debug ./cmd/rechain-ide
(dlv) break main.main
(dlv) continue
```

## Release Preparation

### Pre-Release Checklist

```bash
# 1. Update changelog
make changelog

# 2. Run full test suite
make test-all

# 3. Check security
make security-check

# 4. Update version
./scripts/bump-version.sh minor

# 5. Create release branch
git checkout -b release/v1.2.0

# 6. Final testing
make test
make build
```

### Hotfix Process

```bash
# 1. Create from main
git checkout main
git checkout -b hotfix/critical-fix

# 2. Fix the issue
# Edit files...
git commit -m "fix: resolve critical issue"

# 3. Fast-track review
# Get approval ASAP

# 4. Merge and tag
git checkout main
git merge hotfix/critical-fix
git tag -a v1.2.1 -m "Hotfix v1.2.1"

# 5. Cherry-pick to develop
git checkout develop
git cherry-pick <hotfix-commit>
```

## Troubleshooting Common Issues

### "My commit is too big"

```bash
# Interactive rebase to split
git rebase -i HEAD~5
# Mark commits for editing/squashing
```

### "I committed to wrong branch"

```bash
# Undo last commit, keep changes
git reset HEAD~1

# Checkout correct branch
git checkout correct-branch

# Commit there
git add .
git commit -m "message"
```

### "I need to undo a pushed commit"

```bash
# Revert (creates new commit)
git revert <commit-hash>
git push

# Or reset (use carefully)
git reset --hard HEAD~1
git push origin branch-name --force-with-lease
```

## Time Management

### Pomodoro Technique for Coding

```
25 min: Code
5 min:  Break (check messages)
25 min: Code
5 min:  Break (run tests)
25 min: Code
5 min:  Break (review others)
25 min: Code
15 min: Long break
```

### Daily Schedule Example

| Time | Activity |
|------|----------|
| 9:00 | Check notifications, standup |
| 9:30 | Code (feature work) |
| 11:00 | Break, check CI |
| 11:15 | Continue coding |
| 12:30 | Lunch |
| 13:30 | Code review |
| 15:00 | Testing, documentation |
| 16:00 | Wrap up, commit |
| 17:00 | End |

## Tools Integration

### IDE Shortcuts

| Action | VS Code | GoLand |
|--------|---------|--------|
| Run tests | Ctrl+Shift+T | Ctrl+Shift+F10 |
| Go to definition | F12 | Ctrl+B |
| Find usages | Shift+F12 | Alt+F7 |
| Format code | Shift+Alt+F | Ctrl+Alt+L |
| Quick fix | Ctrl+. | Alt+Enter |

### Git Aliases (Recommended)

```bash
# Add to ~/.gitconfig
[alias]
    st = status
    co = checkout
    br = branch
    ci = commit
    lg = log --oneline --graph
    undo = reset HEAD~1 --mixed
    amend = commit --amend --no-edit
```

## Weekly Review

Every Friday:

1. **Review completed work**
   - Closed issues/PRs
   - Merged features

2. **Plan next week**
   - Prioritize backlog
   - Set goals

3. **Clean up**
   - Delete old branches
   - Update dependencies
   - Clear temporary files

4. **Documentation**
   - Update docs with recent changes
   - Write blog post if major feature

See also:
- [Environment Setup](./ENVIRONMENT_SETUP.md)
- [PR Best Practices](./PR_BEST_PRACTICES.md)
- [First Time Contributors](./FIRST_TIME_CONTRIBUTORS.md)
