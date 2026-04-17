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

Part 3 continuation now also includes:

- Analytics time-series endpoint (`GET /analytics/time-series`)
- Analytics top IPs endpoint (`GET /analytics/top-ips`)
- Domains listing endpoint (`GET /domains`)

## Dashboard scaffold (Part 4 start)

The Nuxt dashboard now includes:

- API layer routes for Auth, Domains, and Analytics (`services/dashboard-nuxt/server/api/**`)
- Pinia stores: `auth` and `domains`
- Dashboard UI with stats cards, traffic chart bars, and top-IP insights

## Deployment scaffold (Part 5 start)

Deployment compose scaffolds now include:

- Core services: Proxy, API, Dashboard, Nginx, Redis, Postgres
- Observability services: Prometheus and Grafana
- Prometheus scrape config at `deployment/prometheus/prometheus.yml`

## Part 6: Security, environment, and data lifecycle

Environment variable surface now includes:

- `JWT_SECRET` (or `JWT_SECRET_FILE`)
- `REDIS_PASSWORD` (or `REDIS_PASSWORD_FILE`)
- `DATABASE_DSN` (or `DATABASE_DSN_FILE`)
- `ORIGIN_SECRET` (or `ORIGIN_SECRET_FILE`)

Security and operations hardening updates include:

- Docker secrets scaffold in `deployment/docker-compose.prod.yml`
- SSL certificate mount path scaffold for edge nginx (`deployment/nginx/certs`)
- Internal/private network segmentation for stateful and app services in production compose
- Startup enforcement for JWT secret in API and secret/env overrides in proxy

Database schema lifecycle additions:

- Core tables: `clients`, `domains`, `proxy_rules`, `request_logs`
- TimescaleDB migration scaffold in `db/migrations/002_phase6_timescale.sql` for:
  - hypertable conversion
  - compression policy
  - retention policy

## Final summary

ShieldProxy is a multi-layer security reverse proxy with:

- Rate limiting
- Bot detection
- Geo filtering
- JS challenge
- JWT validation
- Analytics dashboard

These are implementation foundations intended for iterative hardening in subsequent phases.
