# Groups PostgreSQL Migration Runbook

## Scope

Groups owns a private PostgreSQL database containing `groups`,
`group_members`, `events`, and `event_responses`, plus an embedded migration
history and repeatable SQLite copy command. User and creator IDs are external
identity references without cross-database foreign keys.

Groups no longer reads Users-owned tables. It resolves member and event-creator
display data through the authenticated Users profile contract. Users likewise
uses the authenticated Groups participant contract when searching for invite
candidates, so it no longer reads `group_members`.

## Fresh local environment

1. Copy `.env.example` to `.env` and replace development secrets.
2. Run `docker compose up --build`.
3. Compose starts `groups-db`, runs `groups-migrations`, then starts Groups.
4. Check `http://localhost:8080/api/groups/health`.

PostgreSQL is private to the Compose network and uses the
`groups-postgres-data` volume.

## Existing SQLite cutover

Stop Groups writes for the entire copy:

```powershell
Copy-Item .\social_network.db .\social_network.pre-groups-postgres.db
docker compose stop group-service
docker compose up -d groups-db
docker compose run --rm groups-migrations
docker compose run --rm --entrypoint ./copy-sqlite `
  -v "${PWD}/social_network.db:/migration/social_network.db:ro" `
  groups-migrations -sqlite-path /migration/social_network.db
docker compose up -d group-service user-service chat-service gateway
```

The command preserves IDs and timestamps, upserts all four tables in one
transaction, verifies complete row counts and a SHA-256 content checksum,
advances identity sequences, and commits only after verification. It is safe
to rerun while Groups writes remain stopped and rejects divergent target state.

For production, provision a least-privilege Groups role, require TLS in
`DATABASE_URL`, execute `./migrate` as a release job, and record copy
counts/checksum before restoring traffic.

## Rollback

Before PostgreSQL-only writes occur, stop Groups and return to the prior release
and untouched SQLite snapshot. After new PostgreSQL writes, reconcile and
reverse-migrate the histories before any rollback to SQLite.
