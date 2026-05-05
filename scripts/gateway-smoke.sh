#!/usr/bin/env bash
# Смоук через nginx-шлюз (docker compose up из корня репозитория).
set -euo pipefail

BASE="${GATEWAY_URL:-http://127.0.0.1:8080}"

U1="${U1_UUID:-aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa}"
U2="${U2_UUID:-bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb}"
U3="${U3_UUID:-cccccccc-cccc-cccc-cccc-cccccccccccc}"

echo "== GET $BASE/health"
curl -sS "$BASE/health" | head -c 300; echo
echo "== через префикс /swipe → /v1/candidates"
curl -sS "$BASE/swipe/v1/candidates" -H "X-User-Id: $U1" | head -c 300; echo

echo "== swipe U1 → U2 like"
curl -sS -X POST "$BASE/v1/swipes" \
  -H "Content-Type: application/json" \
  -H "X-User-Id: $U1" \
  -d "{\"swipee_id\":\"$U2\",\"direction\":\"like\"}"
echo

echo "== swipe U2 → U1 like (матч)"
curl -sS -X POST "$BASE/v1/swipes" \
  -H "Content-Type: application/json" \
  -H "X-User-Id: $U2" \
  -d "{\"swipee_id\":\"$U1\",\"direction\":\"like\"}"
echo

echo "== matches U1"
curl -sS "$BASE/v1/matches?limit=10" -H "X-User-Id: $U1"
echo

echo "== duplicate swipe (ожидаем 409)"
code=$(curl -sS -o /tmp/dup.json -w '%{http_code}' -X POST "$BASE/v1/swipes" \
  -H "Content-Type: application/json" \
  -H "X-User-Id: $U1" \
  -d "{\"swipee_id\":\"$U2\",\"direction\":\"dislike\"}")
echo "HTTP $code $(cat /tmp/dup.json)"
