# CI/CD Troubleshooting Guide

This document provides solutions for common CI/CD issues in the REChain Quantum-CrossAI IDE Engine project.

## GitHub Actions

### Workflow Not Triggering

**Symptom:** Push doesn't trigger workflow

**Solutions:**
1. Check if file is in `.github/workflows/` directory
2. Verify YAML syntax is valid
3. Check branch filters in workflow `on:` section
4. Ensure workflows are enabled in repository settings

**Debug:**
```bash
# Check workflow syntax
gh workflow view <workflow-name>

# List recent workflow runs
gh run list --workflow=<workflow-name>
```

### Job Stuck or Hanging

**Symptom:** Job doesn't complete or times out

**Solutions:**
1. Check for infinite loops in scripts
2. Verify network requests aren't hanging
3. Add timeout configuration:
   ```yaml
   jobs:
     build:
       timeout-minutes: 30
   ```

**Cancel stuck run:**
```bash
gh run cancel <run-id>
```

### Permission Denied Errors

**Symptom:** `Error: Resource not accessible by integration`

**Solutions:**
1. Check repository settings → Actions → Workflow permissions
2. Ensure `GITHUB_TOKEN` has required permissions
3. For PRs from forks, use `pull_request_target` carefully

**Example fix:**
```yaml
permissions:
  contents: write
  pull-requests: write
  issues: write
```

### Secret Not Available

**Symptom:** Secret shows as empty or undefined

**Solutions:**
1. Verify secret is set in repository settings
2. Secrets aren't available in PRs from forks
3. Check secret name matches exactly (case-sensitive)

**Workaround for forks:**
```yaml
# Use environment variables for non-sensitive config
env:
  MY_VAR: ${{ vars.MY_VARIABLE }}  # For non-sensitive
  MY_SECRET: ${{ secrets.MY_SECRET }}  # For sensitive
```

## GitLab CI

### Pipeline Not Running

**Symptom:** No pipeline appears after push

**Solutions:**
1. Verify `.gitlab-ci.yml` exists at root
2. Check if CI/CD is enabled in project settings
3. Validate YAML syntax with CI Lint

**Debug:**
```bash
# Validate CI config
curl --location --request POST \
  'https://gitlab.com/api/v4/projects/<project-id>/ci/lint' \
  --form 'content=@.gitlab-ci.yml'
```

### Job Failed: "No runner available"

**Symptom:** Job stays pending indefinitely

**Solutions:**
1. Check if shared runners are enabled
2. Verify runner tags match job requirements
3. Register a specific runner if needed

**Check runner status:**
- Project Settings → CI/CD → Runners

### Cache Not Working

**Symptom:** Dependencies re-download every run

**Solutions:**
1. Verify cache key is unique per branch/job
2. Check cache paths are correct
3. Ensure cache is being uploaded (check job artifacts)

**Example fix:**
```yaml
cache:
  key: ${CI_COMMIT_REF_SLUG}-${CI_JOB_NAME}
  paths:
    - .go-cache/
  policy: pull-push
```

### Docker Login Failed

**Symptom:** `Error response from daemon: unauthorized`

**Solutions:**
1. Verify CI_REGISTRY_USER and CI_REGISTRY_PASSWORD are set
2. Check registry URL is correct
3. Ensure service account has push permissions

**Fix for GitLab Registry:**
```yaml
before_script:
  - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
```

## Common Issues

### Test Failures Only in CI

**Symptoms:** Tests pass locally but fail in CI

**Causes & Solutions:**

1. **Environment differences**
   - Use containerized environments
   - Pin tool versions (Go version, Node version)

2. **Timing issues**
   - Add retries for flaky tests
   - Increase timeouts for slower CI environments

3. **Case sensitivity**
   - Linux CI is case-sensitive, macOS/Windows may not be
   - Fix file path casing

4. **Line endings**
   - Configure `.gitattributes` properly
   - Use `dos2unix` or `unix2dos` if needed

### Build Artifacts Missing

**Symptom:** Expected files not available in subsequent jobs

**Solution:**
```yaml
# In producing job
artifacts:
  paths:
    - bin/
  expire_in: 1 week

# In consuming job
job:
  needs:
    - job: build
      artifacts: true
```

### Large File/Repository Issues

**Symptom:** Clone or checkout takes too long

**Solutions:**
1. Shallow clones:
   ```yaml
   variables:
     GIT_DEPTH: 10
   ```

2. Partial clones (Git 2.27+):
   ```yaml
   variables:
     GIT_STRATEGY: clone
     GIT_DEPTH: 0
   ```

3. Cache dependencies instead of vendoring

## Debugging Workflows

### Enable Debug Logging

**GitHub Actions:**
```yaml
env:
  ACTIONS_STEP_DEBUG: true
  ACTIONS_RUNNER_DEBUG: true
```

**GitLab CI:**
```yaml
variables:
  CI_DEBUG_SERVICES: "true"
```

### Local Testing

**GitHub Actions with act:**
```bash
# Install act
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# Run workflow locally
act -P ubuntu-latest=nektos/act-environments-ubuntu:18.04

# Run specific job
act -j <job-name>
```

**GitLab CI with local runner:**
```bash
# Run specific job locally
gitlab-runner exec docker <job-name>
```

## Performance Optimization

### Speed Up Builds

1. **Use caches effectively**
   - Cache dependencies
   - Cache build outputs when possible

2. **Parallel jobs**
   - Split test suites across multiple jobs
   - Use matrices for testing multiple versions

3. **Optimize Docker images**
   - Use smaller base images
   - Multi-stage builds
   - Layer caching

### Reduce Billable Minutes

1. **Cancel redundant runs**
   ```yaml
   concurrency:
     group: ${{ github.workflow }}-${{ github.ref }}
     cancel-in-progress: true
   ```

2. **Skip unnecessary runs**
   ```yaml
   paths-ignore:
     - 'docs/**'
     - '*.md'
   ```

3. **Use self-hosted runners** for large workloads

## Security Best Practices

### Protect Secrets

1. Never log secrets
2. Use masked variables
3. Rotate secrets regularly
4. Use environment-specific secrets

### Secure CI/CD Configuration

1. Review all workflow changes
2. Use CODEOWNERS for workflow files
3. Enable branch protection
4. Require signed commits

## Getting Help

1. **Check logs:** Always start with the full job log
2. **Minimal reproduction:** Create minimal workflow that reproduces issue
3. **Community:** 
   - GitHub Actions Community: https://github.com/orgs/community/discussions
   - GitLab Forum: https://forum.gitlab.com/
4. **Documentation:**
   - GitHub Actions: https://docs.github.com/en/actions
   - GitLab CI/CD: https://docs.gitlab.com/ee/ci/

## Quick Reference

| Issue | GitHub Fix | GitLab Fix |
|-------|-----------|-----------|
| No runner | Check Actions settings | Enable shared runners |
| Timeout | `timeout-minutes` | `timeout` |
| Secrets | Check repo settings | Check CI/CD variables |
| Cache | `actions/cache` | `cache:` keyword |
| Artifacts | `actions/upload-artifact` | `artifacts:` keyword |
