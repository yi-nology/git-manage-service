#!/bin/bash
# Start both frontend and backend for development
# Usage: bash scripts/dev.sh

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BACKEND_BIN="/tmp/gms-dev"

echo "[1/3] Building backend..."
cd "$PROJECT_ROOT"
CGO_ENABLED=1 go build -o "$BACKEND_BIN" ./cmd/server/main.go

echo "[2/3] Starting backend on :12345..."
"$BACKEND_BIN" --mode=http &
BACKEND_PID=$!
echo "  Backend PID: $BACKEND_PID"

echo "[3/3] Starting frontend on :5173..."
cd "$PROJECT_ROOT/frontend"
exec npx vite --host 0.0.0.0 --port 5173
