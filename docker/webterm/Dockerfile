# Web-terminal bridge for Terminal Velocity.
#
# Purpose: let a player open https://tv.toconnect.me in a browser and
# land directly on the game's login screen — no SSH client install,
# no cloudflared, no keys, no known_hosts negotiation. ttyd
# (https://github.com/tsl0922/ttyd) serves an xterm.js-backed terminal
# over HTTP/WebSocket; this image wraps it in a fire-up-ssh spawn so
# every browser tab opens a fresh SSH session against the game.
#
# ttyd alpine image is BusyBox-based and doesn't ship openssh-client,
# so we add it here.
FROM tsl0922/ttyd:alpine

# openssh-client brings the `ssh` binary; ca-certificates keeps the
# TLS side of any transitive HTTPS calls inside the ssh process sane.
RUN apk add --no-cache openssh-client ca-certificates

# Bundled entrypoint is /usr/bin/ttyd. We don't override ENTRYPOINT
# here — the command comes from docker-compose, which keeps deploy
# tuning (max-clients, endpoint name) in one place alongside the
# rest of the stack config rather than baked into the image.
