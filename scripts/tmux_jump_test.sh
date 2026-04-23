#!/bin/bash
# File: scripts/tmux_jump_test.sh
# Project: Terminal Velocity
# Description: Phase 1 jump regression — navigate and execute a hyperspace jump.
# Version: 1.0.0
# Author: Joshua Ferguson
# Created: 2026-04-23
#
# Drives login -> Main Menu -> Navigation -> Enter on the first available
# jump route -> wait for the jump animation -> return. Compare psql
# players.current_system before vs after to confirm persistence.
#
# Usage: ./scripts/tmux_jump_test.sh [tag]
set -u
SESSION="tvjump_$$"
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
tmux capture-pane -p -t "$SESSION" > "$OUT/jump_01_menu.txt"

# Navigate to Navigation (item #2 — Down once from Launch)
tmux send-keys -t "$SESSION" Down; sleep 0.2
tmux send-keys -t "$SESSION" Enter
sleep 2
tmux capture-pane -p -t "$SESSION" > "$OUT/jump_02_nav.txt"

# Press Enter on first route
tmux send-keys -t "$SESSION" Enter
sleep 1
tmux capture-pane -p -t "$SESSION" > "$OUT/jump_03_jumping.txt"

# Wait for jump to complete (progress bar)
sleep 4
tmux capture-pane -p -t "$SESSION" > "$OUT/jump_04_arrived.txt"

# Back to main menu
tmux send-keys -t "$SESSION" Escape
sleep 1
tmux capture-pane -p -t "$SESSION" > "$OUT/jump_05_menu_after.txt"

echo "captures in $OUT/jump_*.txt"
