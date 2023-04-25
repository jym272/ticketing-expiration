#!/usr/bin/env bash

set -eou pipefail

#{
#  "orders.created": {
#    "id": 44417,
#    "expiresAt": "2023-04-08T02:56:00-03:00"
#  }
#}

# update the order expiration date -> add 10 seconds since now using date and sed
# current time in RFC format (RFC 3339)
#now=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
# update the order expiration date -> add 10 seconds since now using date and sed
#sed -e "s/\"expiresAt\": \".*\"/\"expiresAt\": \"$now\"/" order.json > test.json

# 2023-04-08T09:15:55.506Z
jq '."orders.created".expiresAt = (now + 10 | todateiso8601)' order.json > test.json
export NATS_URL="nats://localhost:4222"
nats pub orders.created "$(jq -r '@json' test.json)"