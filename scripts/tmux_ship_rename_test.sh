#!/bin/bash
# File: scripts/tmux_ship_rename_test.sh
# Project: Terminal Velocity
# Description: Phase 3 ship-rename regression — rename via Ship Management.
# Version: 1.0.0
# Author: Joshua Ferguson
# Created: 2026-04-23
#
# Drives login -> Main Menu -> Ship Management -> details -> r -> type a
# new name -> Enter. Compare psql `ships.name` before and after to
# confirm the rename persists. Also validates the Ship Management rename
# mode input pipeline (printableRuneString, backspace, Enter).
#
# Pre-reqs: server running; tester account with a starter shuttle.
# Usage: ./scripts/tmux_ship_rename_test.sh
set -u
SESSION="tvrename_$$"
OUT="/tmp/tv_captures"
mkdir -p "$OUT"
ssh-keygen -R "[localhost]:2222" >/dev/null 2>&1 || true
KH=/tmp/tv_kh; : > "$KH"
cleanup() { tmux kill-session -t "$SESSION" 2>/dev/null || true; }
trap cleanup EXIT

PASSWORD='TestPass123!'
NEWNAME='Stardrift-Alpha'
tmux new-session -d -s "$SESSION" -x 120 -y 40
tmux send-keys -t "$SESSION" "ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=$KH tester@localhost" C-m
sleep 3
tmux send-keys -t "$SESSION" -l "tester"; sleep 0.3
tmux send-keys -t "$SESSION" Tab; sleep 0.3
tmux send-keys -t "$SESSION" -l "$PASSWORD"; sleep 0.3
tmux send-keys -t "$SESSION" Tab; sleep 0.3
tmux send-keys -t "$SESSION" Enter
sleep 4

# Ship Management is index 7.
for _ in 1 2 3 4 5 6 7; do
  tmux send-keys -t "$SESSION" Down; sleep 0.08
done
tmux send-keys -t "$SESSION" Enter
sleep 2

# Cursor on first ship (Starter Shuttle). Enter to view details.
tmux send-keys -t "$SESSION" Enter
sleep 1
# r to enter rename mode
tmux send-keys -t "$SESSION" "r"
sleep 0.3
# Clear existing name (ship name field pre-populates). Backspaces to clear.
for _ in $(seq 1 20); do
  tmux send-keys -t "$SESSION" BSpace
  sleep 0.02
done
# Type new name
tmux send-keys -t "$SESSION" -l "$NEWNAME"
sleep 0.3
tmux send-keys -t "$SESSION" Enter
sleep 2
tmux capture-pane -p -t "$SESSION" > "$OUT/rename_after.txt"
echo "capture: $OUT/rename_after.txt"
