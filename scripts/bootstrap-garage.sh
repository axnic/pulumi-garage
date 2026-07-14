#!/usr/bin/env bash
# Bootstraps a single-node cluster layout on the `garage` service defined in
# docker-compose.yml, once it's up and healthy (`docker compose up --wait`).
#
# This is done from the host rather than inside the container because the
# dxflrs/garage image ships no shell - only the /garage binary - so there's
# nowhere to run a multi-step init script inside the container itself.
#
# Idempotent: safe to call again on an already-bootstrapped instance (e.g. a
# second `make dev-up` without an intervening `make dev-down`); it detects
# an existing role assignment and skips straight through.
#
# --single-node (which does this automatically) only exists from Garage
# v2.3.0 onward, and this repo's compatibility matrix tests back to v2.0.0
# (see README.md's Compatibility table), so this script - not the flag - is
# what actually works across every supported version.
set -euo pipefail

COMPOSE_FILE="${1:-docker-compose.yml}"
SERVICE="garage"
ZONE="dc1"
CAPACITY="1G"

compose_exec() {
	docker compose -f "$COMPOSE_FILE" exec -T "$SERVICE" "$@"
}

status="$(compose_exec /garage status)"
if ! echo "$status" | grep -q "NO ROLE ASSIGNED"; then
	echo "Garage layout already bootstrapped, nothing to do."
	exit 0
fi

node_id="$(compose_exec /garage node id | tail -1 | cut -d@ -f1)"
echo "Bootstrapping single-node layout for $node_id..."

compose_exec /garage layout assign -z "$ZONE" -c "$CAPACITY" "$node_id"

current_version="$(compose_exec /garage layout show | grep -oE 'Current cluster layout version: [0-9]+' | grep -oE '[0-9]+$')"
next_version=$((current_version + 1))
compose_exec /garage layout apply --version "$next_version"

echo "Garage layout bootstrapped (version $next_version)."
