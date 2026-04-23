#!/bin/bash
# File: scripts/tmux_system_events_test.sh
# Project: Terminal Velocity
# Description: Phase 2 P2.5 regression — verify system_events tick surfaces in News.
# Version: 1.0.0
# Author: Joshua Ferguson
# Created: 2026-04-23
#
# Usage: ./scripts/tmux_system_events_test.sh [tag]
#
# Waits 2+ minutes for the server-side system_events tick handler to
# generate a system-tagged news article, then opens the News screen and
# verifies it references a real star system (matches one of the names
# from star_systems). Requires a running server and a generated
# universe.
#
# Pre-reqs:
#   * Server running on localhost:2222
#   * tester account + generated universe (./genmap -save)
#   * tmux, psql (via docker exec), ssh
#
# Exit 0 on a system-event match, 1 otherwise.
set -u

TAG="${1:-sysev}"
SESSION="tvsysev_$$"
OUT="${TV_CAPTURE_DIR:-/tmp/tv_captures}"
USER_NAME="${TV_TEST_USER:-tester}"
PASSWORD='TestPass123!'
PORT="${TV_SSH_PORT:-2222}"
HOST="${TV_SSH_HOST:-localhost}"
KH=/tmp/tv_kh

mkdir -p "$OUT"
ssh-keygen -R "[$HOST]:$PORT" >/dev/null 2>&1 || true
: > "$KH"
cleanup() { tmux kill-session -t "$SESSION" 2>/dev/null || true; }
trap cleanup EXIT

# Grab the current set of system names for later verification.
systems=$(docker exec terminal-velocity-db psql -U terminal_velocity -d terminal_velocity -tA -c "SELECT name FROM star_systems;" 2>/dev/null || true)
if [ -z "$systems" ]; then
  echo "FAIL — no star_systems rows (run ./genmap -save first)"
  exit 1
fi

# The system_events handler fires every 2 minutes. Wait 140s so we
# definitely catch one at least once.
echo "Waiting 140s for system_events tick..."
sleep 140

# Log in and navigate to News.
tmux new-session -d -s "$SESSION" -x 120 -y 40
tmux send-keys -t "$SESSION" "ssh -p $PORT -o StrictHostKeyChecking=no -o UserKnownHostsFile=$KH $USER_NAME@$HOST" C-m
sleep 3
tmux send-keys -t "$SESSION" -l "$USER_NAME"; sleep 0.3
tmux send-keys -t "$SESSION" Tab; sleep 0.3
tmux send-keys -t "$SESSION" -l "$PASSWORD"; sleep 0.3
tmux send-keys -t "$SESSION" Tab; sleep 0.3
tmux send-keys -t "$SESSION" Enter
sleep 4

# News is at index 20 in the main-menu list.
for _ in $(seq 1 20); do
  tmux send-keys -t "$SESSION" Down; sleep 0.08
done
tmux send-keys -t "$SESSION" Enter
sleep 2
tmux capture-pane -p -t "$SESSION" > "$OUT/sysev_${TAG}.txt"

# Match any line in the capture against one of the system names.
hit=""
while IFS= read -r system; do
  [ -z "$system" ] && continue
  if grep -qE "in $system|Near $system|$system System" "$OUT/sysev_${TAG}.txt"; then
    hit="$system"
    break
  fi
done <<EOF
$systems
EOF

if [ -n "$hit" ]; then
  echo "PASS — system event article references real system: $hit"
  grep -E "in $hit|Near $hit|$hit System" "$OUT/sysev_${TAG}.txt" | head -3
  exit 0
fi

echo "FAIL — no system-event article found referencing any star_system"
echo "Last News screen:"
head -30 "$OUT/sysev_${TAG}.txt"
exit 1
