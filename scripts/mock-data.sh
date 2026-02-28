#!/usr/bin/env bash
# Populate ~/.local/state/ccl/ with mock workers and logs for demo screenshots.
# Usage: ./scripts/mock-data.sh        # create mock data
#        ./scripts/mock-data.sh clean   # remove mock data

set -euo pipefail

STATE_DIR="${CCL_STATE_DIR:-$HOME/.local/state/ccl}"
mkdir -p "$STATE_DIR"

# IDs (look like unix timestamps, sorted chronologically)
IDS=(
  1740700001  # done - refactored auth
  1740700002  # done - added tests
  1740700003  # error - migration failed
  1740700004  # working - building API
  1740700005  # working - implementing search
  1740700006  # pending - dark mode
  1740700007  # pending - docs update
)

if [[ "${1:-}" == "clean" ]]; then
  for id in "${IDS[@]}"; do
    rm -f "$STATE_DIR/${id}.json" "$STATE_DIR/${id}.log"
  done
  echo "Cleaned mock data."
  exit 0
fi

now=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Helper: write a worker JSON file
# Args: id status dir task created [started] [finished] [pid] [session]
write_worker() {
  local id=$1 status=$2 dir=$3 task=$4
  local created=${5:-} started=${6:-} finished=${7:-} pid=${8:-0} session=${9:-}

  python3 -c "
import json, sys
w = {'id': '$id', 'status': '$status', 'directory': '$dir', 'task': '$task'}
pid = $pid
if pid > 0: w['pid'] = pid
session = '$session'
if session: w['session_id'] = session
created = '$created'
if created: w['created_at'] = created
started = '$started'
if started: w['started_at'] = started
finished = '$finished'
if finished: w['finished_at'] = finished
with open('$STATE_DIR/${id}.json', 'w') as f:
    json.dump(w, f, indent=2)
    f.write('\n')
"
}

# Helper: write NDJSON log lines
write_log() {
  local id=$1
  shift
  printf '%s\n' "$@" > "$STATE_DIR/${id}.log"
}

# --- Worker 1: done - refactored auth middleware ---
write_worker 1740700001 done \
  "$HOME/projects/api-gateway" \
  "Refactor auth middleware to support JWT refresh tokens" \
  "2026-02-28T09:15:00Z" "2026-02-28T09:15:02Z" "2026-02-28T09:22:47Z" \
  0 "a1b2c3d4-1111-4000-8000-000000000001"

write_log 1740700001 \
  '{"type":"assistant","content":"I'\''ll refactor the auth middleware to support JWT refresh tokens. Let me start by examining the current implementation."}' \
  '{"type":"tool_use","name":"Read"}' \
  '{"type":"result","subtype":"success"}' \
  '{"type":"assistant","content":"The current middleware only validates access tokens. I need to add refresh token rotation logic and update the token validation pipeline."}' \
  '{"type":"tool_use","name":"Edit"}' \
  '{"type":"result","subtype":"success"}' \
  '{"type":"tool_use","name":"Edit"}' \
  '{"type":"result","subtype":"success"}' \
  '{"type":"assistant","content":"Now let me add the refresh endpoint and update the middleware chain."}' \
  '{"type":"tool_use","name":"Write"}' \
  '{"type":"result","subtype":"success"}' \
  '{"type":"tool_use","name":"Bash"}' \
  '{"type":"result","subtype":"success"}' \
  '{"type":"assistant","content":"All tests pass. The auth middleware now supports:\n- JWT access token validation\n- Automatic refresh token rotation\n- Secure httpOnly cookie handling\n- Token revocation on logout"}'

# --- Worker 2: done - added integration tests ---
write_worker 1740700002 done \
  "$HOME/projects/user-service" \
  "Add integration tests for the user registration flow" \
  "2026-02-28T09:30:00Z" "2026-02-28T09:30:03Z" "2026-02-28T09:41:15Z" \
  0 "a1b2c3d4-2222-4000-8000-000000000002"

write_log 1740700002 \
  '{"type":"assistant","content":"I'\''ll add integration tests for the user registration flow. Let me examine the existing test setup."}' \
  '{"type":"tool_use","name":"Glob"}' \
  '{"type":"result","subtype":"success"}' \
  '{"type":"tool_use","name":"Read"}' \
  '{"type":"result","subtype":"success"}' \
  '{"type":"assistant","content":"I see the test harness uses testcontainers for Postgres. I'\''ll add tests covering:\n1. Successful registration\n2. Duplicate email handling\n3. Password validation\n4. Email verification flow"}' \
  '{"type":"tool_use","name":"Write"}' \
  '{"type":"result","subtype":"success"}' \
  '{"type":"tool_use","name":"Bash"}' \
  '{"type":"result","subtype":"success"}' \
  '{"type":"assistant","content":"All 12 integration tests pass. Coverage for the registration flow is now at 94%."}'

# --- Worker 3: error - database migration failed ---
write_worker 1740700003 error \
  "$HOME/projects/inventory-svc" \
  "Write migration to add soft-delete columns to products table" \
  "2026-02-28T10:00:00Z" "2026-02-28T10:00:02Z" "2026-02-28T10:03:18Z" \
  0 "a1b2c3d4-3333-4000-8000-000000000003"

write_log 1740700003 \
  '{"type":"assistant","content":"I'\''ll create a migration to add soft-delete support to the products table."}' \
  '{"type":"tool_use","name":"Read"}' \
  '{"type":"result","subtype":"success"}' \
  '{"type":"tool_use","name":"Write"}' \
  '{"type":"result","subtype":"success"}' \
  '{"type":"assistant","content":"Let me run the migration to verify it works."}' \
  '{"type":"tool_use","name":"Bash"}' \
  '{"type":"result","subtype":"error"}' \
  '{"type":"assistant","content":"The migration failed — there'\''s a conflicting index on the products table from a previous migration. The existing idx_products_sku unique index needs to be updated to include the deleted_at column for the soft-delete pattern to work correctly."}'

# --- Worker 4: working - building REST API ---
write_worker 1740700004 working \
  "$HOME/projects/wreccless" \
  "Add REST API endpoints for worker management" \
  "2026-02-28T10:10:00Z" "2026-02-28T10:10:01Z" "" \
  0 "a1b2c3d4-4444-4000-8000-000000000004"

write_log 1740700004 \
  '{"type":"assistant","content":"I'\''ll add REST API endpoints for managing workers programmatically. Let me first review the existing CLI structure."}' \
  '{"type":"tool_use","name":"Glob"}' \
  '{"type":"result","subtype":"success"}' \
  '{"type":"tool_use","name":"Read"}' \
  '{"type":"result","subtype":"success"}' \
  '{"type":"tool_use","name":"Read"}' \
  '{"type":"result","subtype":"success"}' \
  '{"type":"assistant","content":"I'\''ll create an HTTP server package with endpoints for listing, creating, and managing workers. Starting with the router and handlers."}' \
  '{"type":"tool_use","name":"Write"}' \
  '{"type":"result","subtype":"success"}' \
  '{"type":"tool_use","name":"Write"}' \
  '{"type":"result","subtype":"success"}' \
  '{"type":"assistant","content":"Router and handlers are set up. Now adding the list and create endpoints with proper JSON serialization..."}'

# --- Worker 5: working - implementing full-text search ---
write_worker 1740700005 working \
  "$HOME/projects/content-platform" \
  "Implement full-text search with Elasticsearch integration" \
  "2026-02-28T10:15:00Z" "2026-02-28T10:15:03Z" "" \
  0 "a1b2c3d4-5555-4000-8000-000000000005"

write_log 1740700005 \
  '{"type":"assistant","content":"I'\''ll implement full-text search using Elasticsearch. Let me check the current data models and search requirements."}' \
  '{"type":"tool_use","name":"Read"}' \
  '{"type":"result","subtype":"success"}' \
  '{"type":"tool_use","name":"Read"}' \
  '{"type":"result","subtype":"success"}' \
  '{"type":"assistant","content":"I see the content model has title, body, tags, and author fields. I'\''ll set up an ES index with appropriate analyzers and implement the search service."}' \
  '{"type":"tool_use","name":"Write"}' \
  '{"type":"result","subtype":"success"}'

# --- Worker 6: pending - dark mode support ---
write_worker 1740700006 pending \
  "$HOME/projects/dashboard-ui" \
  "Add dark mode toggle with system preference detection" \
  "2026-02-28T10:20:00Z"

write_log 1740700006  # empty log for pending

# --- Worker 7: pending - update API docs ---
write_worker 1740700007 pending \
  "$HOME/projects/api-gateway" \
  "Update OpenAPI spec and generate new API documentation" \
  "2026-02-28T10:25:00Z"

write_log 1740700007  # empty log for pending

echo "Created ${#IDS[@]} mock workers in $STATE_DIR"
echo ""
echo "Workers:"
echo "  1740700001  done     Refactor auth middleware"
echo "  1740700002  done     Add integration tests"
echo "  1740700003  error    Database migration (failed)"
echo "  1740700004  working  REST API endpoints"
echo "  1740700005  working  Full-text search"
echo "  1740700006  pending  Dark mode toggle"
echo "  1740700007  pending  API docs update"
echo ""
echo "Run 'ccl ui' to view the TUI dashboard."
echo "Run '$0 clean' to remove mock data."
