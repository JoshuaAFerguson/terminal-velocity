#!/bin/bash
# File: scripts/tmux_buy_test.sh
# Project: Terminal Velocity
# Description: Phase 1 buy-flow regression — login, launch, land, buy 10 food.
# Version: 1.0.0
# Author: Joshua Ferguson
# Created: 2026-04-23
#
# Verifies that Commodity Exchange's B + quantity + Enter actually transacts:
# credits debit, ship_cargo inserts, and the market adjusts stock/demand.
#
# Usage: ./scripts/tmux_buy_test.sh [tag]
# Compare DB state with psql before/after to confirm the transaction.
set -u
SESSION="tvbuy_$$"
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

# Launch
tmux send-keys -t "$SESSION" Enter
sleep 2

# Land
tmux send-keys -t "$SESSION" "l"
sleep 3

# Commodity Exchange
tmux send-keys -t "$SESSION" "c"
sleep 3
tmux capture-pane -p -t "$SESSION" > "$OUT/buy_01_market.txt"

# Move cursor down 9 times to reach Food (cheapest standard commodity)
for _ in 1 2 3 4 5 6 7 8 9; do
  tmux send-keys -t "$SESSION" Down
  sleep 0.1
done
tmux capture-pane -p -t "$SESSION" > "$OUT/buy_02_cursor_food.txt"

# Press B (buy mode)
tmux send-keys -t "$SESSION" "b"
sleep 1
tmux capture-pane -p -t "$SESSION" > "$OUT/buy_03_buy_mode.txt"

# Bump quantity to 10 (so it's visible in cargo)
for _ in 1 2 3 4 5 6 7 8 9; do
  tmux send-keys -t "$SESSION" "+"
  sleep 0.05
done
tmux capture-pane -p -t "$SESSION" > "$OUT/buy_04_qty_10.txt"

# Confirm
tmux send-keys -t "$SESSION" Enter
sleep 2
tmux capture-pane -p -t "$SESSION" > "$OUT/buy_05_after.txt"

echo "captures in $OUT/buy_*.txt"
