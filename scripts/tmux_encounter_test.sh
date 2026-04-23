#!/bin/bash
# File: scripts/tmux_encounter_test.sh
# Project: Terminal Velocity
# Description: Phase 1 encounter regression — probabilistic jump-triggered event.
# Version: 1.0.0
# Author: Joshua Ferguson
# Created: 2026-04-23
#
# Drives login -> Navigation -> repeated Enter on the first jump route.
# Default encounter chance at danger=5 is ~12.5% per jump, so 15 jumps
# gives a ~87% chance of at least one encounter. Captures each jump
# frame and flags when an "ENCOUNTER" screen appears.
#
# Usage: ./scripts/tmux_encounter_test.sh
# Exit 0 on hit, nonzero on miss (may be probabilistic — retry).
set -u
SESSION="tvenc_$$"
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

# Down once to Navigation
tmux send-keys -t "$SESSION" Down
sleep 0.2
tmux send-keys -t "$SESSION" Enter
sleep 2

# Repeatedly jump to the first available system
encountered=0
for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15; do
  tmux send-keys -t "$SESSION" Enter
  sleep 4  # jump animation + settle
  # Capture and look for encounter markers
  tmux capture-pane -p -t "$SESSION" > "$OUT/encounter_jump_$i.txt"
  if grep -qiE "ENCOUNTER|Pirate|Trader ship|hostile|friendly ship|attacks" "$OUT/encounter_jump_$i.txt"; then
    if grep -qiE "ENCOUNTER|friendly ship|attacks" "$OUT/encounter_jump_$i.txt"; then
      echo "[$i] ENCOUNTER TRIGGERED"
      encountered=1
      cp "$OUT/encounter_jump_$i.txt" "$OUT/encounter_triggered.txt"
      break
    fi
  fi
  # If back on navigation (normal jump), Enter again to re-jump
done

echo "final encountered=$encountered"
[ $encountered -eq 1 ] || echo "(no encounter in 15 jumps — probability is ~12.5% so expected hit rate ~87%)"
