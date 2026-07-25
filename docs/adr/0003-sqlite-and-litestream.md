# ADR 0003: SQLite + Litestream over a managed database

## Status

Accepted

## Context

OpenPadel needs durable storage for Sessions, Players, Rounds, and Users, deployed as a single
Fly.io machine. A managed Postgres/MySQL instance would give built-in
replication, backups, and horizontal scale; SQLite on local disk is a single file with none of that
built in.

## Decision

Use SQLite (`mattn/go-sqlite3`, WAL mode, single connection via `SetMaxOpenConns(1)`) on the Fly
volume, replicated continuously to Tigris (S3-compatible) via the `litestream replicate` sidecar.
On container start, the DB restores from the replica if no local file exists.

This keeps deployment to one Go binary + one SQLite file — no separate database service to
provision, network to, or pay for. Litestream closes the durability gap a bare SQLite file would
otherwise have (machine loss, volume corruption).

## Consequences

- No horizontal read/write scaling — one writer, `SetMaxOpenConns(1)`. Fine at current
  personal/friends-and-family usage scale; would need revisiting before any multi-region or
  high-concurrency requirement.
- Restore-on-boot from Tigris means a fresh machine start is not instant if the local volume is
  empty — acceptable for a single always-on Fly machine, would need reconsideration under
  autoscaling that cycles machines frequently.
- Backups are continuous (Litestream) rather than snapshot-based — point-in-time restore, not
  discrete backup files to manage.
- This choice trades managed-DB conveniences (built-in HA, dashboards, connection pooling) for
  single-binary deploy simplicity. Revisit only if a concrete scale or multi-writer need appears.

## Deployment shape

This ADR's single-file DB is what makes the whole product deploy as **one Go binary**. A two-stage
Docker build has Bun compile the SvelteKit frontend, which Go then embeds (`//go:embed all:build`
in `internal/ui/`) and compiles into a single static binary on Alpine (~20 MB). `fly deploy` ships
it to Fly.io (Stockholm / `arn`); the SQLite file lives on a persistent Fly volume at
`/data/openpadel.db`. No separate web server or database service to run.
