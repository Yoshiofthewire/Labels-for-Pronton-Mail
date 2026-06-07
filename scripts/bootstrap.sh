#!/bin/sh
set -eu

mkdir -p "${CONFIG_DIR:-/lumo_lab/config}" "${LOG_DIR:-/lumo_lab/logs}" "${STATE_DIR:-/lumo_lab/state}"

if [ ! -f "${CONFIG_DIR:-/lumo_lab/config}/admin.env" ]; then
  user="${BOOTSTRAP_ADMIN_USER:-admin}"
  pass="${BOOTSTRAP_ADMIN_PASS:-ChangeMeNow123!}"
  {
    echo "ADMIN_USER=${user}"
    echo "ADMIN_PASS=${pass}"
    echo "MUST_CHANGE_PASSWORD=true"
  } >"${CONFIG_DIR:-/lumo_lab/config}/admin.env"
  chmod 600 "${CONFIG_DIR:-/lumo_lab/config}/admin.env"
  echo "Generated first-run admin credentials in config volume"
  echo "Username: ${user}"
  echo "Password: ${pass}"
  echo "Password change is required on first login"
fi
