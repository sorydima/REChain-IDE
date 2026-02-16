# Issue Reporting Guide

This guide helps you create effective bug reports and feature requests for the REChain Quantum-CrossAI IDE Engine project.

## Before Creating an Issue

### Check Existing Issues

1. Search [existing issues](https://github.com/rechain-ai/rechain-ide/issues) for duplicates
2. Check [closed issues](https://github.com/rechain-ai/rechain-ide/issues?q=is%3Aissue+is%3Aclosed) - your issue might be resolved
3. Review [discussions](https://github.com/rechain-ai/rechain-ide/discussions) for Q&A

### For Bug Reports

Verify the issue:
- [ ] Issue occurs in latest version
- [ ] Issue is reproducible
- [ ] Not caused by external factors (OS, dependencies)
- [ ] Not a duplicate of existing issue

### For Feature Requests

Consider:
- [ ] Feature aligns with project goals
- [ ] Similar feature doesn't already exist
- [ ] Use case is clearly defined
- [ ] Alternatives have been considered

## Creating a Bug Report

### Title Format

```
[<Component>] <Brief description>

Examples:
[Compiler] Type inference fails for nested generics
[CLI] Login command hangs on Windows
[Docs] Missing API documentation for /auth endpoints
```

### Required Information

**1. Environment**
```
- OS: [e.g., Windows 11, macOS 14, Ubuntu 22.04]
- Go Version: [e.g., 1.21.0]
- IDE Version: [e.g., 0.1.0]
- Installation method: [e.g., binary, source, docker]
```

**2. Steps to Reproduce**
```
1. Start the application with 'rechain-ide start'
2. Open file 'example.sol'
3. Click on 'Compile' button
4. See error
```

**3. Expected Behavior**
```
The file should compile without errors and produce bytecode.
```

**4. Actual Behavior**
```
Compilation fails with error: "type mismatch in line 42"
```

**5. Logs & Screenshots**
```
Include:
- Full error message
- Relevant log output
- Screenshots if UI issue
- Stack trace if available
```

### Minimal Reproduction

Provide the smallest possible example:

**Good:**
```go
// This 10-line program reproduces the issue
package main

func main() {
    x := make(chan int)
    close(x)
    x <- 1  // Panics here
}
```

**Bad:**
```
My large application crashes when I do things.
```

## Creating a Feature Request

### Title Format

```
[Feature] <Brief description>

Examples:
[Feature] Add support for Solidity 0.9
[Feature] Implement dark mode theme
[Feature] Add keyboard shortcuts customization
```

### Structure

**1. Problem Statement**
```
As a smart contract developer, I want to be able to 
deploy contracts to multiple networks simultaneously 
so that I can save time when testing on different chains.
```

**2. Proposed Solution**
```
Add a "Multi-Deploy" feature that:
1. Allows selecting multiple networks in the deploy dialog
2. Shows progress for each deployment
3. Reports success/failure per network
```

**3. Alternatives Considered**
```
- Using shell scripts to deploy sequentially (current workaround)
- Using external deployment tools (loses IDE integration)
```

**4. Additional Context**
```
This would be particularly useful for:
- Testing on testnets before mainnet
- Deploying to multiple chains (Ethereum, Polygon, BSC)
- CI/CD integration
```

## Issue Labels

We use labels to categorize issues:

| Label | Meaning | Use When |
|-------|---------|----------|
| `bug` | Something is broken | Confirmed defect |
| `enhancement` | New feature | Feature request |
| `docs` | Documentation | Doc improvements |
| `help wanted` | External help needed | Good for contributors |
| `good first issue` | Easy for newcomers | Simple, well-defined |
| `priority::high` | Urgent | Security, crashes |
| `question` | Need clarification | Unclear if bug or usage |

## Issue Lifecycle

```
Opened → Triage → Accepted/Declined → In Progress → Closed
   ↓         ↓            ↓              ↓
 Labels   Assigned    Milestone      Fixed/Completed
```

### Triage Process

1. **New Issue** → Gets `status::triage` label
2. **Within 48 hours** → Maintainer reviews
3. **Decision**:
   - Accepted → `status::in-progress` or added to backlog
   - Declined → Closed with explanation
   - Needs info → `needs-info` label added

### Response Times

| Priority | First Response | Resolution |
|----------|----------------|------------|
| Critical | 24 hours | 72 hours |
| High | 48 hours | 1 week |
| Medium | 1 week | 2 weeks |
| Low | 2 weeks | Next release |

## Getting Your Issue Resolved Faster

### Do

✅ Provide minimal reproduction steps
✅ Include full error messages
✅ Share environment details
✅ Be responsive to follow-up questions
✅ Test with latest version
✅ Use issue templates
✅ One issue per report

### Don't

❌ Create duplicate issues
❌ Comment "+1" or "me too" (use reactions)
❌ Report multiple bugs in one issue
❌ Delete the issue template
❌ Use issues for support questions (use Discussions)

## After Submission

### What to Expect

1. **Automatic**: Labels applied, triage begins
2. **Within 48 hours**: Initial maintainer response
3. **Ongoing**: Questions for clarification
4. **Resolution**: Fix committed or issue closed

### How to Help

- **Answer questions**: Clarify your issue promptly
- **Test fixes**: Try proposed solutions
- **Close if resolved**: When your issue is fixed
- **Update if changed**: If problem evolves

## Special Issue Types

### Security Issues

**DO NOT** create public issues for security vulnerabilities.

Instead:
1. Email security@rechain.ai
2. Include reproduction steps
3. Allow 90 days before public disclosure
4. We'll credit you in the security advisory

See [SECURITY.md](../SECURITY.md)

### Performance Issues

Include:
- CPU/memory profiles if possible
- Dataset size or code complexity
- Hardware specifications
- Comparison with previous versions

### Documentation Issues

Specify:
- Which page/file has the issue
- What's incorrect or missing
- Suggested correction

## Questions?

- General questions: [GitHub Discussions](https://github.com/rechain-ai/rechain-ide/discussions)
- Real-time chat: [Discord](https://discord.gg/rechain)
- Email: dev@rechain.ai (for sensitive matters)

## Quick Reference

| Issue Type | Template | Response Time |
|------------|----------|---------------|
| Bug | Bug Report | 48 hours |
| Feature | Feature Request | 1 week |
| Security | Email security@rechain.ai | 24 hours |
| Question | Discussions | Community |
