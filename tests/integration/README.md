# PAG integration tests

M1 integration qualification will use Linux network namespaces to model:

```text
client namespace <-> PAG/router namespace <-> upstream namespace
```

The first qualification target is the anti-bypass invariant:

- direct client-to-upstream forwarding is blocked;
- IPv4 and IPv6 are both tested;
- a configured application gateway is reachable only through its intended PAG path;
- removing or stopping a required gateway does not open forwarding;
- invalid configuration cannot weaken the generated default-deny ruleset.

These tests require Linux network namespaces and nftables privileges and therefore remain separate from ordinary unit tests.

The M1 renderer is intentionally non-mutating until this harness is in place.
