#!/usr/bin/env bash
set -euo pipefail

ROOT=$(go env GOROOT)
cp "$ROOT/lib/wasm/wasm_exec.js" "$(dirname "$0")/../web/wasm_exec.js"

GOOS=js GOARCH=wasm go build -o "$(dirname "$0")/../web/image_zoo.wasm" ./cmd/wasm

echo "Built web/image_zoo.wasm"
