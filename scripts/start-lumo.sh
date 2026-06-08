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

cd "$LUMO_DIR"

# Wait for auth file to be provided before starting Lumo
while [ ! -f "$AUTH_FILE" ]; do
  echo "Waiting for Lumo auth file at $AUTH_FILE ..."
  sleep 10
done

cp "$AUTH_FILE" "$LUMO_DIR/auth.json"

kill_stale_lumo() {
  for attempt in 1 2 3; do
    pids=$(lsof -ti:3333 2>/dev/null || true)
    if [ -z "$pids" ]; then
      break
    fi
    echo "Killing stale process on port 3333 (attempt $attempt): $pids"
    echo "$pids" | xargs -r kill -9 2>/dev/null || true
    sleep 1
  done
  sleep 2
  lsof -ti:3333 2>/dev/null && echo "WARNING: Port 3333 still in use after cleanup attempts" || true
}

kill_stale_lumo

if ! node lumo.js ghost:true; then
  echo "Lumo process exited unexpectedly; keeping container alive for recovery."
  exec tail -f /dev/null
fi
