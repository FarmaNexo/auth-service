#!/usr/bin/env bash
# validate-p1-login-evidence.sh
#
# Valida que el login persiste evidencia legal (LPDP) en el refresh token:
# ip_address y user_agent reales del request. Antes se guardaban en blanco.
#
# Requiere el stack local arriba y auth-service reiniciado con el código nuevo.
set -euo pipefail

AUTH=http://localhost:4001
TS=$(date +%s)
EMAIL="e2e_p1_${TS}@farmanexo.com"
PASS="TestPassword123!"
UA="FarmaNexo-P1-Validator/1.0"
OUT=$(mktemp); trap 'rm -f "$OUT"' EXIT

ok()   { echo -e "\033[1;32m✅ $*\033[0m"; }
fail() { echo -e "\033[1;31m❌ $*\033[0m"; cat "$OUT" 2>/dev/null | head -20; exit 1; }

req() {
  curl -sS -X "$1" -o "$OUT" -w "%{http_code}" \
    -H "Content-Type: application/json" -H "User-Agent: $UA" \
    ${3:+--data "$3"} "$2"
}

code=$(req POST "$AUTH/api/v1/auth/register" \
  "{\"email\":\"$EMAIL\",\"password\":\"$PASS\",\"full_name\":\"P1 Test\",\"phone\":\"999111222\",\"accepted_terms\":true,\"accepted_privacy\":true,\"marketing_opt_in\":true}")
[[ "$code" = "201" ]] || fail "register devolvió $code"
ok "usuario registrado ($EMAIL)"

code=$(req POST "$AUTH/api/v1/auth/login" "{\"email\":\"$EMAIL\",\"password\":\"$PASS\"}")
[[ "$code" = "200" ]] || fail "login devolvió $code"
ok "login 200 (User-Agent enviado: $UA)"

sleep 1
ROW=$(docker exec farmanexo-postgres psql -U admin -d auth_db -t -A -F'|' \
  -c "select coalesce(ip_address,''), coalesce(user_agent,'') from auth.refresh_tokens t join auth.users u on u.id=t.user_id where u.email='$EMAIL' order by t.created_at desc limit 1")
IP="${ROW%%|*}"; AGENT="${ROW#*|}"
echo "   refresh_tokens → ip_address='$IP'  user_agent='$AGENT'"
[[ -n "$IP" ]]        || fail "ip_address quedó vacío"
[[ "$AGENT" = "$UA" ]] || fail "user_agent no coincide (esperaba '$UA')"

echo
echo -e "\033[1;32m═══ P1 login evidence OK — ip_address y user_agent persistidos ═══\033[0m"
