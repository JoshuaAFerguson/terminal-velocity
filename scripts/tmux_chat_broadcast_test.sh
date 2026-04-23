#!/bin/bash
# File: scripts/tmux_chat_broadcast_test.sh
# Project: Terminal Velocity
# Description: Phase 4 chat broadcast regression — two sessions share state.
# Version: 1.0.0
# Author: Joshua Ferguson
# Created: 2026-04-23
#
# Opens two SSH sessions for the tester account, navigates both to Chat,
# sends a Global message from session A, waits past one chatPollTick()
# cycle (~1s) for session B's auto-refresh, and asserts the message
# appears in session B's capture. Proves the server-wide chat.Manager
# actually fans messages across SSH connections — previously each
# session had its own chat.Manager and messages stayed stranded.
#
# Pre-reqs: server running; tester account with TestPass123!.
# Usage: ./scripts/tmux_chat_broadcast_test.sh
# Exit 0 on match, 1 on miss (investigate with captures under /tmp/tv_captures/).
set -u
OUT="/tmp/tv_captures"
mkdir -p "$OUT"
ssh-keygen -R "[localhost]:2222" >/dev/null 2>&1 || true
KH=/tmp/tv_kh; : > "$KH"

sA="tvchatA_$$"
sB="tvchatB_$$"
cleanup() {
  tmux kill-session -t "$sA" 2>/dev/null || true
  tmux kill-session -t "$sB" 2>/dev/null || true
}
trap cleanup EXIT

login() {
  local s=$1
  tmux new-session -d -s "$s" -x 120 -y 40
  tmux send-keys -t "$s" "ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=$KH tester@localhost" C-m
  sleep 3
  tmux send-keys -t "$s" -l "tester"; sleep 0.3
  tmux send-keys -t "$s" Tab; sleep 0.3
  tmux send-keys -t "$s" -l "TestPass123!"; sleep 0.3
  tmux send-keys -t "$s" Tab; sleep 0.3
  tmux send-keys -t "$s" Enter
  sleep 4
}

login "$sA"
login "$sB"

# Navigate both to Chat. Chat is at index 13 (Launch=0, Nav=1, Trading=2,
# Cargo=3, Shipyard=4, Outfitter=5, AdvOutfit=6, ShipMgmt=7, PilotRecord=8,
# Missions=9, Quests=10, Achievements=11, Leaderboards=12, Players=13,
# Chat=14).
goto_chat() {
  local s=$1
  for _ in $(seq 1 14); do
    tmux send-keys -t "$s" Down; sleep 0.05
  done
  tmux send-keys -t "$s" Enter
  sleep 1
}

goto_chat "$sA"
goto_chat "$sB"

# Session A enters message mode and sends
tmux send-keys -t "$sA" "i"; sleep 0.5
MSG="hello-from-session-A-$$"
tmux send-keys -t "$sA" -l "$MSG"; sleep 0.3
tmux send-keys -t "$sA" Enter
# Wait past one chatPollTick() cycle (1s cadence) so Session B's poll
# fires and re-renders to include the freshly-fanned message.
sleep 3
tmux capture-pane -p -t "$sA" > "$OUT/chat_sessionA.txt"
tmux capture-pane -p -t "$sB" > "$OUT/chat_sessionB.txt"

echo "Session A capture: $OUT/chat_sessionA.txt"
echo "Session B capture: $OUT/chat_sessionB.txt"

if grep -q "$MSG" "$OUT/chat_sessionB.txt"; then
  echo "PASS — session B sees session A's message: $MSG"
  exit 0
fi
echo "FAIL — session B did not receive session A's message"
echo "--- session B ---"
tail -25 "$OUT/chat_sessionB.txt"
exit 1
