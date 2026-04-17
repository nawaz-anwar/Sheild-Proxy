# Operations Runbook (Phase 1 MVP)

## Health checks
- Edge: `GET /healthz`
- Proxy: `GET /healthz`, `GET /metrics`
- API: `GET /healthz`, `GET /metrics`
- Dashboard: `GET /health`

## Incident triage
1. Check edge/nginx routing and DNS.
2. Check domain status in API.
3. Check proxy logs for host lookup failures.
4. Check Postgres for domains marked `active`.

## Origin hardening
- Block direct origin access at firewall.
- Allow only edge CIDRs.
- Validate `X-Shield-Verified` and `X-Shield-Client-ID` on origin.

## Security checklist
- Rotate secrets (`jwt_secret`, `origin_secret`, `redis_password`, `database_dsn`, `postgres_password`, `grafana_admin_password`) on a regular schedule.
- Use Docker secrets for production (`deployment/docker-compose.prod.yml` expects externally managed secrets).
- Enable SSL at edge by mounting certificates into `deployment/nginx/certs` and enabling TLS server directives.
- Restrict DB access to internal network paths only; do not publish Postgres/Redis ports in production.
