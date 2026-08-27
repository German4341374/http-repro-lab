#!/usr/bin/env sh
set -eu

cleanup() {
  if [ -n "${MOCK_PID:-}" ]; then kill "$MOCK_PID" 2>/dev/null || true; fi
  if [ -n "${PROD_PID:-}" ]; then kill "$PROD_PID" 2>/dev/null || true; fi
}
trap cleanup EXIT INT TERM

mkdir -p .repro-workspace/integration
go run ./cmd/mock-api >.repro-workspace/integration/mock.log 2>&1 &
MOCK_PID=$!
i=0
until curl --fail --silent http://127.0.0.1:9090/health >/dev/null; do
  i=$((i + 1)); [ "$i" -lt 30 ] || { echo 'mock API did not become healthy' >&2; exit 1; }; sleep 1
done
go run ./cmd/http-repro reproduce fixtures/echo-request.json --target http://127.0.0.1:9090 --allow-private --allow-write --output .repro-workspace/integration/response.json >/dev/null
grep -q '"statusCode": 200' .repro-workspace/integration/response.json
go run ./cmd/http-repro analyze fixtures/auth-401.har --output .repro-workspace/integration/report >/dev/null
test -f .repro-workspace/integration/report/index.html

MOCK_API_ADDR=127.0.0.1:9091 MOCK_API_VARIANT=production go run ./cmd/mock-api >.repro-workspace/integration/production.log 2>&1 &
PROD_PID=$!
i=0
until curl --fail --silent http://127.0.0.1:9091/health >/dev/null; do
  i=$((i + 1)); [ "$i" -lt 30 ] || { echo 'production mock did not become healthy' >&2; exit 1; }; sleep 1
done
if go run ./cmd/http-repro compare fixtures/environment-request.json --target-a http://127.0.0.1:9090 --target-b http://127.0.0.1:9091 --allow-private --output .repro-workspace/integration/comparison.json; then
  echo 'expected environment mismatch' >&2
  exit 1
else
  code=$?
  [ "$code" -eq 1 ] || exit "$code"
fi
grep -Fq 'header.content-type' .repro-workspace/integration/comparison.json
grep -Fq 'json.$.id' .repro-workspace/integration/comparison.json
