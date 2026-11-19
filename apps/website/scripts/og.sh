#!/usr/bin/env bash
# Takes each language's Open Graph image from the dev server's /og/<lang> card into public/og-<lang>.png. Uses Playwright through bunx, so it needs no dependency of the site; the first run downloads its browser.
set -euo pipefail
cd "$(dirname "$0")/.."

port=5176
stop() {
  # bunx starts the server as a child of its own, so the server is found by its port rather than by bunx's process.
  lsof -ti "tcp:$port" -sTCP:LISTEN | xargs kill 2>/dev/null || true
}
trap stop EXIT INT TERM

bunx playwright install chromium >/dev/null
# Vite's own CLI: TanStack Start serves pages in dev only from the Vite copy its plugin imports, which `vp dev` does not run.
bunx vite --port "$port" --strictPort >/dev/null 2>&1 &

ready=
for _ in $(seq 1 60); do
  curl -fsS "http://localhost:$port/og/zh" >/dev/null 2>&1 && ready=1 && break
  sleep 1
done
[ -n "$ready" ] || { echo "the dev server did not serve /og/zh" >&2; exit 1; }

for lang in zh en ja; do
  bunx playwright screenshot --viewport-size=1200,630 --wait-for-timeout=1500 "http://localhost:$port/og/$lang" "public/og-$lang.png"
done
