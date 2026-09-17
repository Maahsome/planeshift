#!/usr/bin/env bash

set -euo pipefail

PUSHCUT_URL="https://api.pushcut.io/DjzWWLXJrWaKXVTPqAL1L/notifications/AI%20Attention"

MESSAGE='planeshift needs attention'

JSON_PAYLOAD=$(jq -n \
    --arg title "AI Complete - planeshift" \
    --arg text "$MESSAGE" \
    '{
        title: $title,
        text: $text,
    }'
)

curl -s -X POST "$PUSHCUT_URL" \
    -H "Content-Type: application/json" \
    -d "$JSON_PAYLOAD" > /dev/null

