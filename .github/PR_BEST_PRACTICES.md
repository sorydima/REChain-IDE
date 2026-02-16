# Pull Request Best Practices

This guide outlines best practices for creating and reviewing Pull Requests (GitHub) / Merge Requests (GitLab) in the REChain Quantum-CrossAI IDE Engine project.

## Before Creating a PR

### 1. Branch Naming

Follow the convention: `<type>/<description>`

| Type | Purpose | Example |
|------|---------|---------|
| `feature/` | New features | `feature/add-login` |
| `bugfix/` | Bug fixes | `bugfix/fix-memory-leak` |
| `hotfix/` | Critical production fixes | `hotfix/security-patch` |
| `docs/` | Documentation updates | `docs/update-readme` |
| `refactor/` | Code refactoring | `refactor/simplify-parser` |
| `release/` | Release preparation | `release/v1.2.0` |

### 2. Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[(optional scope)]: <description>

[optional body]

[optional footer(s)]
```

Examples:
```
feat(auth): add OAuth2 login support

Implement OAuth2 flow for Google and GitHub providers.
Closes #123

fix(compiler): resolve type inference issue

The parser was incorrectly handling nested generic types.
Fixes #456

docs(api): update REST endpoint documentation

Added examples for all new endpoints introduced in v0.2.
```

### 3. PR Checklist

Run the check script:
```bash
./scripts/check-pr.sh
```

Or manually verify:
- [ ] Code compiles/builds successfully
- [ ] All tests pass
- [ ] Code follows style guidelines (gofmt, linting)
- [ ] Documentation updated if needed
- [ ] Commit messages follow convention
- [ ] No merge conflicts
- [ ] Branch is up to date with target

## Creating a Great PR

### PR Title

Follow the same format as commit messages:

```
feat: add user authentication
fix: resolve race condition in worker pool
docs: update API reference
test: add integration tests for payment flow
```

### PR Description

Use the template provided. Include:

1. **Summary**: What changed and why
2. **Changes**: List of specific changes
3. **Testing**: How you tested the changes
4. **Screenshots**: For UI changes
5. **Breaking Changes**: If any

Example:
```markdown
## Summary
Add OAuth2 authentication support for Google and GitHub providers.

## Changes
- Implemented OAuth2 flow in `auth/oauth.go`
- Added configuration for provider credentials
- Created login/logout handlers
- Updated middleware to validate tokens

## Testing
- Unit tests for OAuth2 flow
- Integration tests with mock providers
- Manual testing with Google OAuth

## Breaking Changes
None - new feature only.

## Related Issues
Closes #123
```

### PR Size

Keep PRs focused and reasonably sized:

| Size | Lines of Code | Review Time |
|------|---------------|-------------|
| Small | < 50 | 10-15 min |
| Medium | 50-200 | 15-30 min |
| Large | 200-500 | 30-60 min |
| XLarge | > 500 | Should be split |

**If your PR is too large:**
- Split into smaller, logical PRs
- Use stacked PRs (PRs that depend on other PRs)
- Separate refactoring from feature changes

## During Review

### For Authors

1. **Respond promptly**: Address comments within 24-48 hours
2. **Be open to feedback**: Don't take criticism personally
3. **Explain decisions**: If you disagree, explain your reasoning
4. **Update promptly**: Push changes after addressing comments

2. **Resolve conversations**: Mark as resolved when fixed

### For Reviewers

1. **Review within 24-48 hours**: Don't block team progress
2. **Be constructive**: Suggest improvements, don't just criticize
3. **Explain reasoning**: Help author understand why changes are needed
4. **Distinguish requirements**:
   - **Must fix**: Blocking issues (bugs, security, tests)
   - **Should fix**: Important but not blocking
   - **Could fix**: Nice to have, author's discretion
   - **Consider**: Suggestions for thought

**Example review comments:**

```
**Must fix:** This introduces a race condition. Use mutex or channel.

**Should fix:** Consider extracting this into a separate function.

**Could fix:** Variable name could be more descriptive (suggestions?).

**Consider:** Would a switch statement be clearer here?
```

## Common Review Feedback

### Code Quality

| Issue | Solution |
|-------|----------|
| Missing tests | Add unit/integration tests |
| No documentation | Add comments, update README |
| Complex logic | Simplify or add explanatory comments |
| Magic numbers | Use named constants |
| Duplicate code | Extract to shared function |

### Git Issues

| Issue | Solution |
|-------|----------|
| Too many commits | Squash related commits |
| Merge commits | Rebase on latest target branch |
| Conflicts | Rebase and resolve |
| WIP commits | Clean up before submitting |

## Handling Rejection

If your PR needs significant changes:

1. **Don't take it personally**: Code review improves quality
2. **Ask clarifying questions**: Ensure you understand feedback
3. **Make required changes**: Address blocking issues
4. **Explain trade-offs**: If you disagree, discuss alternatives
5. **Split if needed**: Large rewrites might warrant new PR

## Special Cases

### Draft PRs

Use draft status when:
- PR is not ready for review
- Want early feedback on approach
- Work in progress

Mark as "Ready for review" when complete.

### Emergency Fixes

For critical production issues:
1. Label as `hotfix` and `priority::critical`
2. Include incident context in description
3. Request expedited review
4. Consider pair-review for speed

### Breaking Changes

When introducing breaking changes:
1. Clearly label in PR title: `feat!: new API (BREAKING)`
2. Document migration path
3. Announce in advance if possible
4. Update relevant documentation

## After Merge

1. **Delete branch**: Keep repository clean
2. **Close related issues**: Use `Closes #X` in PR
3. **Update changelog**: If applicable
4. **Deploy**: Follow deployment process
5. **Monitor**: Watch for issues after release

## Tools & Automation

### Automated Checks

Our CI runs:
- Build verification
- Test suite
- Linting/formatting
- Security scanning
- Dependency checking

### PR Automation

- Labels auto-applied based on size and type
- Stale PRs marked after 30 days
- Welcome message for new contributors

## Questions?

- Check [CONTRIBUTING.md](../CONTRIBUTING.md)
- Ask in [Discussions](https://github.com/rechain-ai/rechain-ide/discussions)
- Join our [Discord](https://discord.gg/rechain)
