#!/usr/bin/env bash
set -euo pipefail

if [ -n "${GOOGLE_SERVICES_JSON:-}" ]; then
  cp "$GOOGLE_SERVICES_JSON" ./google-services.json
  echo "Copied google-services.json"
fi

if [ -n "${GOOGLE_SERVICES_INFO_PLIST:-}" ]; then
  cp "$GOOGLE_SERVICES_INFO_PLIST" ./GoogleService-Info.plist
  echo "Copied GoogleService-Info.plist"
fi
