# PAG — Ploos Application Gateway

**PAG** is a modern, open-source application-level proxy firewall for Linux.

> **Proxy first. Default deny. Protocol aware. Modular by design.**

PAG does not merely inspect supported application protocols. It terminates, validates, applies policy to, and proxies them across security zones.

## Minimal core

Protocol implementations do **not** live in `pag-core`. Each application gateway is an independently installable and removable component.

A deployment should install only the protocol gateways its policy requires:

```text
pag-core
├── pag-dns       optional
├── pag-http      optional
├── pag-smtp      optional
├── pag-ftp       optional
└── pag-irc       optional
```

For example, a firewall allowing only DNS and HTTP across a zone boundary needs `pag-core`, `pag-dns`, and `pag-http`. SMTP, FTP, and IRC parser/proxy code need not exist on that system.

PAG distinguishes three states:

- **installed** — the gateway implementation exists on the system;
- **enabled** — the gateway is available to PAG;
- **allowed** — explicit policy permits traffic through that gateway.

A missing gateway never causes fallback forwarding. If policy requires `pag-smtp` and that module is unavailable, the connection is denied.

## Security model

- Default deny between security zones.
- No protocol implementation in core.
- Protocol gateways are independently installable/removable.
- Fail closed when a required gateway is missing or unhealthy.
- nftables supplies L3/L4 containment, redirection, and anti-bypass.
- PAG gateways form the L7 security boundary.
- IPv4 and IPv6 are first-class requirements.
- Structured logging and automation-friendly configuration are architectural requirements.

## M0

M0 establishes the project identity, security model, modularity rules, repository skeleton, architecture documentation, roadmap, and CI baseline.

See [ROADMAP.md](ROADMAP.md) and [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## License

Software in this repository is licensed under the MIT License unless a file or directory states otherwise.
