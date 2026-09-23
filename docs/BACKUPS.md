# Backups

## What gets backed up, and how often

The entire **production QuoteTrack Postgres database** (Render free tier) is
dumped with `pg_dump` as a gzipped plain-SQL file:

- **Schedule:** every day at **03:00 UTC**, via `.github/workflows/backup.yml`
  (GitHub Actions — Render cron jobs require a paid plan, so the schedule
  lives here). A manual run is available via the workflow's *Run workflow*
  button (`workflow_dispatch`).
- **Destination:** Cloudflare R2 bucket, under `quotetrack/quotetrack-YYYY-MM-DD.sql.gz`
  (R2 free tier: 10 GB storage, zero egress fees).
- **Contents:** all tables (`users`, `customers`, `quotes`, `quote_items`,
  `password_reset_tokens`, `schema_migrations`, …), `--no-owner
  --no-privileges` so it restores cleanly under any role.
- **Local testing:** `PROD_DATABASE_URL=postgres://... ./scripts/backup.sh`
  writes `backups/quotetrack-YYYY-MM-DD.sql.gz` locally (the `backups/`
  directory is gitignored — dumps contain production data, never commit them).

## Manual setup (human) — one time

No account was created on your behalf; do these steps yourself:

1. **Create the bucket**
   1. Sign up / log in at <https://dash.cloudflare.com> → **R2 Object Storage**.
   2. **Create bucket** → name it `quotetrack-backups` (or similar), region
      auto, plan **Free**.
2. **Create API credentials**
   1. R2 → **Manage R2 API Tokens** → **Create API token**.
   2. Permissions: **Object Read & Write**, scoped to that bucket only.
   3. Copy the **Access Key ID**, **Secret Access Key**, and note your
      **account endpoint** (`https://<ACCOUNT_ID>.r2.cloudflarestorage.com`,
      shown in the R2 dashboard sidebar under "Account ID).
3. **Add GitHub secrets** — repo → **Settings → Secrets and variables →
   Actions → New repository secret**, with these names (values come from the
   steps above):
   - `PROD_DATABASE_URL` — from the Render dashboard: Postgres instance →
     **Connections → External Database URL** (includes `sslmode=require`).
   - `R2_ENDPOINT` — the account endpoint URL from step 2.3.
   - `R2_BUCKET` — the bucket name from step 1.2.
   - `R2_ACCESS_KEY_ID` — from step 2.3.
   - `R2_SECRET_ACCESS_KEY` — from step 2.3.
4. **Verify:** run the *Daily Postgres backup* workflow manually
   (**Actions → Daily Postgres backup → Run workflow**) and confirm an object
   appears in the R2 bucket. The daily 03:00 UTC run takes over from there.
5. **Optional retention:** in R2 → bucket → **Settings → Object
   Lifecycles**, add a rule to delete objects older than e.g. 30 days so the
   10 GB free tier never fills up.

## How to restore from a dump

Plain SQL dump → restore with `psql` (no `pg_restore` needed).

**Into a brand-new Render Postgres instance:**

```bash
gunzip -c quotetrack-YYYY-MM-DD.sql.gz | psql "$DATABASE_URL"
```

**Into a local database (e.g. after `docker compose up -d`):**

```bash
gunzip -c backups/quotetrack-YYYY-MM-DD.sql.gz \
  | psql "postgres://quotetrack:quotetrack@localhost:5432/quotetrack?sslmode=disable"
```

Notes:

- Run the app once (or any path that runs `database.Migrate`) afterwards —
  `schema_migrations` is part of the dump, so already-applied migrations are
  respected automatically.
- The dump uses `--no-owner --no-privileges`, so objects are owned by the
  restoring role; no `ALTER OWNER` fixes needed.
- Always test a restore into a scratch database before you need it for real.

## Troubleshooting

- **`pg_dump: server version mismatch`** — Render runs a Postgres major
  version newer than the `postgresql-client` in the workflow image. Install
  the matching client version in `backup.yml` (e.g. from the official
  Postgres apt repo).
- **SSL error connecting** — the URL must include `?sslmode=require` (the
  Render External Database URL does).
- **Dump too large for the 10 GB free tier** — enable the lifecycle rule
  above; at QuoteTrack's current size this is far off.
