# Reviewing Checklist

This checklist helps reviewers ensure thorough and consistent code reviews.

## Before Starting Review

- [ ] I understand the context and purpose of this PR
- [ ] I've read the related issue(s)
- [ ] I have sufficient time for a thorough review
- [ ] CI checks are passing (or I understand why they're not)

## Code Quality

### Functionality

- [ ] Code works as intended
- [ ] Edge cases are handled
- [ ] Error handling is appropriate
- [ ] No obvious bugs or logic errors
- [ ] Performance is acceptable

### Testing

- [ ] Tests cover new functionality
- [ ] Tests are passing
- [ ] Edge cases are tested
- [ ] Error conditions are tested
- [ ] Test names are descriptive

### Code Style

- [ ] Code follows project conventions
- [ ] Naming is clear and consistent
- [ ] Functions are appropriately sized
- [ ] No unnecessary complexity
- [ ] Comments are helpful (not excessive)

### Security

- [ ] No hardcoded secrets
- [ ] Input is validated
- [ ] No injection vulnerabilities
- [ ] Authentication/authorization checked
- [ ] Sensitive data handled properly

## Documentation

- [ ] Code comments explain "why" not "what"
- [ ] Public APIs are documented
- [ ] README updated if needed
- [ ] Breaking changes documented
- [ ] Examples provided for complex features

## Git Hygiene

- [ ] Commits are logically organized
- [ ] Commit messages follow convention
- [ ] No unnecessary files committed
- [ ] Branch is up to date with target
- [ ] History is clean (rebased if needed)

## Review Feedback

### Types of Comments

**Must Fix (Blocking):**
- Security issues
- Logic errors
- Test failures
- Breaking changes without documentation

**Should Fix (Non-blocking but important):**
- Performance concerns
- Missing tests
- Unclear variable names
- Code duplication

**Could Fix (Suggestions):**
- Alternative approaches
- Optimization opportunities
- Style preferences

**Consider (For thought):**
- Design considerations
- Future improvements
- Architecture discussions

## During Review

1. **Read description first**
   - Understand what changed and why

2. **Check overall approach**
   - Does the solution make sense?
   - Is there a simpler way?

3. **Review commit by commit**
   - Easier to follow logical progression

4. **Test the code**
   - Pull and run locally if unsure
   - Test edge cases

5. **Be constructive**
   - Explain reasoning
   - Suggest improvements, don't just criticize
   - Acknowledge good work

## Review Comments Template

### Good Comment Examples

```
**Must fix:** This doesn't handle the case where `input` is nil, 
which will cause a panic. Please add a nil check.
```

```
**Should fix:** Consider extracting this into a separate function 
for better readability and testability.

Suggested:
```go
func validateInput(input string) error {
    if input == "" {
        return errors.New("input cannot be empty")
    }
    return nil
}
```
```

```
**Could fix:** You could simplify this by using `strings.TrimPrefix()`
instead of manual string manipulation.
```

```
**Consider:** Have you considered the impact on memory usage for 
large files? We might want to process this in chunks.
```

### Comment Prefixes

| Prefix | Meaning | Action Required |
|--------|---------|-----------------|
| **Must fix** | Blocking issue | Must be resolved |
| **Should fix** | Important but not blocking | Strongly recommended |
| **Could fix** | Suggestion | Optional |
| **Consider** | Thought starter | No action needed |
| **Nit** | Minor preference | Author's choice |
| **Question** | Need clarification | Respond only |
| **Praise** | Good work | Acknowledge |

## After Review

### If Approving

- [ ] All "Must fix" items resolved
- [ ] "Should fix" items addressed or agreed to defer
- [ ] PR description is clear
- [ ] Ready to merge

### If Requesting Changes

- [ ] All blocking issues clearly marked
- [ ] Specific guidance provided
- [ ] Expected changes documented
- [ ] Timeline communicated

### If Commenting (not blocking)

- [ ] Non-blocking nature is clear
- [ ] Author can merge when ready
- [ ] Optional suggestions marked as such

## Special Cases

### Security Review

- [ ] Authentication changes reviewed by security team
- [ ] New dependencies checked for vulnerabilities
- [ ] Secrets/credentials properly handled
- [ ] Input validation reviewed
- [ ] Authorization logic verified

### Performance-Critical Code

- [ ] Benchmarks included
- [ ] Memory usage analyzed
- [ ] No unnecessary allocations
- [ ] Algorithmic complexity acceptable
- [ ] Profiling results reviewed

### Breaking Changes

- [ ] Migration guide provided
- [ ] Version bumped appropriately
- [ ] Changelog updated
- [ ] Deprecation warnings added (if applicable)
- [ ] Stakeholders notified

## Review Etiquette

### Do

✅ Be respectful and professional
✅ Explain your reasoning
✅ Suggest improvements, don't just point out problems
✅ Acknowledge good work with positive comments
✅ Respond promptly to follow-up questions
✅ Be open to discussion

### Don't

❌ Use dismissive or rude language
❌ Block without clear explanation
❌ Nitpick without good reason
❌ Ignore the PR for days without comment
❌ Make personal comments about the author
❌ Approve without actually reviewing

## Quick Reference

| Time | Action |
|------|--------|
| < 50 lines | 10-15 min review |
| 50-200 lines | 15-30 min review |
| 200-500 lines | 30-60 min review |
| > 500 lines | Suggest splitting |

| Priority | Response Time |
|----------|---------------|
| Critical/Hotfix | < 4 hours |
| High | < 24 hours |
| Normal | < 48 hours |
| Low | < 1 week |

See also:
- [PR Best Practices](./PR_BEST_PRACTICES.md)
- [Contributing Guide](../CONTRIBUTING.md)
