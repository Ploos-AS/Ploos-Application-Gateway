# PAG Architecture

PAG is an application-level gateway firewall. Supported application traffic is terminated by a protocol-aware proxy rather than merely forwarded after packet inspection.

## Security boundaries
1. **Network containment** — Linux routing, namespaces and nftables constrain traffic.
2. **PAG core** — validates configuration, evaluates policy, manages modules and emits audit events.
3. **Protocol modules** — parse, validate and proxy a specific application protocol.
4. **Observability plane** — consumes structured events and metrics without controlling forwarding.

## Traffic model
A connection crossing a protected zone boundary is denied unless policy permits it. For a supported protocol, policy directs traffic to its PAG module. The module establishes a separate upstream connection after validation and policy evaluation.

## Core/module contract
A module has a stable identifier, listeners, protocol configuration, lifecycle operations, session metadata, policy requests/decisions, audit events and health state. M1 will prototype the interface before an ABI/API is frozen.

## Failure behavior
Security-relevant parse errors, invalid configuration, unavailable mandatory modules and ambiguous policy fail closed by default. Any availability exception must be explicit policy.

## TLS
PAG distinguishes tunnelling from interception. HTTPS CONNECT can be proxied without decrypting TLS. TLS interception is not an M0 feature and, if added, requires a separate trust model, certificate lifecycle, policy and auditing.

## nftables
nftables supplies packet-level containment, redirection and anti-bypass rules. It does not substitute for PAG protocol parsing and validation.

## Isolation
Protocol modules should not require unrestricted host privileges. Later milestones will use independent identities, resource limits, capability minimization, namespaces and seccomp where appropriate.
