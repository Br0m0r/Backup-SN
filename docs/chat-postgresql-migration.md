# Chat PostgreSQL Migration Runbook

## Scope

Chat owns a private PostgreSQL database containing `messages` and
`group_messages`, an embedded migration history, and an idempotent SQLite copy
command that verifies both tables before committing. User, sender, recipient,
and group IDs are external domain references; the Chat schema intentionally has
no cross-database foreign keys.

Chat still reads profiles and follow relationships from shared SQLite. Those
reads are temporary and do not include message tables.

## Fresh local environment

1. Copy `.env.example` to `.env` and replace the development secrets.
2. Run `docker compose up --build`.
3. Compose starts `chat-db`, runs `chat-migrations`, then starts Chat.
4. Check `http://localhost:8080/api/chat/health`.

The PostgreSQL port is private and its local data is stored in the
`chat-postgres-data` volume.

## Existing SQLite cutover

Stop Chat writes for the entire copy and cutover window:

```powershell
Copy-Item .\social_network.db .\social_network.pre-chat-postgres.db
docker compose stop chat-service
docker compose up -d chat-db
docker compose run --rm chat-migrations
docker compose run --rm --entrypoint ./copy-sqlite `
  -v "${PWD}/social_network.db:/migration/social_network.db:ro" `
  chat-migrations -sqlite-path /migration/social_network.db
docker compose up -d chat-service gateway
```

The copy preserves IDs and timestamps, upserts both tables in one transaction,
compares complete row counts and SHA-256 content checksums, advances identity
sequences, and commits only after verification. It is safe to rerun while Chat
writes remain stopped and intentionally rejects divergent target data.

For production, provision a least-privilege Chat role, require TLS in
`DATABASE_URL`, run `./migrate` as a release job, copy from a read-only SQLite
snapshot during a planned pause, and record the verified counts/checksum before
restoring traffic.

## Rollback

Before PostgreSQL-only writes occur, stop Chat and return to the previous release
and untouched SQLite snapshot. After PostgreSQL accepts writes, do not point the
old service at SQLite without a reverse migration and reconciliation of both
histories.
