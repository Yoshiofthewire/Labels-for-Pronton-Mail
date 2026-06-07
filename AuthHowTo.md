# Auth Setup How-To

This guide explains how to collect the credentials and files needed for this project to access Proton Mail and Lumo.

## What You Need

- Proton auth, using one of these modes:
  - Username/password mode: PROTON_USERNAME, PROTON_PASSWORD, optional PROTON_TOTP
  - Token mode: PROTON_UID, PROTON_ACCESS_TOKEN, PROTON_REFRESH_TOKEN
- Lumo auth session file for local Lumo API V2:
  - /lumo_lab/config/lumo-auth.json

## 1) Proton Mail: Username and Password Mode

Use this mode if you want the app to log in directly with account credentials.

1. Open Proton Mail and confirm you can sign in normally.
2. Decide which mailbox account this app should read.
3. Fill these environment variables in your local .env file:
   - PROTON_USERNAME
   - PROTON_PASSWORD
   - PROTON_TOTP (only if your account uses TOTP 2FA)
4. Leave token variables empty when using this mode:
   - PROTON_UID
   - PROTON_ACCESS_TOKEN
   - PROTON_REFRESH_TOKEN

Notes:
- If 2FA is enabled and PROTON_TOTP is missing, login will fail.
- TOTP codes are short-lived. Update value if needed.

## 2) Proton Mail: Token Mode

Use this mode if you already have Proton API session values.

1. Collect these values from your Proton API session workflow:
   - PROTON_UID
   - PROTON_ACCESS_TOKEN
   - PROTON_REFRESH_TOKEN
2. Fill them in your local .env file.
3. Leave username/password variables empty when using this mode:
   - PROTON_USERNAME
   - PROTON_PASSWORD
   - PROTON_TOTP

Optional:
- Set PROTON_API_HOST only if you need a non-default Proton API host.

## 3) Lumo: Create the Auth File

This project runs local Lumo API V2 in the same container. It expects a valid auth session file.

Required path inside the container:
- /lumo_lab/config/lumo-auth.json

How to get it:
1. Follow the login/auth steps from the Lumo API V2 project to generate its auth.json.
2. Save the generated file on your host.
3. Make sure your compose config maps it into this project at:
   - /lumo_lab/config/lumo-auth.json

In this repo, the startup script copies that file into the local Lumo runtime before launch.

## 4) Fill Your .env File

At minimum, set:
- LUMO_API_KEY (if your Lumo route requires it)
- LUMO_BASE_URL (default is local: http://127.0.0.1:3333)
- LUMO_LOCAL_ENABLED=true (for local Lumo process)
- One Proton mode only:
  - Either username/password (+ optional TOTP)
  - Or UID/access/refresh tokens

Then start:

- docker compose up --build -d

## 5) Verify in UI

1. Open the web UI at http://localhost:5866
2. Log in with admin credentials from /lumo_lab/config/admin.env (created on first run).
3. Open Configuration page.
4. Run Lumo Test.
5. If needed, use logs endpoint to troubleshoot:
   - /api/logs?lines=200

## Troubleshooting

- Proton auth fails:
  - Confirm you filled exactly one auth mode.
  - If using 2FA, verify PROTON_TOTP is present and current.
- Lumo local process idle or fails:
  - Confirm lumo-auth.json exists at /lumo_lab/config/lumo-auth.json.
  - Check container logs for lumo process startup messages.
- Labels are not applied:
  - Verify Proton credentials are valid.
  - Ensure labels.allowlist is not empty.
  - Check daemon and API logs.

## Security Tips

- Never commit real credentials or tokens.
- Keep .env private.
- Rotate passwords/tokens if exposed.
