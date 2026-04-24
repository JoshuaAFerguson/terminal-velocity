# Landing page + reverse proxy in front of the web terminal.
#
# Single-container static HTML + nginx that:
#   - serves /  → docker/landing/html/index.html (the marketing page)
#   - serves /play → reverse-proxies to the ttyd webterm container,
#     with WebSocket upgrade headers required for the xterm.js
#     terminal to actually work
#   - serves /stats → reverse-proxies to the game server's /health
#     endpoint, so the landing page can show live player counts
#     without CORS gymnastics
FROM nginx:alpine

# Drop the stock `server {}` that listens on :80 and serves /usr/share/nginx/html
# unchanged. Our config fully replaces it — we want tight control over the
# WebSocket proxy at /play and the limited /stats passthrough.
RUN rm -f /etc/nginx/conf.d/default.conf

COPY nginx.conf /etc/nginx/conf.d/landing.conf
COPY html/ /usr/share/nginx/html/
