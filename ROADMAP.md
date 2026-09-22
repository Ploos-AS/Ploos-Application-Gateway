# PAG Roadmap

## M0 — Foundation
- Define PAG identity and security model.
- Establish repository layout.
- Document trusted core and gateway boundaries.
- Require independently installable/removable protocol gateways.
- Define installed/enabled/allowed semantics.
- Add configuration example.
- Add build/test skeleton and CI.

## M1 — Core control plane
- Configuration parser and validation.
- Zone model and policy engine.
- External gateway discovery/registration.
- Gateway capability and version negotiation.
- Gateway lifecycle and health.
- Fail-closed handling for missing/disabled/incompatible gateways.
- Structured logging and health/readiness.
- nftables integration design and dry-run output.
- Prove that core builds and runs without any protocol gateway installed.

## M2 — First useful gateways
- `pag-dns` as a separately packaged DNS application gateway.
- `pag-http` as a separately packaged explicit HTTP proxy.
- Policy decisions and audit trail.
- IPv4/IPv6 qualification.
- Network-namespace integration tests.
- Anti-bypass qualification.

## M3 — Protocol expansion
- Separately packaged `pag-smtp`, `pag-ftp`, and `pag-irc`.
- Per-gateway limits and hardening.

## M4 — Production hardening
- Privilege/capability minimization.
- seccomp / namespace isolation.
- Resource limits and parser fuzzing.
- Prometheus metrics.
- Upgrade and rollback strategy.
- Policy-derived minimal appliance/image builds.

## Later
- Explicitly designed HTTPS interception, if implemented.
- MQTT, SIP, NTP and SOCKS gateways.
- Management API and optional web UI.
- Appliance images.
