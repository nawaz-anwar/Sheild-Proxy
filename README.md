# Sheild-Proxy

ShieldProxy monorepo scaffold with three deployable services:

- `services/proxy-go` (Go edge proxy)
- `services/api-nest` (NestJS API)
- `services/dashboard-nuxt` (Vue/Nuxt dashboard)

## Compose entrypoint

A top-level `docker-compose.yml` is included for local orchestration and maps directly to the service roles:

- `proxy`
- `api`
- `dashboard`
- plus `postgres` and `redis`

## Go proxy scaffold (Part 2 foundations)

The Go proxy now includes scaffolded modules for:

- YAML configuration with Redis/PostgreSQL/JWT/GeoIP sections
- Multi-layer domain store shape (L1 memory with L2/L3 provider interfaces)
- Sliding-window rate limiting (domain + IP)
- GeoIP filtering (lookup abstraction + policy)
- Header analysis with SEO bot allow behavior
- Filter engine pipeline (`Allowlist → Ban → Rate Limit → Geo → Headers`)
- JS challenge (SHA256 puzzle challenge issue/verify)
- JWT cookie token issuance/validation with IP+UA binding
- Reverse handler flow scaffold (`domain lookup → token validation → filter → challenge/forward`)
- Origin request HMAC signature injection headers
- Main entry runtime init hooks for Redis + PostgreSQL dependencies

## NestJS API scaffold (Part 3 start)

The API now exposes scaffold responsibilities for:

- Auth (`POST /auth/register`, `POST /auth/login`)
- Domain management (`/domains/register`, `/domains/:id/status`, `/domains/:id/verify-dns`, `/domains/:id/rules`)
- Rules service layer backed by `proxy_rules`
- Analytics overview endpoint (`GET /analytics/overview`)

These are implementation foundations intended for iterative hardening in subsequent phases.
