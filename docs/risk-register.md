# Risk Register

## MVP Risks
- Integration risk: API contracts not enforced across components.
- Performance risk: Multi-LLM merge introduces latency spikes.
- Security risk: Sandbox policy gaps allow data leakage.
- UX risk: Diff application in IDE is inconsistent across clients.
- Infra risk: Distributed cache invalidation inconsistencies.

## Mitigations
- Enforce schema validation in CI.
- Benchmark merge paths and set timeouts.
- Threat model and policy testing for sandbox.
- Shared diff application library.
- Add cache versioning and TTL guards.
