# PAG Roadmap

## M0 — Foundation
- Define PAG identity and security model.
- Establish repository layout.
- Document trusted core and module boundaries.
- Define initial module contract.
- Add configuration example.
- Add build/test skeleton and CI.

## M1 — Core control plane
- Configuration parser and validation.
- Zone model and policy engine.
- Module lifecycle.
- Structured logging and health/readiness.
- nftables integration design and dry-run output.

## M2 — First useful gateway
- DNS application gateway.
- Explicit HTTP proxy.
- Policy decisions and audit trail.
- IPv4/IPv6 qualification.
- Network-namespace integration tests.

## M3 — Protocol expansion
- SMTP, FTP and IRC gateways.
- Per-module limits and hardening.

## M4 — Production hardening
- Privilege/capability minimization.
- seccomp / namespace isolation.
- Resource limits and parser fuzzing.
- Prometheus metrics.
- Upgrade and rollback strategy.

## Later
- Explicitly designed HTTPS interception, if implemented.
- MQTT, SIP, NTP and SOCKS modules.
- Management API and optional web UI.
- Appliance images.
