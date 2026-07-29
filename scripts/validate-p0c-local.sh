#!/usr/bin/env bash
# validate-p0c-local.sh
#
# Valida P0-C en auth-service contra localhost:4001:
#  1) Registro transaccional: tras register+login, los 3 consentimientos LPDP
#     (terms/privacy/marketing) están persistidos (user + consents commiteados juntos).
#  2) Breach detection (refresh token rotation reuse):
#       - login → refresh1
#       - refresh(refresh1) → refresh2 (refresh1 queda revocado)
#       - refresh(refresh1) de nuevo → 401 + revoca TODA la familia
#       - refresh(refresh2) → 401 (la sesión legítima también se invalidó)
set -euo pipefail

AUTH=http://localhost:4001
TS=$(date +%s)
EMAIL="e2e_p0c_${TS}@farmanexo.com"
PASS="TestPassword123!"

OUT_DIR=$(mktemp -d); trap 'rm -rf "$OUT_DIR"' EXIT
LAST="$OUT_DIR/last.json"

ok()   { echo -e "\033[1;32m✅ $*\033[0m"; }
info() { echo -e "\n\033[1;36m── $*\033[0m"; }
fail() { echo -e "\033[1;31m❌ $*\033[0m"; [[ -f "$LAST" ]] && python -m json.tool "$LAST" 2>/dev/null | head -20; exit 1; }

pyget() { cat "$LAST" | python -c "import json,sys; d=json.load(sys.stdin); v=$1; print('null' if v is None else v)"; }

# req METHOD URL BODY BEARER → echoes http_code; body in $LAST
req() {
  local method=$1 url=$2 body=${3:-} bearer=${4:-}
  local -a a=(-sS -X "$method" -o "$LAST" -w "%{http_code}" -H "Content-Type: application/json" -H "Accept: application/json")
  [[ -n "$bearer" ]] && a+=(-H "Authorization: Bearer $bearer")
  [[ -n "$body" ]]   && a+=(--data "$body")
  curl "${a[@]}" "$url" 2>/dev/null || echo "000"
}

# ─── 1. Registro transaccional ───────────────────────────────────────────────
info "1. Register (user + 3 consents en una transacción)"
code=$(req POST "$AUTH/api/v1/auth/register" "{\"email\":\"$EMAIL\",\"password\":\"$PASS\",\"full_name\":\"P0C Test\",\"phone\":\"999111222\",\"accepted_terms\":true,\"accepted_privacy\":true,\"marketing_opt_in\":true}")
[[ "$code" = "201" ]] || fail "register devolvió $code (esperaba 201)"
ok "usuario registrado (201)"

info "2. Login → access1 + refresh1"
code=$(req POST "$AUTH/api/v1/auth/login" "{\"email\":\"$EMAIL\",\"password\":\"$PASS\"}")
[[ "$code" = "200" ]] || fail "login devolvió $code"
ACCESS1=$(pyget "d['datos']['access_token']")
REFRESH1=$(pyget "d['datos']['refresh_token']")
[[ -n "$REFRESH1" && "$REFRESH1" != "null" ]] || fail "no se obtuvo refresh1"
ok "login OK (refresh1 len=${#REFRESH1})"

info "3. Verificar consents persistidos (prueba que la tx de registro commiteó el batch)"
code=$(req GET "$AUTH/api/v1/auth/me/consents" "" "$ACCESS1")
[[ "$code" = "200" ]] || fail "GET consents devolvió $code"
TOTAL=$(pyget "d['datos']['total']")
TYPES=$(pyget "sorted(set(c['consent_type'] for c in d['datos']['consents']))")
[[ "$TOTAL" -ge 3 ]] || fail "esperaba ≥3 consents, hay $TOTAL (¿tx no commiteó el batch?)"
ok "consents persistidos: total=$TOTAL tipos=$TYPES"

# ─── Breach detection ────────────────────────────────────────────────────────
info "4. Refresh con refresh1 → refresh2 (refresh1 queda revocado)"
code=$(req POST "$AUTH/api/v1/auth/refresh" "{\"refresh_token\":\"$REFRESH1\"}")
[[ "$code" = "200" ]] || fail "primer refresh devolvió $code (esperaba 200)"
REFRESH2=$(pyget "d['datos']['refresh_token']")
[[ -n "$REFRESH2" && "$REFRESH2" != "null" && "$REFRESH2" != "$REFRESH1" ]] || fail "refresh2 inválido o igual a refresh1"
ok "rotación OK (refresh2 ≠ refresh1)"

info "5. REUSO de refresh1 (ya revocado) → debe disparar breach detection (401)"
code=$(req POST "$AUTH/api/v1/auth/refresh" "{\"refresh_token\":\"$REFRESH1\"}")
[[ "$code" = "401" ]] || fail "reuso de refresh1 devolvió $code (esperaba 401)"
MSG=$(pyget "d['meta']['mensajes'][0]['mensaje']")
ok "reuso rechazado (401): $MSG"

info "6. refresh2 (sesión legítima) → debe estar TAMBIÉN revocado por el breach (401)"
code=$(req POST "$AUTH/api/v1/auth/refresh" "{\"refresh_token\":\"$REFRESH2\"}")
[[ "$code" = "401" ]] || fail "refresh2 devolvió $code — la familia NO se revocó (breach detection incompleta)"
ok "refresh2 invalidado (401) — toda la familia de tokens fue revocada"

echo
echo -e "\033[1;32m═══ P0-C OK ═══\033[0m"
echo "  - Registro transaccional: user + 3 consents commiteados juntos ✅"
echo "  - Breach detection: reuso de token revocado → revoca toda la familia ✅"
