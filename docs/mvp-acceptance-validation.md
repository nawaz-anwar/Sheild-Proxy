# Phase 1 MVP Acceptance Validation

## Goal alignment
- End-to-end onboarding readiness: **implemented**
  - Domain register endpoint: `POST /domains/register`
  - Domain status endpoint: `GET /domains/:id/status`
  - DNS verification endpoint: `POST /domains/:id/verify-dns`
- End-to-end proxy flow: **implemented**
  - Edge Nginx forwards to Go proxy
  - Proxy performs active host lookup and reverse proxy routing
  - Proxy injects `X-Shield-Verified` + `X-Shield-Client-ID`

## Validation checklist
1. Apply DB migration `db/migrations/001_phase1_mvp.sql`.
2. Start stack with `deployment/docker-compose.dev.yml`.
3. Run `tests/integration/phase1_flow.sh`.
4. Verify proxy service health: `GET /healthz`; metrics: `GET /metrics`.
5. Verify API health: `GET /healthz`; metrics: `GET /metrics`.
6. Verify dashboard health: `GET /api/health`.

## Current status
- Build/lint/test executed for all three services.
- Direct dependency advisories resolved for newly added packages.
