#!/bin/bash
# File: scripts/tmux_presence_test.sh
# Project: Terminal Velocity
# Description: Phase 4 presence-in-viewport regression.
# Version: 1.0.0
# Author: Joshua Ferguson
# Created: 2026-04-23
#
# Logs in as tester (session A) and alice (session B), both into
# Launch / Space View. After the 2s poll cycle, tester's viewport
# should show alice's ship glyph + name, or vice versa.
#
# Pre-reqs:
#   * server running, tester + alice accounts with passwords TestPass123!
#   * alice assigned to the same star system as tester (SQL setup)
# Usage: ./scripts/tmux_presence_test.sh
set -u
OUT="/tmp/tv_captures"
mkdir -p "$OUT"
ssh-keygen -R "[localhost]:2222" >/dev/null 2>&1 || true
KH=/tmp/tv_kh; : > "$KH"

sA="tvpresA_$$"
sB="tvpresB_$$"
cleanup() {
  tmux kill-session -t "$sA" 2>/dev/null || true
  tmux kill-session -t "$sB" 2>/dev/null || true
}
trap cleanup EXIT

PASSWORD='TestPass123!'
login() {
  local s=$1 user=$2
  tmux new-session -d -s "$s" -x 120 -y 40
  tmux send-keys -t "$s" "ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=$KH $user@localhost" C-m
  sleep 3
  tmux send-keys -t "$s" -l "$user"; sleep 0.3
  tmux send-keys -t "$s" Tab; sleep 0.3
  tmux send-keys -t "$s" -l "$PASSWORD"; sleep 0.3
  tmux send-keys -t "$s" Tab; sleep 0.3
  tmux send-keys -t "$s" Enter
  sleep 4
}

login "$sA" "tester"
login "$sB" "alice"

# Launch (index 0) both — opens space view.
tmux send-keys -t "$sA" Enter
sleep 2
tmux send-keys -t "$sB" Enter
sleep 2

# Wait past one poll cycle (2s) for presence data to flow in.
sleep 4

tmux capture-pane -p -t "$sA" > "$OUT/presence_tester.txt"
tmux capture-pane -p -t "$sB" > "$OUT/presence_alice.txt"

echo "capture A: $OUT/presence_tester.txt"
echo "capture B: $OUT/presence_alice.txt"

# Session A (tester) should see alice in the viewport.
if grep -qE "Alice-1|alice" "$OUT/presence_tester.txt"; then
  echo "PASS — tester's viewport shows alice"
elif grep -qE "Stardrift-Alpha|tester" "$OUT/presence_alice.txt"; then
  echo "PASS — alice's viewport shows tester"
else
  echo "FAIL — neither viewport references the other player"
  echo "--- tester's viewport ---"
  head -30 "$OUT/presence_tester.txt"
  exit 1
fi
