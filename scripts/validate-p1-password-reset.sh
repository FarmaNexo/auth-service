#!/usr/bin/env bash
# validate-p1-password-reset.sh
#
# Valida el flujo forgot-password / reset-password contra localhost:4001:
#  register -> login(old) -> forgot -> [token del correo en Mailpit] -> reset ->
#  login(old)=401, login(new)=200, refresh(old)=401 (sesiones revocadas).
#  + edge cases: email inexistente=200 (anti-enumeración), token basura=400.
#
# En local el correo se envía por SMTP a Mailpit (contenedor del Helper). El token se
# extrae del correo capturado vía la API de Mailpit (http://localhost:8025). No se
# envían correos reales: en Dev/Prod el proveedor es SES (EMAIL_PROVIDER=ses).
set -euo pipefail

AUTH=http://localhost:4001
MAILPIT="${MAILPIT:-http://localhost:8025}"
TS=$(date +%s); EMAIL="reset_${TS}@farmanexo.com"; OLD="OldPassword123!"; NEW="NuevaClave456!"

ok()   { echo -e "\033[1;32m✅ $*\033[0m"; }
fail() { echo -e "\033[1;31m❌ $*\033[0m"; exit 1; }
code() { curl -sS -m8 -o /dev/null -w "%{http_code}" "$@" 2>/dev/null; }
jget() { python -c "import json,sys;print(json.load(sys.stdin).get('datos',{}).get('$1',''))" 2>/dev/null; }

[[ "$(code $AUTH/health)" = "200" ]] || fail "auth-service no responde en $AUTH"

[[ "$(code -X POST $AUTH/api/v1/auth/register -H 'Content-Type: application/json' --data "{\"email\":\"$EMAIL\",\"password\":\"$OLD\",\"full_name\":\"Reset Test\",\"phone\":\"999000111\",\"accepted_terms\":true,\"accepted_privacy\":true,\"marketing_opt_in\":false}")" = "201" ]] || fail "register"
ok "usuario registrado ($EMAIL)"

REF=$(curl -sS -m8 -X POST $AUTH/api/v1/auth/login -H 'Content-Type: application/json' --data "{\"email\":\"$EMAIL\",\"password\":\"$OLD\"}" | jget refresh_token)
[[ -n "$REF" ]] || fail "login inicial"
ok "login con password original + refresh token emitido"

[[ "$(code -X POST $AUTH/api/v1/auth/forgot-password -H 'Content-Type: application/json' --data "{\"email\":\"$EMAIL\"}")" = "200" ]] || fail "forgot-password"
ok "forgot-password aceptado (200)"

sleep 2
MID=$(curl -sS -m5 "$MAILPIT/api/v1/search?query=to:$EMAIL" 2>/dev/null | python -c "import json,sys; m=json.load(sys.stdin).get('messages',[]); print(m[0]['ID'] if m else '')")
[[ -n "$MID" ]] || fail "no se encontró el correo en Mailpit (¿Mailpit arriba? ¿EMAIL_PROVIDER=smtp en local?)"
TOKEN=$(curl -sS -m5 "$MAILPIT/api/v1/message/$MID" 2>/dev/null | python -c "import json,sys,re; d=json.load(sys.stdin); s=(d.get('HTML') or '')+(d.get('Text') or ''); m=re.search(r'token=([a-f0-9]+)', s); print(m.group(1) if m else '')")
[[ -n "$TOKEN" ]] || fail "no se pudo extraer el token del correo en Mailpit"
ok "token de reset extraído del correo en Mailpit (len=${#TOKEN})"

[[ "$(code -X POST $AUTH/api/v1/auth/reset-password -H 'Content-Type: application/json' --data "{\"token\":\"$TOKEN\",\"new_password\":\"$NEW\"}")" = "200" ]] || fail "reset-password"
ok "contraseña restablecida (200)"

[[ "$(code -X POST $AUTH/api/v1/auth/login -H 'Content-Type: application/json' --data "{\"email\":\"$EMAIL\",\"password\":\"$OLD\"}")" = "401" ]] || fail "la password vieja NO debería seguir sirviendo"
ok "password vieja rechazada (401)"

[[ "$(code -X POST $AUTH/api/v1/auth/login -H 'Content-Type: application/json' --data "{\"email\":\"$EMAIL\",\"password\":\"$NEW\"}")" = "200" ]] || fail "la password nueva debería servir"
ok "password nueva aceptada (200)"

[[ "$(code -X POST $AUTH/api/v1/auth/refresh -H 'Content-Type: application/json' --data "{\"refresh_token\":\"$REF\"}")" = "401" ]] || fail "el refresh token viejo debería estar revocado"
ok "sesiones previas revocadas (refresh viejo → 401)"

[[ "$(code -X POST $AUTH/api/v1/auth/forgot-password -H 'Content-Type: application/json' --data "{\"email\":\"noexiste_${TS}@farmanexo.com\"}")" = "200" ]] || fail "anti-enumeración"
ok "email inexistente → 200 genérico (anti-enumeración)"
[[ "$(code -X POST $AUTH/api/v1/auth/reset-password -H 'Content-Type: application/json' --data '{"token":"tokenfalso","new_password":"CualquierClave99!"}')" = "400" ]] || fail "token inválido"
ok "token inválido → 400"

echo
echo -e "\033[1;32m═══ Password reset OK ═══\033[0m"
