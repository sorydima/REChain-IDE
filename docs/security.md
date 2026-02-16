# Security and Privacy

This document outlines baseline security and privacy principles.

## Principles
- Privacy-first: local execution by default when feasible.
- Least privilege: model drivers run with minimal capabilities.
- Policy enforcement: all access is gated by policy decisions.
- Auditing: critical operations emit structured logs.

## Sandbox
- All code execution happens in a sandbox.
- Sandboxes prohibit unauthorized network access by default.
- Sandbox escape attempts are logged and blocked.

## Data handling
- Sensitive data is tagged and access-controlled.
- Cached data has TTL and is invalidated on policy changes.
