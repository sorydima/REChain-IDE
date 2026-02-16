# Git Workflow Diagrams

Visual guides for common git workflows in the REChain Quantum-CrossAI IDE Engine project.

## Branching Strategy

```
main (production)
│
├─ hotfix/critical-fix ─────┐
│                           │
└───────────────────────────┴─────── merge ─────> main
                            │
develop (integration) <───────┘
│
├─ feature/new-ui ──────────┐
│                           │
├─ feature/api-v2 ──────────┤
│                           │
└─ bugfix/memory-leak ──────┤
                            │
                            └─────── merge ─────> develop
                                                     │
                                                     │
main <─────────────────────────────────────────────────┘ (release)
```

## Feature Development Workflow

```
1. Start from develop
   git checkout develop
   git pull origin develop

2. Create feature branch
   git checkout -b feature/my-feature

3. Make commits
   git commit -m "feat: add new feature"

4. Push branch
   git push origin feature/my-feature

5. Create PR → develop

6. Code review

7. Merge to develop

8. Eventually release → main
```

## Hotfix Workflow

```
main (production has bug)
│
├─ hotfix/security-patch ───┐
│                           │
│                           ├── merge ────> main (immediate fix)
│                           │                   │
develop <───────────────────┴───────────────────┤ (cherry-pick or
                                                │  merge back)
```

## Commit Message Flow

```
Working Directory
        │
        │ git add
        ▼
   Staging Area
        │
        │ git commit -m "type(scope): description"
        ▼
    Local Repo
        │
        │ git push
        ▼
   Remote Repo (GitHub/GitLab)
        │
        │ PR/MR opened
        ▼
    Code Review
        │
        │ Approved
        ▼
    Merge to main/develop
```

## PR Lifecycle

```
┌─────────────┐
│   Create    │
│    PR       │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│    CI       │────── fail ───┐
│   Checks    │               │
└──────┬──────┘               │
       │ pass                 │
       ▼                      │
┌─────────────┐               │
│   Review    │───── reject ─┤
│   Process   │               │
└──────┬──────┘               │
       │ approve              │
       ▼                      │
┌─────────────┐               │
│   Merge     │               │
│   to Base   │               │
└──────┬──────┘               │
       │                      │
       └───────────────────────┘
              │
              ▼
        ┌─────────┐
        │ Deploy  │
        │         │
        └─────────┘
```

## Release Workflow

```
develop
  │
  │ Feature complete, ready for release
  ▼
┌───────────────┐
│ Create        │─── git checkout -b release/v1.2.0
│ release/v1.2.0│
└───────┬───────┘
        │
        │ Version bump, final testing
        ▼
┌───────────────┐
│ Merge to main │─── version tag v1.2.0
│               │
└───────┬───────┘
        │
        │ Deploy to production
        ▼
┌───────────────┐
│ Merge back to │─── ensure develop has all changes
│ develop       │
└───────────────┘
```

## Git Commands Quick Reference

### Daily Workflow

```bash
# Update local branches
git fetch origin
git checkout develop
git pull origin develop

# Create feature
git checkout -b feature/my-feature

# Regular commits
git add .
git commit -m "feat: description"

# Push and PR
git push origin feature/my-feature
# Create PR on GitHub/GitLab

# After merge, cleanup
git checkout develop
git pull origin develop
git branch -d feature/my-feature
```

### Fixing Mistakes

```bash
# Undo unstaged changes
git checkout -- <file>

# Undo staged changes
git reset HEAD <file>

# Amend last commit
git commit --amend

# Undo last commit, keep changes
git reset HEAD~1

# Undo last commit, discard changes
git reset --hard HEAD~1

# Rebase interactive (clean history)
git rebase -i HEAD~5
```

### Resolving Conflicts

```bash
# During merge/rebase with conflicts
git status                    # See conflicted files

# Edit files to resolve
# Mark as resolved
git add <resolved-file>

# Continue
git rebase --continue
# or
git merge --continue
```

## CI/CD Integration

```
Local Development                    Remote Repository
     │                                    │
     │ git push origin feature/xyz        │
     └───────────────────────────────────>│
                                          │
                                          ▼
                                    ┌─────────────┐
                                    │ CI Pipeline │
                                    │  Triggered  │
                                    └──────┬──────┘
                                           │
                           ┌───────────────┼───────────────┐
                           │               │               │
                           ▼               ▼               ▼
                     ┌─────────┐    ┌─────────┐    ┌─────────┐
                     │  Build  │───>│  Test   │───>│ Security│
                     └─────────┘    └─────────┘    └────┬────┘
                                                          │
                                                          ▼
                                          ┌─────────────────┐
                                          │  PR Checks Pass │
                                          └────────┬────────┘
                                                   │
                              ┌────────────────────┼────────────────────┐
                              │                    │                    │
                              ▼                    ▼                    ▼
                        ┌──────────┐      ┌──────────┐      ┌──────────┐
                        │ Review 1 │      │ Review 2 │      │ Approved │
                        └──────────┘      └──────────┘      └─────┬─────┘
                                                                  │
                                                                  ▼
                                                        ┌──────────────────┐
                                                        │ Merge to develop │
                                                        └────────┬─────────┘
                                                                 │
                                                                 ▼
                                                        ┌──────────────────┐
                                                        │ Release to main  │
                                                        └──────────────────┘
```

## Best Practices Summary

1. **Branch naming**: `type/description` (e.g., `feature/add-login`)
2. **Commit format**: `type(scope): description`
3. **PR size**: Keep under 500 lines when possible
4. **Review turnaround**: 24-48 hours target
5. **Clean history**: Rebase before merging, squash if needed
6. **Delete branches**: After merge, clean up
7. **Tests pass**: Before creating PR
8. **Documentation**: Update if needed

See [PR_BEST_PRACTICES.md](./PR_BEST_PRACTICES.md) for more details.
