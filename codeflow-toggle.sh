#!/usr/bin/env bash
#
# codeflow-toggle.sh — toggle the codeflow container stack on/off.
#
# Running this script flips the state:
#   - if codeflow containers are running  -> stops them
#   - if codeflow containers are stopped  -> starts them
#
# Usage:
#   ./codeflow-toggle.sh           # toggle
#   ./codeflow-toggle.sh up        # force start
#   ./codeflow-toggle.sh down      # force stop
#   ./codeflow-toggle.sh status    # just print current state
#
set -euo pipefail

# Containers managed by this script.
CONTAINERS=(codeflow-postgres-1 codeflow-redis-1)

# Returns 0 if the named container exists and is running, 1 otherwise.
is_running() {
  local name="$1"
  [ "$(docker inspect -f '{{.State.Running}}' "$name" 2>/dev/null)" = "true" ]
}

# Returns 0 if the named container exists at all.
exists() {
  docker inspect "$1" >/dev/null 2>&1
}

print_status() {
  echo "Codeflow container status:"
  for c in "${CONTAINERS[@]}"; do
    if ! exists "$c"; then
      echo "  - $c: NOT FOUND"
    elif is_running "$c"; then
      echo "  - $c: running"
    else
      echo "  - $c: stopped"
    fi
  done
}

start_all() {
  echo "Starting codeflow containers..."
  for c in "${CONTAINERS[@]}"; do
    if ! exists "$c"; then
      echo "  ! $c not found — skipping"
      continue
    fi
    docker start "$c" >/dev/null && echo "  + started $c"
  done
}

stop_all() {
  echo "Stopping codeflow containers..."
  for c in "${CONTAINERS[@]}"; do
    if ! exists "$c"; then
      echo "  ! $c not found — skipping"
      continue
    fi
    docker stop "$c" >/dev/null && echo "  - stopped $c"
  done
}

# Decide whether the stack is currently "up" — true if ANY managed
# container is running.
stack_is_up() {
  for c in "${CONTAINERS[@]}"; do
    if is_running "$c"; then
      return 0
    fi
  done
  return 1
}

ACTION="${1:-toggle}"

case "$ACTION" in
  up|start)
    start_all
    ;;
  down|stop)
    stop_all
    ;;
  status)
    print_status
    exit 0
    ;;
  toggle)
    if stack_is_up; then
      stop_all
    else
      start_all
    fi
    ;;
  *)
    echo "Unknown action: $ACTION" >&2
    echo "Usage: $0 [toggle|up|down|status]" >&2
    exit 1
    ;;
esac

echo
print_status
