# Onboarding Guide (Phase 1 MVP)

1. Register a domain in API (`POST /domains/register`).
2. Create DNS TXT record: `_shield-verify.<domain>` = returned `dnsToken`.
3. Trigger verification (`POST /domains/:id/verify-dns`).
4. Once active, route traffic to Shield Proxy edge.
5. At origin, only trust traffic with `X-Shield-Verified: true` and expected `X-Shield-Client-ID`.
