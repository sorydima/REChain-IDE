# Contributing to REChain Quantum-CrossAI IDE Engine on GitLab

This guide covers contributing via GitLab. For general contribution guidelines, see the main [CONTRIBUTING.md](../CONTRIBUTING.md).

## GitLab Workflow

### 1. Fork the Repository

1. Go to the project page on GitLab
2. Click the **Fork** button
3. Choose your namespace
4. Wait for the fork to complete

### 2. Clone Your Fork

```bash
git clone https://gitlab.com/YOUR_USERNAME/rechain-ide.git
cd rechain-ide
```

### 3. Add Upstream Remote

```bash
git remote add upstream https://gitlab.com/rechain-ai/rechain-ide.git
git fetch upstream
```

### 4. Create a Branch

```bash
# Sync with upstream
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/my-feature
```

### 5. Make Changes and Commit

```bash
# Make changes
# Edit files...

# Commit following conventional commits
git add .
git commit -m "feat(scope): description"
```

### 6. Push and Create MR

```bash
# Push to your fork
git push origin feature/my-feature

# Create MR via GitLab UI
```

## Merge Request Process

### Creating an MR

1. Go to your fork on GitLab
2. Click **Merge Requests** → **New Merge Request**
3. Select source branch (your feature branch)
4. Select target branch (upstream develop)
5. Fill in the MR template

### MR Template

Use the provided template which includes:
- Summary of changes
- Related issues (use `Closes #123`)
- Type of change
- Testing performed
- Checklist

### Review Process

1. **Pipeline must pass**
   - GitLab CI runs automatically
   - Fix any failures before requesting review

2. **Code review**
   - Minimum 1 approval required
   - Address feedback promptly
   - Resolve discussions after fixing

3. **Merge**
   - Use "Merge commit" strategy
   - Delete source branch after merge

## GitLab CI/CD

Our GitLab CI includes:

| Stage | Jobs |
|-------|------|
| build | Build Go binaries, Docker images |
| test | Unit tests, linting, formatting |
| security | SAST, dependency scanning, secrets detection |
| docs | Documentation build and validation |
| deploy | Staging deployment |
| release | Release creation |

### Checking Pipeline Status

```bash
# View pipeline status
gitlab-runner status

# Retry failed job
# (via GitLab UI → Pipelines → Retry)
```

## GitLab-Specific Features

### Quick Actions

Use these in MRs and issues:

```
/assign @username
/label ~bug ~priority::high
/milestone %"Upcoming Release"
/estimate 4h
/spend 2h
/close
/reopen
```

### Issue Board

View and manage issues on the [Issue Board](../-/boards):
- Columns represent labels/statuses
- Drag issues between columns
- Filter by label, assignee, milestone

### Wiki

Documentation can also be added to the [Wiki](../-/wikis).

### Snippets

Share code snippets via [Snippets](../-/snippets).

## Differences from GitHub

| Feature | GitHub | GitLab |
|---------|--------|--------|
| PR/MR | Pull Request | Merge Request |
| Actions | GitHub Actions | GitLab CI/CD |
| Quick Actions | N/A | `/command` syntax |
| Issue Board | Projects | Built-in boards |
| Wiki | Separate | Integrated |

## Best Practices

### Do

✅ Use the MR template  
✅ Link related issues with `Closes #123`  
✅ Keep MRs focused and reasonably sized  
✅ Respond to feedback within 24-48 hours  
✅ Use WIP prefix for incomplete MRs: `WIP: Add feature`  

### Don't

❌ Create MRs from main branch  
❌ Force push after review has started  
❌ Ignore CI failures  
❌ Mark as ready before all checks pass  

## Getting Help

- **GitLab issues**: Create issue with `question` label
- **Email**: dev@rechain.ai
- **Discord**: https://discord.gg/rechain

## See Also

- [CONTRIBUTING.md](../CONTRIBUTING.md) - General contribution guidelines
- [PR_BEST_PRACTICES.md](./PR_BEST_PRACTICES.md) - PR best practices
- [DEVELOPMENT_WORKFLOW.md](./DEVELOPMENT_WORKFLOW.md) - Development workflow
