#!/bin/bash
set -e

# ── Config ────────────────────────────────────────────────────────────────────
VPS_USER="root"
VPS_HOST="23.94.151.125"
VPS_DIR="/root/telegram-bots-golang/telestory-api-based"
BINARY_NAME="application"
# ─────────────────────────────────────────────────────────────────────────────

echo "▶ Building for Linux..."
GOOS=linux GOARCH=amd64 go build -o "$BINARY_NAME" ./cmd/server/main.go

echo "▶ Copying files to VPS..."
ssh "$VPS_USER@$VPS_HOST" "mkdir -p $VPS_DIR/migrations"
scp "$BINARY_NAME"        "$VPS_USER@$VPS_HOST:$VPS_DIR/"
scp ".env.prod"           "$VPS_USER@$VPS_HOST:$VPS_DIR/"
scp migrations/*.sql      "$VPS_USER@$VPS_HOST:$VPS_DIR/migrations/"

echo "▶ Restarting server..."
ssh "$VPS_USER@$VPS_HOST" bash <<EOF
  cd $VPS_DIR
  pkill -f "$BINARY_NAME" || true
  nohup ./$BINARY_NAME -env prod > server.log 2>&1 &
  echo "Server started (PID \$!)"
EOF

echo "✓ Deploy complete"
