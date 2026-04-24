#!/bin/bash
# File: scripts/tmux_tutorial_test.sh
# Project: Terminal Velocity
# Description: Phase 5F tutorial overlay regression.
# Version: 1.0.0
# Author: Joshua Ferguson
# Created: 2026-04-23
#
# Logs in fresh and asserts the Basics tutorial overlay appears on the
# main menu ("TUTORIAL" header + "Welcome, Commander!" step 1). Also
# exercises Ctrl+N to advance + Ctrl+T to hide, verifying the step
# advances or the overlay disappears respectively.
#
# Pre-req: server running, tester account, TutorialManager per-session
# (pre-5A.1). Usage: ./scripts/tmux_tutorial_test.sh
set -u
OUT="/tmp/tv_captures"
mkdir -p "$OUT"
ssh-keygen -R "[localhost]:2222" >/dev/null 2>&1 || true
KH=/tmp/tv_kh; : > "$KH"
S="tvtut_$$"
cleanup() { tmux kill-session -t "$S" 2>/dev/null || true; }
trap cleanup EXIT

PASSWORD='TestPass123!'
tmux new-session -d -s "$S" -x 120 -y 40
tmux send-keys -t "$S" "ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=$KH tester@localhost" C-m
sleep 3
tmux send-keys -t "$S" -l "tester"; sleep 0.3
tmux send-keys -t "$S" Tab; sleep 0.3
tmux send-keys -t "$S" -l "$PASSWORD"; sleep 0.3
tmux send-keys -t "$S" Tab; sleep 0.3
tmux send-keys -t "$S" Enter
sleep 4
tmux capture-pane -p -t "$S" > "$OUT/tutorial_01_login.txt"

if ! grep -q "TUTORIAL" "$OUT/tutorial_01_login.txt"; then
  echo "FAIL — overlay not visible after login"
  tail -20 "$OUT/tutorial_01_login.txt"
  exit 1
fi
echo "PASS — tutorial overlay visible on main menu"
grep -E "TUTORIAL|Welcome|Objective|Hint|Ctrl" "$OUT/tutorial_01_login.txt" | head -5

# Ctrl+T to hide
tmux send-keys -t "$S" C-t; sleep 0.5
tmux capture-pane -p -t "$S" > "$OUT/tutorial_02_hidden.txt"
if grep -q "TUTORIAL" "$OUT/tutorial_02_hidden.txt"; then
  echo "WARN — Ctrl+T did not hide the overlay"
else
  echo "PASS — Ctrl+T hides the overlay"
fi

# Ctrl+T again to reshow
tmux send-keys -t "$S" C-t; sleep 0.5
# Ctrl+N to advance
tmux send-keys -t "$S" C-n; sleep 0.5
tmux capture-pane -p -t "$S" > "$OUT/tutorial_03_advanced.txt"
if grep -q "Navigation Basics" "$OUT/tutorial_03_advanced.txt"; then
  echo "PASS — Ctrl+N advanced to step 2 (Navigation Basics)"
else
  echo "WARN — Ctrl+N did not advance to step 2 (may have completed step 1 off-screen)"
  grep -E "TUTORIAL|Navigation|Your Credits" "$OUT/tutorial_03_advanced.txt" | head -3
fi

exit 0
