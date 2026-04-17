#!/usr/bin/env bash
set -euo pipefail

API_BASE="${API_BASE:-http://localhost:3000}"
DOMAIN_ID="${DOMAIN_ID:-}"

if [[ -z "${DOMAIN_ID}" ]]; then
  echo "[info] Registering domain"
  RESPONSE="$(curl -fsS -X POST "${API_BASE}/domains/register" \
    -H 'content-type: application/json' \
    -d '{"clientName":"demo-client","domain":"example.com","upstreamUrl":"http://localhost:3001"}')"
  echo "[debug] register response: ${RESPONSE}"
  DOMAIN_ID="$(printf '%s' "$RESPONSE" | sed -n 's/.*"domainId":"\([^"]*\)".*/\1/p')"
fi

if [[ -z "${DOMAIN_ID}" ]]; then
  echo "[error] domain id not found" >&2
  exit 1
fi

echo "[info] Checking status for domain id: ${DOMAIN_ID}"
curl -fsS "${API_BASE}/domains/${DOMAIN_ID}/status" | tee /dev/stderr >/dev/null

echo "[info] Triggering DNS verification (expected true only when TXT is configured)"
curl -fsS -X POST "${API_BASE}/domains/${DOMAIN_ID}/verify-dns" | tee /dev/stderr >/dev/null

echo "[info] Phase 1 integration flow executed"
