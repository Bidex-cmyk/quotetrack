#!/usr/bin/env bash
#
# Dump the QuoteTrack production database to a local gzipped SQL file.
#
# Usage:
#   PROD_DATABASE_URL="postgres://..." ./scripts/backup.sh
#
# Writes: backups/quotetrack-YYYY-MM-DD.sql.gz  (UTC date; re-runs the same
# day overwrite that day's file, so it is idempotent).
#
# Notes:
#   - Requires pg_dump (postgresql-client) on PATH.
#   - Render's External Database URL already carries sslmode=require. If your
#     URL does not, export PGSSLMODE=require before running.
#   - The dump contains all production data. Keep it out of git (the
#     backups/ directory is gitignored) and encrypt it at rest in storage.

set -euo pipefail

: "${PROD_DATABASE_URL:?PROD_DATABASE_URL must be set}"

if ! command -v pg_dump >/dev/null 2>&1; then
  echo "error: pg_dump not found. Install postgresql-client first." >&2
  exit 1
fi

mkdir -p backups
out="backups/quotetrack-$(date -u +%Y-%m-%d).sql.gz"

# --no-owner/--no-privileges: restore into any role/database cleanly.
# pipefail: a pg_dump failure fails the script even behind the pipe.
pg_dump "$PROD_DATABASE_URL" --no-owner --no-privileges --no-password | gzip > "$out"

if [ ! -s "$out" ]; then
  echo "error: dump is empty ($out)" >&2
  exit 1
fi

echo "Wrote $out ($(du -h "$out" | cut -f1))"
