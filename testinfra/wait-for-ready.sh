#!/bin/bash
# Waits for the OpenWrt acceptance VM's JSON-RPC API to become reachable
# (see testinfra/). Polls the /cgi-bin/luci/rpc/auth endpoint with a login
# request until it returns a JSON body containing a "result" field (a login
# token), or until the retry budget is exhausted.
#
# Usage: ./testinfra/wait-for-ready.sh [base-url]
#   base-url  Base URL of the OpenWrt instance, e.g. http://192.168.56.2
#             (defaults to `terraform output -raw openwrt_host` from testinfra/)
#
# Exits 0 once the API responds, 1 if it never does within the retry budget.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

BASE_URL="${1:-}"
if [ -z "$BASE_URL" ]; then
  BASE_URL=$(cd "$SCRIPT_DIR" && terraform output -raw openwrt_host)
fi

RETRIES=60
SLEEP_SECS=2

echo "Waiting for OpenWrt JSON-RPC API at ${BASE_URL} to become ready..."

for i in $(seq 1 "$RETRIES"); do
  RESPONSE=$(curl -sS -m 5 -X POST "${BASE_URL}/cgi-bin/luci/rpc/auth" \
    -H 'Content-Type: application/json' \
    -d '{"id":1,"method":"login","params":["root","root"]}' 2>/dev/null || true)

  if printf '%s' "$RESPONSE" | grep -q '"result"'; then
    echo "OpenWrt JSON-RPC API is ready at ${BASE_URL} (attempt ${i}/${RETRIES})"
    exit 0
  fi

  sleep "$SLEEP_SECS"
done

echo "ERROR: OpenWrt JSON-RPC API at ${BASE_URL} did not become ready after $((RETRIES * SLEEP_SECS))s" >&2
exit 1
