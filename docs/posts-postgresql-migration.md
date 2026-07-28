# Posts PostgreSQL Migration Runbook

## Scope

Posts owns a private PostgreSQL database containing `posts`, `post_viewers`, and
`comments`, plus an embedded migration history and repeatable SQLite copy
command. User, viewer, and group IDs are external references without
cross-database foreign keys.

Posts no longer reads Users-owned tables. It uses authenticated internal
contracts for profile summaries, profile search, and accepted-following IDs.
Users likewise reads profile posts through an authenticated Posts contract.

## Fresh local environment

1. Copy `.env.example` to `.env` and replace development secrets.
2. Run `docker compose up --build`.
3. Compose starts `posts-db`, runs `posts-migrations`, then starts Posts.
4. Check `http://localhost:8080/api/posts/health`.

PostgreSQL is private to the Compose network and uses the
`posts-postgres-data` volume.

## Existing SQLite cutover

Stop Posts writes for the entire copy:

```powershell
Copy-Item .\social_network.db .\social_network.pre-posts-postgres.db
docker compose stop post-service
docker compose up -d posts-db
docker compose run --rm posts-migrations
docker compose run --rm --entrypoint ./copy-sqlite `
  -v "${PWD}/social_network.db:/migration/social_network.db:ro" `
  posts-migrations -sqlite-path /migration/social_network.db
docker compose up -d post-service user-service gateway
```

The command preserves IDs and timestamps, upserts all three tables in one
transaction, verifies complete row counts and SHA-256 content checksums,
advances identity sequences, and commits only after verification. It is safe to
rerun while Posts writes remain stopped and rejects divergent target state.

Run the existing post-media migration before this database cutover if legacy
filesystem image paths remain.

For production, provision a least-privilege Posts role, require TLS in
`DATABASE_URL`, execute `./migrate` as a release job, and record copy
counts/checksum before restoring traffic.

## Rollback

Before PostgreSQL-only writes occur, stop Posts and return to the prior release
and untouched SQLite snapshot. After new PostgreSQL writes, reconcile and
reverse-migrate the histories before any rollback to SQLite.
