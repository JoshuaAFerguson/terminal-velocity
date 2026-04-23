#!/bin/bash
# File: scripts/tmux_outfit_test.sh
# Project: Terminal Velocity
# Description: Phase 1 outfitter-install regression — buy a weapon.
# Version: 1.0.0
# Author: Joshua Ferguson
# Created: 2026-04-23
#
# Drives login -> Main Menu -> Outfitter -> Enter on the first weapon ->
# Enter to confirm. Compare psql credits + ship_weapons before vs after
# to confirm the install persisted (not just the credit debit).
#
# Usage: ./scripts/tmux_outfit_test.sh [tag]
set -u
SESSION="tvoutfit_$$"
OUT="/tmp/tv_captures"
mkdir -p "$OUT"
ssh-keygen -R "[localhost]:2222" >/dev/null 2>&1 || true
KH=/tmp/tv_kh; : > "$KH"
cleanup() { tmux kill-session -t "$SESSION" 2>/dev/null || true; }
trap cleanup EXIT

PASSWORD='TestPass123!'
tmux new-session -d -s "$SESSION" -x 120 -y 40
tmux send-keys -t "$SESSION" "ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=$KH tester@localhost" C-m
sleep 3
tmux send-keys -t "$SESSION" -l "tester"; sleep 0.3
tmux send-keys -t "$SESSION" Tab; sleep 0.3
tmux send-keys -t "$SESSION" -l "$PASSWORD"; sleep 0.3
tmux send-keys -t "$SESSION" Tab; sleep 0.3
tmux send-keys -t "$SESSION" Enter
sleep 4

# Down 5 to Outfitter (Launch, Nav, Trading, Cargo, Shipyard, Outfitter)
for _ in 1 2 3 4 5; do
  tmux send-keys -t "$SESSION" Down
  sleep 0.1
done
tmux send-keys -t "$SESSION" Enter
sleep 3
tmux capture-pane -p -t "$SESSION" > "$OUT/outfit_01_browse.txt"

# cursor is on first weapon (Pulse Laser). Press Enter to enter confirm mode.
tmux send-keys -t "$SESSION" Enter
sleep 1
tmux capture-pane -p -t "$SESSION" > "$OUT/outfit_02_confirm.txt"

# Enter again to execute install
tmux send-keys -t "$SESSION" Enter
sleep 3
tmux capture-pane -p -t "$SESSION" > "$OUT/outfit_03_after.txt"

echo "captures in $OUT/outfit_*.txt"
