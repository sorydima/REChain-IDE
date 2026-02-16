# Security Checklist for Contributors

This checklist helps contributors ensure their changes meet security requirements.

## Pre-Submission Checklist

### Code Security

- [ ] No hardcoded secrets (passwords, API keys, tokens)
- [ ] No private keys committed
- [ ] No `.env` files with real credentials
- [ ] Input validation implemented
- [ ] Output encoding/escaping where needed
- [ ] SQL injection prevention (if applicable)
- [ ] XSS prevention (if applicable)
- [ ] CSRF protection (if applicable)

### Dependencies

- [ ] No new dependencies with known vulnerabilities
- [ ] Dependencies are from trusted sources
- [ ] Minimum required permissions for dependencies
- [ ] License compatibility verified

### Authentication & Authorization

- [ ] Authentication checks in place
- [ ] Authorization verified for sensitive operations
- [ ] Session management secure
- [ ] Token handling secure (no logging, proper storage)

### Data Handling

- [ ] Sensitive data encrypted at rest (if applicable)
- [ ] Sensitive data encrypted in transit (HTTPS/TLS)
- [ ] PII handled according to regulations
- [ ] Data minimization practiced
- [ ] Secure data deletion implemented

### Configuration

- [ ] Security-sensitive config in environment variables
- [ ] Default configurations are secure
- [ ] No debug mode in production
- [ ] Error messages don't leak sensitive info

## File Review Checklist

### Files to Check

- [ ] New/changed Go files
- [ ] Configuration files
- [ ] Docker files
- [ ] Scripts
- [ ] Documentation (for security notes)

### What to Look For

#### In Code
```
DON'T:
- password = "hardcoded"
- apiKey := "sk-123456"
- exec(userInput)
- eval(userInput)
- log.Debugf("Token: %s", token)
- return err (with sensitive details)
- http (without TLS)

DO:
- password := os.Getenv("PASSWORD")
- apiKey := os.Getenv("API_KEY")
- sanitize input before use
- use parameterized queries
- log.Debugf("Request received")
- return errors.New("authentication failed")
- https only
```

#### In Configuration
```
DON'T:
- debug: true (in production)
- auth: disabled
- cors: "*"
- log_level: debug (in production)

DO:
- debug: false
- auth: required
- cors: ["https://example.com"]
- log_level: info
```

## Testing Security

- [ ] Unit tests for security functions
- [ ] Integration tests for auth flows
- [ ] Fuzzing for input validation
- [ ] Security scans pass (`make security-check`)

## Documentation

- [ ] Security considerations documented
- [ ] API authentication documented
- [ ] Deployment security notes included
- [ ] Breaking security changes noted

## CI/CD Security

- [ ] Secrets not in CI logs
- [ ] Secrets properly masked
- [ ] No sensitive data in artifacts
- [ ] Branch protection rules followed

## Common Vulnerabilities to Avoid

### OWASP Top 10

1. **Injection** - Validate and sanitize all inputs
2. **Broken Authentication** - Use strong auth mechanisms
3. **Sensitive Data Exposure** - Encrypt sensitive data
4. **XML External Entities** - Disable XXE if not needed
5. **Broken Access Control** - Enforce authorization
6. **Security Misconfiguration** - Secure defaults
7. **Cross-Site Scripting** - Encode output
8. **Insecure Deserialization** - Validate input
9. **Using Components with Known Vulnerabilities** - Update deps
10. **Insufficient Logging & Monitoring** - Log security events

### Go-Specific

- [ ] `defer resp.Body.Close()` always used
- [ ] `http.Server` with proper timeouts
- [ ] `crypto/rand` for randomness (not `math/rand`)
- [ ] `regexp` not used for complex parsing
- [ ] `unsafe` package avoided unless necessary
- [ ] Goroutine leaks prevented

## Security Testing Commands

```bash
# Run security scan
make security-check

# Check for secrets
git-secrets --scan

# Check dependencies
govulncheck ./...

# Run linter with security rules
golangci-lint run --enable=gosec

# Check for hardcoded secrets
./scripts/security-scan.sh
```

## Reporting Security Issues

If you find a security issue:

**DO NOT** create a public issue.

Instead:
- Email: security@rechain.ai
- Include reproduction steps
- Allow time for fix before disclosure

## Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Guide](https://golang.org/doc/security)
- [GitHub Security](https://docs.github.com/en/code-security)
- [GitLab Security](https://docs.gitlab.com/ee/user/application_security/)

## Review Process

Security-sensitive changes require:
1. Security team review
2. Additional testing
3. Documentation approval

Label PRs with `security` for priority review.

---

**Remember**: Security is everyone's responsibility. When in doubt, ask!
