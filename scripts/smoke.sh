#!/usr/bin/env bash
set -euo pipefail
base="${DICOM_BASE_URL:-http://127.0.0.1:8080}"
curl -fsS "$base/healthz" | grep -q 'ok'
curl -fsS "$base/readyz" | grep -q 'ready'
curl -fsS "$base/api/v1/instances?limit=10" | grep -q 'items'
curl -fsS "$base/api/v1/destinations" | grep -q 'archive'
curl -fsS "$base/api/v1/audit/export" | grep -q 'events'
echo "smoke ok"
