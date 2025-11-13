#!/usr/bin/env bash

# Start all backend services for the social network project (Linux/macOS)
# Usage: bash ./start-services.sh

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DB_PATH="${REPO_ROOT}/social_network.db"
AUTH_SERVICE_URL="http://localhost:8081"
LOG_DIR="${REPO_ROOT}/logs"

SERVICES=(
  "Auth Service|8081|services/auth|false"
  "Users Service|8082|services/users|true"
  "Posts Service|8083|services/posts|true"
  "Groups Service|8084|services/groups|true"
  "Chat Service|8085|services/chat|true"
  "Notifications Service|8086|services/notifications|true"
)

if ! command -v go >/dev/null 2>&1; then
  echo "Error: Go command not found. Install Go and ensure it is in your PATH." >&2
  exit 1
fi

if [[ ! -f "$DB_PATH" ]]; then
  echo "Error: Database file not found at '$DB_PATH'." >&2
  echo "Run migrations or update DB_PATH inside start-services.sh if your DB lives elsewhere." >&2
  exit 1
fi

mkdir -p "$LOG_DIR"

pids=()

cleanup() {
  if [[ ${#pids[@]} -gt 0 ]]; then
    echo -e "\nStopping services..."
    for pid in "${pids[@]}"; do
      if kill -0 "$pid" >/dev/null 2>&1; then
        kill "$pid" >/dev/null 2>&1 || true
      fi
    done
    wait "${pids[@]}" 2>/dev/null || true
  fi
}

trap cleanup EXIT INT TERM

start_service() {
  local name="$1"
  local port="$2"
  local rel_path="$3"
  local needs_auth="$4"

  local service_dir="${REPO_ROOT}/${rel_path}"

  if [[ ! -d "$service_dir" ]]; then
    echo "Warning: Skipping ${name} (directory '${service_dir}' not found)."
    return
  fi

  local log_file="${LOG_DIR}/${name// /_}.log"

  (
    cd "$service_dir"
    export DATABASE_PATH="$DB_PATH"
    if [[ "$needs_auth" == "true" ]]; then
      export AUTH_SERVICE_URL="$AUTH_SERVICE_URL"
    else
      unset AUTH_SERVICE_URL
    fi

    echo "[$(date '+%H:%M:%S')] Starting ${name} on port ${port}..."
    exec go run main.go
  ) > >(tee "$log_file") 2>&1 &

  local pid=$!
  pids+=("$pid")
  echo "  → ${name} running (PID ${pid}). Logs: ${log_file}"
}

echo "Starting backend services using database ${DB_PATH}"
echo "Logs will stream to ${LOG_DIR}/SERVICE_NAME.log"
echo

for service in "${SERVICES[@]}"; do
  IFS="|" read -r name port path needs_auth <<< "$service"
  start_service "$name" "$port" "$path" "$needs_auth"
done

echo
echo "All services launched. Press Ctrl+C to stop them."

wait -n || wait
