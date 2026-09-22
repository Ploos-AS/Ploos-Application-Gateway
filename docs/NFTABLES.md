# PAG nftables integration

## M1 scope

M1 implements **render-only** nftables support. PAG can generate a ruleset for inspection, tests, and later validation, but M1 does not apply it to the host.

This is intentional: firewall mutation is introduced only after the generated security properties are covered by integration tests.

## Security invariant

PAG proxy policy must never turn into ordinary forwarding.

The generated `inet pag` forward chain therefore has a default `drop` policy. M1 emits no generic forwarding `accept` rules for application traffic. A proxied session is expected to terminate locally at the separately installed gateway, which creates its own upstream session.

This establishes the anti-bypass baseline before transparent redirection is added.

## Future steps

- dedicated input rules for gateway listeners;
- transparent redirect/TProxy where appropriate;
- established/related handling with explicit reasoning;
- interface/zone sets;
- atomic ruleset validation and application;
- rollback;
- network-namespace integration tests;
- IPv4/IPv6 qualification.

Generated identifiers are validated before rendering. Configuration text is never copied blindly into nft syntax.
