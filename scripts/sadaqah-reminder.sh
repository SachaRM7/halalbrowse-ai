#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${SADAQAH_TRACKER_BASE_URL:-http://localhost:5000}"
INTERNAL_API_TOKEN="${INTERNAL_API_TOKEN:-internal-dev-token}"

curl --fail --silent --show-error \
  -X POST \
  -H "X-Internal-Token: ${INTERNAL_API_TOKEN}" \
  "${BASE_URL}/api/v1/internal/send-reminders"
