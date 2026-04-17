# Redis Key Conventions (Phase 1 placeholders)

- `shield:ratelimit:{domain}:{ip}:{window}` (TTL: 60s)
- `shield:challenge:{domain}:{ip}` (TTL: 900s)

Phase 1 does not enforce challenge mode yet; these keys reserve naming for Phase 2.
