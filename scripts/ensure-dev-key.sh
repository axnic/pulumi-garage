#!/usr/bin/env bash
# Creates a default "dev" S3 access key against GARAGE_ADMIN_ENDPOINT if one
# doesn't already exist - used by the devcontainer's postCreateCommand so
# there's a working S3 credential the moment the container finishes
# creating, no manual step needed.
#
# Idempotent: if "dev" already exists, prints a note instead of recreating
# it - Garage never returns a key's secretAccessKey a second time, so a
# repeat run can't recover it, only a fresh key could.
set -euo pipefail

: "${GARAGE_ADMIN_ENDPOINT:?GARAGE_ADMIN_ENDPOINT must be set}"
: "${GARAGE_ADMIN_TOKEN:?GARAGE_ADMIN_TOKEN must be set}"

existing="$(curl -sf -H "Authorization: Bearer $GARAGE_ADMIN_TOKEN" "$GARAGE_ADMIN_ENDPOINT/v2/GetKeyInfo?search=dev" || true)"
if [[ "$existing" == *'"accessKeyId"'* ]]; then
	echo "Dev access key 'dev' already exists. Its secret was only ever printed"
	echo "once, at creation - recreate the devcontainer for a fresh key+secret."
	exit 0
fi

response="$(curl -sf -X POST -H "Authorization: Bearer $GARAGE_ADMIN_TOKEN" -H "Content-Type: application/json" \
	-d '{"name":"dev"}' "$GARAGE_ADMIN_ENDPOINT/v2/CreateKey")"

# Garage's Admin API returns pretty-printed JSON ("key": "value", with a
# space after the colon) - this tolerates that instead of assuming compact
# "key":"value", which silently extracts nothing otherwise.
json_string_field() { grep -oE "\"$1\"[[:space:]]*:[[:space:]]*\"[^\"]*\"" | head -1 | sed -E 's/.*:[[:space:]]*"([^"]*)"/\1/'; }

access_key_id="$(echo "$response" | json_string_field accessKeyId)"
secret_access_key="$(echo "$response" | json_string_field secretAccessKey)"

echo "Created a default 'dev' S3 access key:"
echo "  GARAGE_DEV_ACCESS_KEY_ID=$access_key_id"
echo "  GARAGE_DEV_SECRET_ACCESS_KEY=$secret_access_key"
echo "It isn't granted any bucket permissions yet - grant it some with a"
echo "BucketKeyPermission resource, or create your own Key as part of"
echo "whatever Pulumi program you're testing instead."
