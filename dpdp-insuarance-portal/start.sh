#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PORT=3020
PID_FILE="$SCRIPT_DIR/.insurance.pid"

# Kill every process listening on a given TCP port
kill_port() {
  local port="$1"
  local pids
  pids=$(lsof -ti tcp:"$port" 2>/dev/null || true)
  if [ -n "$pids" ]; then
    echo "  Freeing port $port (PID(s): $(echo "$pids" | tr '\n' ' '))..."
    echo "$pids" | xargs kill -9 2>/dev/null || true
    sleep 0.3
  fi
}

# ── stop ──────────────────────────────────────────────────────────────────────
stop() {
  if [ -f "$PID_FILE" ]; then
    local pid
    pid=$(cat "$PID_FILE")
    if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null || true
      echo "  Stopped PID $pid"
    fi
    rm -f "$PID_FILE"
  fi

  # Always kill by port — catches orphaned processes
  kill_port "$PORT"

  echo "Insurance portal stopped."
}

# ── start ─────────────────────────────────────────────────────────────────────
start() {
  stop

  if [ ! -d "$SCRIPT_DIR/node_modules" ]; then
    echo "Installing dependencies..."
    (cd "$SCRIPT_DIR" && npm install)
  fi

  echo "Starting insurance portal on http://localhost:$PORT ..."
  cd "$SCRIPT_DIR"
  node server.js > "$SCRIPT_DIR/insurance.log" 2>&1 &
  echo $! > "$PID_FILE"
  cd - > /dev/null

  echo ""
  echo "Insurance portal is running."
  echo "  Portal  → http://localhost:$PORT"
  echo "  Log     → insurance.log"
  echo ""
  echo "Run './start.sh stop' to stop."
}

# ── dispatch ──────────────────────────────────────────────────────────────────
case "${1:-start}" in
  stop)  stop  ;;
  start) start ;;
  *)
    echo "Usage: $0 [start|stop]"
    exit 1
    ;;
esac
