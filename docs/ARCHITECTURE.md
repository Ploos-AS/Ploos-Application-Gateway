# PAG Architecture

PAG is an application-level gateway firewall. Supported application traffic is terminated by a protocol-aware proxy rather than merely forwarded after packet inspection.

## Architecture principle: no protocol implementation in core

`pag-core` contains no DNS, HTTP, SMTP, FTP, IRC, MQTT, SIP, or other application-protocol parser/proxy implementation.

Every protocol gateway is an independently installable and removable component. This minimizes attack surface and allows an appliance or host to contain only the protocol implementations required by its policy.

A deployment that permits only DNS and HTTP can therefore contain:

```text
pag-core
pag-dns
pag-http
```

It does not need SMTP, FTP, IRC, or other protocol parser code.

## Installed, enabled, allowed

PAG treats these as separate states:

1. **Installed** — gateway implementation is present.
2. **Enabled** — gateway is configured and available to core.
3. **Allowed** — an explicit policy permits a flow through that gateway.

Installation or enablement never implicitly grants network access.

## Security boundaries

1. **Network containment** — Linux routing, namespaces and nftables constrain traffic.
2. **PAG core** — validates configuration, evaluates policy, discovers/manages gateways and emits audit events.
3. **Protocol gateways** — independently installed components that parse, validate and proxy one application protocol or tightly related protocol family.
4. **Observability plane** — consumes structured events and metrics without controlling forwarding.

## Traffic model

A connection crossing a protected zone boundary is denied unless policy permits it. For a supported protocol, policy directs traffic to its PAG gateway. The gateway terminates the client-side session and establishes a separate upstream session only as policy permits.

There is no automatic fallback from a required application gateway to ordinary L3/L4 forwarding.

## Core/gateway contract

A gateway exposes a stable identity and capabilities to core. The M1 prototype will cover:

- gateway ID and version;
- supported protocol/capabilities;
- lifecycle and health;
- listener declaration;
- configuration validation;
- session metadata;
- policy requests/decisions;
- structured audit events.

The transport/API is deliberately not frozen in M0. M1 must prove that gateways can remain separately packaged and separately executable before compatibility guarantees are introduced.

## Discovery

Core must not assume a fixed compiled-in set of gateways. Gateway discovery will use explicit configuration/registration. Unknown, absent, disabled, incompatible, or unhealthy gateways referenced by policy are errors and fail closed.

## Failure behavior

Security-relevant parse errors, invalid configuration, unavailable mandatory gateways and ambiguous policy fail closed by default. Availability exceptions must be explicit policy and must never silently bypass an application gateway.

## TLS

PAG distinguishes tunnelling from interception. HTTPS CONNECT can be proxied without decrypting TLS. TLS interception is not an M0 feature and, if added, requires a separate trust model, certificate lifecycle, policy and auditing.

## nftables

nftables supplies packet-level containment, redirection and anti-bypass rules. It does not substitute for PAG protocol parsing and validation. PAG should generate or validate rules that prevent traffic from bypassing a gateway required by policy.

## Isolation

Gateways should run as separate processes rather than dynamically loading untrusted protocol parsing code into core. They should not require unrestricted host privileges. Later milestones will add independent identities, resource limits, capability minimization, namespaces and seccomp where appropriate.

## Packaging

Packaging is part of the security model. Distribution packages and appliance images should allow protocol gateways to be omitted completely. Future image tooling should be able to derive the required gateway set from deployment policy.
