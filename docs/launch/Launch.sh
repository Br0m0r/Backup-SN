#!/usr/bin/env bash
set -euo pipefail

# Linux launcher for the social-network stack.
# Mirrors the behaviour of start-services.ps1 using Bash-native tooling.

SKIP_INSTALL=0
SKIP_FRONTEND=0
FRONTEND_HOST="localhost"
FRONTEND_PORT=5173
IS_ROOT=0

if [[ "${EUID:-$(id -u)}" -eq 0 ]]; then
  IS_ROOT=1
fi

while [[ $# -gt 0 ]]; do
  case "$1" in
    --skip-install) SKIP_INSTALL=1 ;;
    --skip-frontend) SKIP_FRONTEND=1 ;;
    --host)
      FRONTEND_HOST="${2:-localhost}"
      shift
      ;;
    --port)
      FRONTEND_PORT="${2:-5173}"
      shift
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 1
      ;;
  esac
  shift
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$SCRIPT_DIR"
DB_PATH="$REPO_ROOT/social_network.db"
FRONTEND_DIR="$REPO_ROOT/frontend"
LOG_DIR="$REPO_ROOT/logs"
AUTH_URL="http://localhost:8081"
mkdir -p "$LOG_DIR"

SERVICE_NAMES=()
SERVICE_PIDS=()

log_section() {
  echo -e "\n=== $1 ==="
}

log_info() {
  echo "  - $1"
}

detect_pkg_mgr() {
  if command -v apt-get >/dev/null 2>&1; then
    echo "apt-get"
  elif command -v dnf >/dev/null 2>&1; then
    echo "dnf"
  elif command -v pacman >/dev/null 2>&1; then
    echo "pacman"
  elif command -v zypper >/dev/null 2>&1; then
    echo "zypper"
  elif command -v brew >/dev/null 2>&1; then
    echo "brew"
  else
    echo ""
  fi
}

run_privileged() {
  local mgr="$1"; shift
  if [[ "$mgr" == "brew" ]]; then
    "$@"
    return
  fi

  if [[ $IS_ROOT -eq 1 ]]; then
    "$@"
    return
  fi

  if command -v sudo >/dev/null 2>&1; then
    sudo "$@"
  else
    echo "Need administrative privileges to run '$*', but sudo is unavailable." >&2
    exit 1
  fi
}

install_packages() {
  local mgr="$1"; shift
  local pkgs=("$@")

  case "$mgr" in
    apt-get)
      run_privileged "$mgr" apt-get update
      run_privileged "$mgr" apt-get install -y "${pkgs[@]}"
      ;;
    dnf)
      run_privileged "$mgr" dnf install -y "${pkgs[@]}"
      ;;
    pacman)
      run_privileged "$mgr" pacman -Sy --noconfirm "${pkgs[@]}"
      ;;
    zypper)
      run_privileged "$mgr" zypper install -y "${pkgs[@]}"
      ;;
    brew)
      brew install "${pkgs[@]}"
      ;;
    *)
      return 1
      ;;
  esac
}

ensure_command() {
  local cmd="$1"
  local title="$2"
  shift 2
  local -a specs=("$@")

  if command -v "$cmd" >/dev/null 2>&1; then
    log_info "$title already installed."
    return
  fi

  if [[ $SKIP_INSTALL -eq 1 ]]; then
    echo "$title is required but missing. Install it or rerun without --skip-install." >&2
    exit 1
  fi

  local mgr
  mgr="$(detect_pkg_mgr)"
  if [[ -z "$mgr" ]]; then
    echo "No supported package manager found to install $title. Install it manually." >&2
    exit 1
  fi

  local fallback=""
  local -a packages=()
  for spec in "${specs[@]}"; do
    if [[ "$spec" != *:* ]]; then
      fallback="$spec"
      continue
    fi
    local key="${spec%%:*}"
    local value="${spec#*:}"
    if [[ "$key" == "default" ]]; then
      fallback="$value"
      continue
    fi
    if [[ "$key" == "$mgr" ]]; then
      read -ra packages <<<"$value"
      break
    fi
  done

  if [[ ${#packages[@]} -eq 0 ]]; then
    if [[ -n "$fallback" ]]; then
      read -ra packages <<<"$fallback"
    else
      echo "No package mapping provided for $title and manager '$mgr'." >&2
      exit 1
    fi
  fi

  log_info "$title missing. Installing with $mgr..."
  install_packages "$mgr" "${packages[@]}"

  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "$title installation appears to have failed. Please install manually." >&2
    exit 1
  fi
}

cleanup() {
  if [[ ${#SERVICE_PIDS[@]} -eq 0 ]]; then
    exit 0
  fi
  echo -e "\nStopping services..."
  for pid in "${SERVICE_PIDS[@]}"; do
    if kill -0 "$pid" >/dev/null 2>&1; then
      kill "$pid" >/dev/null 2>&1 || true
      wait "$pid" >/dev/null 2>&1 || true
    fi
  done
}
trap cleanup EXIT INT TERM

ensure_database() {
  if [[ ! -f "$DB_PATH" ]]; then
    log_info "Creating SQLite database at $DB_PATH"
    touch "$DB_PATH"
  else
    log_info "Using existing database at $DB_PATH"
  fi
}

ensure_go_modules() {
  log_section "Downloading Go modules"
  (cd "$REPO_ROOT" && go mod download)
}

ensure_root_node_modules() {
  if [[ ! -f "$REPO_ROOT/package.json" ]]; then
    return
  fi
  log_section "Installing shared Node.js dependencies"
  (cd "$REPO_ROOT" && npm install --no-fund --no-audit)
}

ensure_frontend() {
  if [[ $SKIP_FRONTEND -eq 1 ]]; then
    return
  fi
  log_section "Installing frontend dependencies"
  (cd "$FRONTEND_DIR" && npm install --no-fund --no-audit)
}

start_service() {
  local name="$1"
  local rel_path="$2"
  local env_defs="$3"

  local service_path="$REPO_ROOT/$rel_path"
  if [[ ! -d "$service_path" ]]; then
    log_info "Skipping $name (missing directory $service_path)."
    return
  fi

  local log_file="$LOG_DIR/${name// /_}.log"
  log_info "Starting $name (logs -> $log_file)"

  (
    cd "$service_path"
    export DATABASE_PATH="$DB_PATH"
    for pair in $env_defs; do
      if [[ -n "$pair" ]]; then
        export "$pair"
      fi
    done
    nohup go run main.go >"$log_file" 2>&1
  ) &

  SERVICE_PIDS+=("$!")
  SERVICE_NAMES+=("$name")
  sleep 0.3
}

start_frontend() {
  if [[ $SKIP_FRONTEND -eq 1 ]]; then
    log_info "Skipping frontend start."
    return
  fi
  local log_file="$LOG_DIR/frontend.log"
  log_info "Starting frontend (http://$FRONTEND_HOST:$FRONTEND_PORT) logs -> $log_file"
  (
    cd "$FRONTEND_DIR"
    nohup npm run dev -- --host "$FRONTEND_HOST" --port "$FRONTEND_PORT" >"$log_file" 2>&1
  ) &
  SERVICE_PIDS+=("$!")
  SERVICE_NAMES+=("Frontend")
}

open_browser() {
  if [[ $SKIP_FRONTEND -eq 1 ]]; then
    return
  fi
  local url="http://$FRONTEND_HOST:$FRONTEND_PORT"
  if command -v xdg-open >/dev/null 2>&1; then
    xdg-open "$url" >/dev/null 2>&1 &
  elif command -v sensible-browser >/dev/null 2>&1; then
    sensible-browser "$url" >/dev/null 2>&1 &
  elif [[ "$OSTYPE" == "darwin"* ]] && command -v open >/dev/null 2>&1; then
    open "$url" >/dev/null 2>&1 &
  else
    log_info "Unable to auto-open browser. Visit $url manually."
    return
  fi
  log_info "Opening browser at $url"
}

log_section "Checking prerequisites"
ensure_command go "Go" \
  "apt:golang-go" \
  "dnf:golang" \
  "pacman:go" \
  "zypper:go" \
  "brew:go" \
  "default:golang"
ensure_command node "Node.js (LTS)" \
  "apt:nodejs npm" \
  "dnf:nodejs npm" \
  "pacman:nodejs npm" \
  "zypper:nodejs npm" \
  "brew:node" \
  "default:nodejs"
ensure_command npm "npm" \
  "apt:npm" \
  "dnf:npm" \
  "pacman:npm" \
  "zypper:npm" \
  "brew:node" \
  "default:npm"
ensure_command gcc "C build tools (for go-sqlite3)" \
  "apt:build-essential" \
  "dnf:gcc" \
  "pacman:base-devel" \
  "zypper:gcc" \
  "brew:gcc" \
  "default:gcc"

ensure_database
ensure_go_modules
ensure_root_node_modules
ensure_frontend

SERVICES=(
  "Auth Service|services/auth|"
  "Users Service|services/users|AUTH_SERVICE_URL=$AUTH_URL"
  "Posts Service|services/posts|AUTH_SERVICE_URL=$AUTH_URL PORT=8083"
  "Groups Service|services/groups|AUTH_SERVICE_URL=$AUTH_URL"
  "Chat Service|services/chat|AUTH_SERVICE_URL=$AUTH_URL"
  "Notifications Service|services/notifications|AUTH_SERVICE_URL=$AUTH_URL"
)

log_section "Starting backend services"
for svc in "${SERVICES[@]}"; do
  IFS="|" read -r name path envs <<<"$svc"
  start_service "$name" "$path" "$envs"
done

log_section "Starting frontend"
start_frontend

log_section "Summary"
log_info "Auth Service:          http://localhost:8081"
log_info "Users Service:         http://localhost:8082"
log_info "Posts Service:         http://localhost:8083"
log_info "Groups Service:        http://localhost:8084"
log_info "Chat Service:          http://localhost:8085"
log_info "Notifications Service: http://localhost:8086"
if [[ $SKIP_FRONTEND -eq 0 ]]; then
  log_info "Frontend:              http://$FRONTEND_HOST:$FRONTEND_PORT"
fi
log_info "Logs directory:        $LOG_DIR"
echo ""
echo "Use: tail -f logs/Auth_Service.log   # to watch logs"
echo "Press Ctrl+C to stop all services."

open_browser

while true; do
  sleep 1
done
