#!/usr/bin/env bash

set -euo pipefail

PUSHCUT_URL="https://api.pushcut.io/DjzWWLXJrWaKXVTPqAL1L/notifications/AI%20Attention"

MESSAGE=$(
  awk '
    /^## / {
      in_files = 0
      files = ""
    }
    /^### Files Changed/ {
      in_files = 1
      files = ""
      next
    }
    in_files && /^### / {
      in_files = 0
    }
    in_files && /^- / {
      line = $0
      sub(/^- /, "", line)
      gsub(/`/, "", line)
      if (files != "") {
        files = files "\n"
      }
      files = files line
    }
    END {
      print files
    }
  ' CHAT_LOG.md
)

JSON_PAYLOAD=$(jq -n \
    --arg title "AI Complete - modron" \
    --arg text "$MESSAGE" \
    '{
        title: $title,
        text: $text,
    }'
)

curl -s -X POST "$PUSHCUT_URL" \
    -H "Content-Type: application/json" \
    -d "$JSON_PAYLOAD" > /dev/null

