#!/bin/bash
# File: scripts/tmux_menu_walk.sh
# Project: Terminal Velocity
# Description: Walk each main-menu entry via tmux, capture rendered state
# Version: 1.0.0
# Author: Joshua Ferguson
# Created: 2026-04-22
#
# Usage: ./scripts/tmux_menu_walk.sh [tag]
# Produces /tmp/tv_captures/walk_<tag>_<NN>_<name>.txt for every menu entry.

set -u

TAG="${1:-walk}"
SESSION="tvwalk_$$"
OUTDIR="${TV_CAPTURE_DIR:-/tmp/tv_captures}"
USER_NAME="${TV_TEST_USER:-tester}"
PASSWORD="${TV_TEST_PASSWORD:-TestPass123!}"
PORT="${TV_SSH_PORT:-2222}"
HOST="${TV_SSH_HOST:-localhost}"
KNOWN_HOSTS="${TV_KNOWN_HOSTS:-/tmp/tv_known_hosts}"

mkdir -p "$OUTDIR"
ssh-keygen -R "[$HOST]:$PORT" >/dev/null 2>&1 || true
: > "$KNOWN_HOSTS"

cleanup() { tmux kill-session -t "$SESSION" 2>/dev/null || true; }
trap cleanup EXIT

step=0
cap() {
  local name="$1"
  local path
  path="$OUTDIR/walk_${TAG}_$(printf '%02d' "$step")_${name}.txt"
  tmux capture-pane -p -t "$SESSION" > "$path"
  echo "[$step] $name → $path"
  step=$((step + 1))
}

tmux new-session -d -s "$SESSION" -x 120 -y 40

# Connect + log in
tmux send-keys -t "$SESSION" \
  "ssh -o StrictHostKeyChecking=accept-new -o UserKnownHostsFile=$KNOWN_HOSTS -p $PORT ${USER_NAME}@$HOST" Enter
sleep 2
tmux send-keys -t "$SESSION" -l "$USER_NAME"
sleep 0.3
tmux send-keys -t "$SESSION" Tab
sleep 0.2
tmux send-keys -t "$SESSION" -l "$PASSWORD"
sleep 0.3
tmux send-keys -t "$SESSION" Tab
sleep 0.2
tmux send-keys -t "$SESSION" Enter
sleep 3
cap "main_menu"

# Main menu has 23 items based on first walk. Drive down and capture each
# entry's screen. Press Enter to enter, Esc to go back.
#
# Order observed on first walk:
# 00 Launch, 01 Navigation, 02 Trading, 03 Cargo Hold, 04 Shipyard,
# 05 Outfitter, 06 Advanced Outfitting, 07 Ship Management, 08 Missions,
# 09 Quests, 10 Achievements, 11 Leaderboards, 12 Players, 13 Chat,
# 14 Factions, 15 Trade, 16 PvP Combat, 17 News, 18 Help, 19 Settings,
# 20 Tutorials, 21 Admin Panel, 22 Quit

menu_items=(launch navigation trading cargo shipyard outfitter outfitter_enhanced \
            ship_mgmt missions quests achievements leaderboards players chat \
            factions trade pvp news help settings tutorials admin)

# From "Launch" at position 0, walk down and enter each.
for i in "${!menu_items[@]}"; do
  name="${menu_items[$i]}"
  # Enter the item
  tmux send-keys -t "$SESSION" Enter
  sleep 1.2
  cap "enter_${i}_${name}"
  # Escape back
  tmux send-keys -t "$SESSION" Escape
  sleep 0.4
  cap "back_${i}_${name}"
  # Move to next
  tmux send-keys -t "$SESSION" Down
  sleep 0.15
done

echo ""
echo "Done — $(ls "$OUTDIR"/walk_${TAG}_*.txt | wc -l) frames captured."
