#!/bin/sh
set -eu

LUMO_DIR="${LUMO_LOCAL_DIR:-/opt/lumo-api-v2}"
AUTH_FILE="${LUMO_AUTH_FILE:-/lumo_lab/config/lumo-auth.json}"

if [ "${LUMO_LOCAL_ENABLED:-true}" != "true" ]; then
  echo "Local Lumo process disabled by LUMO_LOCAL_ENABLED"
  exec tail -f /dev/null
fi

if [ ! -d "$LUMO_DIR" ]; then
  echo "Lumo directory not found: $LUMO_DIR"
  exec tail -f /dev/null
fi

if [ -f "$AUTH_FILE" ]; then
  cp "$AUTH_FILE" "$LUMO_DIR/auth.json"
fi

if [ ! -f "$LUMO_DIR/auth.json" ]; then
  echo "Missing Lumo auth file at $AUTH_FILE"
  echo "Place auth file there or disable local Lumo process."
  exec tail -f /dev/null
fi

cd "$LUMO_DIR"
exec node lumo.js
