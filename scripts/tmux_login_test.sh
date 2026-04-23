#!/bin/bash
# File: scripts/tmux_login_test.sh
# Project: Terminal Velocity
# Description: Drive an SSH login through tmux to verify the TUI end-to-end
# Version: 1.0.0
# Author: Joshua Ferguson
# Created: 2026-04-22
#
# Why tmux: it allocates a real PTY and translates send-keys through the same
# line discipline a human user's terminal would. Python's pty.fork() harness
# can't reliably simulate Enter, so regressions in the login flow went
# unnoticed until this script caught them.
#
# Pre-reqs:
#   * Server is running on localhost:2222 (make run)
#   * `tester` account exists with password TestPass123!
#     ./accounts create -username tester -email tester@local.test -password
#   * tmux installed
#
# Usage:
#   ./scripts/tmux_login_test.sh [tag]
#
# Captures each step of the login flow to /tmp/tv_captures/tmux_<tag>_<N>_*.txt
# so you can diff against a known-good baseline. Exit status is zero on a
# successful login (main menu reached), nonzero otherwise.

set -u

TAG="${1:-login}"
SESSION="tvtest_$$"
OUTDIR="${TV_CAPTURE_DIR:-/tmp/tv_captures}"
USER_NAME="${TV_TEST_USER:-tester}"
PASSWORD="${TV_TEST_PASSWORD:-TestPass123!}"
PORT="${TV_SSH_PORT:-2222}"
HOST="${TV_SSH_HOST:-localhost}"
KNOWN_HOSTS="${TV_KNOWN_HOSTS:-/tmp/tv_known_hosts}"

mkdir -p "$OUTDIR"
# Drop any stale host key so the ssh connection doesn't abort.
ssh-keygen -R "[$HOST]:$PORT" >/dev/null 2>&1 || true
: > "$KNOWN_HOSTS"

cleanup() {
  tmux kill-session -t "$SESSION" 2>/dev/null || true
}
trap cleanup EXIT

step=0
cap() {
  local name="$1"
  local path
  path="$OUTDIR/tmux_${TAG}_$(printf '%02d' "$step")_${name}.txt"
  tmux capture-pane -p -t "$SESSION" > "$path"
  echo "[$step] $name → $path"
  step=$((step + 1))
}

# 120x40 is the comfortable default for this TUI.
tmux new-session -d -s "$SESSION" -x 120 -y 40

tmux send-keys -t "$SESSION" \
  "ssh -o StrictHostKeyChecking=accept-new -o UserKnownHostsFile=$KNOWN_HOSTS -p $PORT ${USER_NAME}@$HOST" Enter
sleep 2
cap "post_connect"

tmux send-keys -t "$SESSION" -l "$USER_NAME"
sleep 0.5
cap "after_username"

tmux send-keys -t "$SESSION" Tab
sleep 0.3
cap "after_tab_to_password"

tmux send-keys -t "$SESSION" -l "$PASSWORD"
sleep 0.5
cap "after_password"

tmux send-keys -t "$SESSION" Tab
sleep 0.3
cap "after_tab_to_login"

tmux send-keys -t "$SESSION" Enter
sleep 1
cap "after_enter"

sleep 2
cap "settle_2s"

sleep 3
cap "settle_5s"

# A successful login should have left us on the main menu.
final="$OUTDIR/tmux_${TAG}_$(printf '%02d' "$((step - 1))")_settle_5s.txt"
# Match either the old plain-text layout or the new boxed one (which uses
# "= MAIN MENU =" as the header marker).
if grep -qE "Main Menu|MAIN MENU" "$final"; then
  echo ""
  echo "PASS — reached main menu"
  exit 0
fi
if grep -q "Error\|error" "$final"; then
  echo ""
  echo "FAIL — login flow produced an error:"
  grep -E "Error|error" "$final" | head -5
  exit 2
fi
echo ""
echo "FAIL — did not reach main menu. Last frame:"
cat "$final"
exit 1
