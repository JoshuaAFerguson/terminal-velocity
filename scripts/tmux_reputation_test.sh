#!/bin/bash
# File: scripts/tmux_reputation_test.sh
# Project: Terminal Velocity
# Description: Phase 2 regression — attack a faction encounter, verify rep drops.
# Version: 1.0.0
# Author: Joshua Ferguson
# Created: 2026-04-23
#
# Drives login -> Navigation -> jump until a faction-backed encounter fires
# (police, faction, merchant, distress). Resolves salvage/pirate/asteroid
# encounters with the default option so the jump loop continues.
#
# Pre-run:
#   UPDATE ships SET fuel=100 WHERE owner_id = tester.id;
#   DELETE FROM player_reputation WHERE player_id = tester.id;
# Post-run:
#   SELECT faction_id, reputation FROM player_reputation WHERE player_id = tester.id;
#   SELECT is_criminal FROM players WHERE username='tester';
# Expected: one row with reputation = -20 (police patrol) or -20 on the
# encounter's FactionID; is_criminal = true when the patrol was police.
set -u
SESSION="tvrep_$$"
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

# Down to Navigation, Enter
tmux send-keys -t "$SESSION" Down
sleep 0.2
tmux send-keys -t "$SESSION" Enter
sleep 2

# Jump repeatedly until we hit an encounter. Each loop = 1 Enter + wait +
# check for "ENCOUNTER" in the pane.
for i in $(seq 1 80); do
  tmux send-keys -t "$SESSION" Enter
  sleep 4
  tmux capture-pane -p -t "$SESSION" > "$OUT/rep_jump_$i.txt"
  if grep -q "=== ENCOUNTER ===" "$OUT/rep_jump_$i.txt"; then
    # Attack option only exists on encounters with a faction. Match on
    # the "Faction:" field that renders for police/merchant/faction/
    # distress encounters.
    if grep -q "^Faction: " "$OUT/rep_jump_$i.txt"; then
      echo "[$i] faction encounter — attacking"
      tmux send-keys -t "$SESSION" Down
      sleep 0.3
      tmux send-keys -t "$SESSION" Enter
      sleep 3
      tmux capture-pane -p -t "$SESSION" > "$OUT/rep_after_attack.txt"
      break
    fi
    # Salvage / pirate / other — resolve with the default option so the
    # loop can keep jumping.
    echo "[$i] non-faction encounter — resolving"
    tmux send-keys -t "$SESSION" Enter
    sleep 2
  fi
done

echo "captures in $OUT/rep_*.txt"
