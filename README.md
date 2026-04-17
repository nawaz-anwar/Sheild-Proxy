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

These are implementation foundations intended for iterative hardening in subsequent phases.
