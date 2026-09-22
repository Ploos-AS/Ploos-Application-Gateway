# pag-dns

`pag-dns` is the first real PAG protocol gateway.

## M2 prototype

- separate executable; no DNS parser/proxy code is linked into `pag-core`;
- UDP and TCP DNS listeners;
- separate upstream session;
- minimum DNS header validation;
- exactly one question required;
- bounded UDP/TCP message size;
- connection/read deadlines for TCP;
- malformed messages are dropped rather than forwarded.

This is deliberately a small clean-room DNS gateway, not a general-purpose recursive resolver.

## Next hardening

- full question-name parser with compression safety;
- opcode/type/class policy;
- response transaction/question matching;
- UDP truncation and TCP fallback behavior;
- EDNS size policy;
- rate/connection limits;
- structured audit events;
- control socket implementing the PAG gateway contract;
- namespace qualification for UDP and TCP;
- fuzz tests.

The default upstream in the prototype is only a convenience for manual development; production policy must make upstream selection explicit.
