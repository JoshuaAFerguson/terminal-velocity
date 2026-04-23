#!/bin/bash
# File: scripts/tmux_duel_challenge_test.sh
# Project: Terminal Velocity
# Description: Phase 4 PvP-duel challenge regression.
# Version: 1.0.0
# Author: Joshua Ferguson
# Created: 2026-04-23
#
# Opens two SSH sessions (alice + tester), parks alice on the PvP
# Combat screen, navigates tester to the Players screen, presses 'c'
# on alice to issue a duel challenge, and verifies alice's side shows
# the incoming challenge after one pvpPollTick() cycle (~2s).
#
# Pre-reqs: tester + alice accounts with TestPass123!, both online.
# Usage: ./scripts/tmux_duel_challenge_test.sh
set -u
OUT="/tmp/tv_captures"
mkdir -p "$OUT"
ssh-keygen -R "[localhost]:2222" >/dev/null 2>&1 || true
KH=/tmp/tv_kh; : > "$KH"

sA="tvduelT_$$"
sB="tvduelA_$$"
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

# Session B (alice) logs in first so tester's Players screen has her.
login "$sB" "alice"
login "$sA" "tester"

# Alice opens PvP Combat screen (index 20). We navigate Down 20 times
# from the Launch cursor.
for _ in $(seq 1 20); do
  tmux send-keys -t "$sB" Down; sleep 0.05
done
tmux send-keys -t "$sB" Enter
sleep 2
tmux capture-pane -p -t "$sB" > "$OUT/duel_alice_before.txt"

# Tester opens Players screen (index 13) and presses c on alice.
# tester cursor starts at 0. alice shows up if she's online and in
# tester's filtered list.
for _ in $(seq 1 13); do
  tmux send-keys -t "$sA" Down; sleep 0.05
done
tmux send-keys -t "$sA" Enter
sleep 2
tmux capture-pane -p -t "$sA" > "$OUT/duel_tester_players.txt"

# On the Players screen, cursor starts at 0 = the first player in the
# filter. If that's tester themselves, press Down once to advance.
head_line=$(grep -m1 -E "^> |^\* |^>" "$OUT/duel_tester_players.txt" | head -1)
if echo "$head_line" | grep -q "tester"; then
  tmux send-keys -t "$sA" Down; sleep 0.2
fi

# Press c → challenge
tmux send-keys -t "$sA" "c"
sleep 3
tmux capture-pane -p -t "$sA" > "$OUT/duel_tester_pvp.txt"

# Wait past alice's poll tick (2s)
sleep 3
tmux capture-pane -p -t "$sB" > "$OUT/duel_alice_after.txt"

echo "Tester PvP capture: $OUT/duel_tester_pvp.txt"
echo "Alice PvP capture:  $OUT/duel_alice_after.txt"

if grep -qiE "challenges? .* duel|tester|challenger" "$OUT/duel_alice_after.txt"; then
  echo "PASS — alice's PvP screen shows the incoming challenge"
  grep -E "tester|challenger|challenge" "$OUT/duel_alice_after.txt" | head -3
  exit 0
fi
echo "FAIL — alice didn't see tester's challenge"
echo "--- alice PvP ---"
head -30 "$OUT/duel_alice_after.txt"
exit 1
