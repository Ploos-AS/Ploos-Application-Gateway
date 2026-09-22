# PAG security invariants

These invariants are architectural requirements, not optional defaults.

1. **Default deny** — unmatched inter-zone traffic is denied.
2. **No generic application accept** — M1 policy supports `proxy` and `deny`; a plain forwarding `accept` action is intentionally absent.
3. **No protocol implementation in core** — protocol parsers and proxies remain separately installable gateways.
4. **No implicit fallback** — missing, disabled, incompatible, unreachable, malformed, or unhealthy required gateways fail closed.
5. **No interface ambiguity** — one interface cannot belong to multiple PAG zones.
6. **Explicit zone boundary** — inter-zone rules name both source and destination zones.
7. **Anti-bypass** — traffic required to use an application gateway must not become ordinary forwarded traffic.
8. **Untrusted configuration** — identifiers are validated before they are rendered into firewall syntax.
9. **Process separation** — protocol parsing is not dynamically loaded into `pag-core`.
10. **Dry-run before mutation** — M1 renders nftables rules but does not modify the host firewall.

Future milestones may extend the policy language, but changes must preserve these properties or explicitly document and test a replacement security model.
