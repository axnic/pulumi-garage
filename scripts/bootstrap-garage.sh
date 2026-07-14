#!/usr/bin/env bash
# Bootstraps a single-node cluster layout on the Garage instance at
# GARAGE_ADMIN_ENDPOINT, entirely over the Admin API (no `docker exec`/
# `docker compose exec`) - so it works identically whether it's called from
# the host, from inside the devcontainer (where the "garage" container isn't
# in a compose project this script can address by name/exec), or in CI.
#
# Idempotent: safe to call again on an already-bootstrapped instance (e.g. a
# second `make dev-up` without an intervening `make dev-down`); it detects
# an existing role assignment and skips straight through.
#
# --single-node (which does this automatically, inside the container) only
# exists from Garage v2.3.0 onward, and this repo's compatibility matrix
# tests back to v2.0.0 (see README.md's Compatibility table), so this
# script - not the flag - is what actually works across every supported
# version.
set -euo pipefail

: "${GARAGE_ADMIN_ENDPOINT:?GARAGE_ADMIN_ENDPOINT must be set}"
: "${GARAGE_ADMIN_TOKEN:?GARAGE_ADMIN_TOKEN must be set}"

ZONE="dc1"
CAPACITY=1000000000 # 1 GB, in bytes - the admin API wants a byte count, not a suffixed string like the garage CLI's -c flag does

auth=(-H "Authorization: Bearer $GARAGE_ADMIN_TOKEN")

# Garage's Admin API returns pretty-printed JSON ("key": "value", with a
# space after the colon) - these helpers tolerate that instead of assuming
# compact "key":"value", which silently extracts nothing otherwise.
json_string_field() { grep -oE "\"$1\"[[:space:]]*:[[:space:]]*\"[^\"]*\"" | head -1 | sed -E 's/.*:[[:space:]]*"([^"]*)"/\1/'; }
json_number_field() { grep -oE "\"$1\"[[:space:]]*:[[:space:]]*[0-9]+" | head -1 | grep -oE '[0-9]+$'; }

status="$(curl -sf "${auth[@]}" "$GARAGE_ADMIN_ENDPOINT/v2/GetClusterStatus")"
node_id="$(echo "$status" | json_string_field id)"

if echo "$status" | grep -qE '"role"[[:space:]]*:[[:space:]]*\{'; then
	echo "Garage layout already bootstrapped, nothing to do."
	exit 0
fi

echo "Bootstrapping single-node layout for $node_id..."

update_response="$(curl -sf -X POST "${auth[@]}" -H "Content-Type: application/json" \
	-d "{\"roles\":[{\"id\":\"$node_id\",\"zone\":\"$ZONE\",\"capacity\":$CAPACITY,\"tags\":[]}]}" \
	"$GARAGE_ADMIN_ENDPOINT/v2/UpdateClusterLayout")"

current_version="$(echo "$update_response" | json_number_field version)"
next_version=$((current_version + 1))

curl -sf -X POST "${auth[@]}" -H "Content-Type: application/json" \
	-d "{\"version\":$next_version}" \
	"$GARAGE_ADMIN_ENDPOINT/v2/ApplyClusterLayout" >/dev/null

echo "Garage layout bootstrapped (version $next_version)."
