# Sheild-Proxy

Phase 1 MVP monorepo for a core reverse-proxy onboarding flow.

## Services
- `services/proxy-go`: host lookup + reverse proxy + verified headers + metrics/health
- `services/api-nest`: domain registration, DNS verification, status, basic metrics/health
- `services/dashboard-nuxt`: basic dashboard shell and health/metrics endpoints

## Shared
- `config/platform.env.example` and `config/platform.yaml`: single source of truth for Phase 1 config
- `db/migrations/001_phase1_mvp.sql`: PostgreSQL schema for clients/domains/proxy rules/request logs
- `deployment/`: nginx/openresty edge + compose profiles
- `docs/`: onboarding + operations guidance

## Phase scope
This repository currently implements **Phase 1 (MVP core proxy)** with placeholders for later phases.
