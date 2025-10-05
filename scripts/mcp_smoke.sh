#!/usr/bin/env bash
set -euo pipefail

# mcp_smoke.sh - End-to-end smoke test for Teable MCP over stdio
# Requirements:
# - Go toolchain installed
# - Network access to fetch mcp-go mcptest
# - Postgres/Redis already available per server config

ROOT_DIR="/Users/leven/space/easy/gotama/newapps/server"
export PERMISSIONS_DISABLED="1"

cd "$ROOT_DIR"

# 1) Start MCP server in background (stdio)
#    We capture PID to cleanup on exit
GO_CMD="go run ./cmd/mcp"
$GO_CMD > /dev/null 2>&1 &
MCP_PID=$!
trap 'kill -9 $MCP_PID >/dev/null 2>&1 || true' EXIT
sleep 2

echo "[1/6] MCP server started (PID=$MCP_PID)"

# 2) List tools using mcptest (from mcp-go)
MCPT="go run github.com/mark3labs/mcp-go/mcptest@v0.42.0-beta.1"

echo "[2/6] Listing tools..."
$MCPT --stdio --command "${GO_CMD}" --list-tools || true

# Helper to call a tool and extract id from JSON
call_tool() {
  local tool_name="$1"
  local json_args="$2"
  $MCPT --stdio --command "${GO_CMD}" --call "${tool_name}" --args "${json_args}"
}

extract_id() {
  # read stdin JSON, try to pull common id fields
  python3 - "$@" <<'PY'
import sys, json
s=sys.stdin.read()
try:
  j=json.loads(s)
  # try direct id
  if isinstance(j, dict) and 'id' in j:
    print(j['id'])
    sys.exit(0)
  # try nested data
  for k in ('result','data'):
    if isinstance(j, dict) and k in j and isinstance(j[k], dict) and 'id' in j[k]:
      print(j[k]['id'])
      sys.exit(0)
except Exception:
  pass
print("")
PY
}

# 3) Create Space
echo "[3/6] Creating space..."
SPACE_JSON='{"name":"AI Demo Space","user_id":"u_demo","description":"autotest"}'
SPACE_OUT=$(call_tool createSpace "$SPACE_JSON" || true)
echo "$SPACE_OUT"
SPACE_ID=$(printf "%s" "$SPACE_OUT" | extract_id)
if [ -z "$SPACE_ID" ]; then echo "Failed to parse space id"; exit 1; fi
echo "SPACE_ID=$SPACE_ID"

# 4) Create Base
echo "[4/6] Creating base..."
BASE_JSON=$(printf '{"space_id":"%s","name":"Demo Base","user_id":"u_demo"}' "$SPACE_ID")
BASE_OUT=$(call_tool createBase "$BASE_JSON" || true)
echo "$BASE_OUT"
BASE_ID=$(printf "%s" "$BASE_OUT" | extract_id)
if [ -z "$BASE_ID" ]; then echo "Failed to parse base id"; exit 1; fi
echo "BASE_ID=$BASE_ID"

# 5) Create Table
echo "[5/6] Creating table..."
TABLE_JSON=$(printf '{"base_id":"%s","name":"Tasks","user_id":"u_demo"}' "$BASE_ID")
TABLE_OUT=$(call_tool createTable "$TABLE_JSON" || true)
echo "$TABLE_OUT"
TABLE_ID=$(printf "%s" "$TABLE_OUT" | extract_id)
if [ -z "$TABLE_ID" ]; then echo "Failed to parse table id"; exit 1; fi
echo "TABLE_ID=$TABLE_ID"

# 6) Create Field
echo "[6/6] Creating field (Title)..."
FIELD_JSON=$(printf '{"table_id":"%s","name":"Title","type":"text","required":false,"is_unique":false,"is_primary":false,"field_order":1,"user_id":"u_demo"}' "$TABLE_ID")
FIELD_OUT=$(call_tool createField "$FIELD_JSON" || true)
echo "$FIELD_OUT"

# 7) Create Record
echo "[7/6] Creating record..."
RECORD_JSON=$(printf '{"table_id":"%s","data":{"Title":"Hello MCP"},"user_id":"u_demo"}' "$TABLE_ID")
RECORD_OUT=$(call_tool createRecord "$RECORD_JSON" || true)
echo "$RECORD_OUT"

echo "Done. IDs: SPACE_ID=$SPACE_ID BASE_ID=$BASE_ID TABLE_ID=$TABLE_ID"
