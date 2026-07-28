# Auth PostgreSQL Migration Runbook

## Scope

Auth owns a private PostgreSQL database containing credential-bearing
`accounts` and login `sessions`. Profile fields and follows remain owned by
Users. Auth allocates the stable account ID during registration and provisions
the matching Users profile through an authenticated, idempotent contract.

Users temporarily stores an empty password compatibility value while it
remains on the legacy combined SQLite table. Users never reads or exposes that
value, and its PostgreSQL extraction removes the column.

## Fresh local environment

1. Copy `.env.example` to `.env` and replace development secrets.
2. Run `docker compose up --build`.
3. Compose starts `auth-db`, runs `auth-migrations`, then starts Auth.
4. Check `http://localhost:8080/api/auth/health`.

PostgreSQL is private to the Compose network and uses the
`auth-postgres-data` volume.

## Existing SQLite cutover

Stop registration, login, logout, and session writes for the copy:

```powershell
Copy-Item .\social_network.db .\social_network.pre-auth-postgres.db
docker compose stop auth-service user-service
docker compose up -d auth-db
docker compose run --rm auth-migrations
docker compose run --rm --entrypoint ./copy-sqlite `
  -v "${PWD}/social_network.db:/migration/social_network.db:ro" `
  auth-migrations -sqlite-path /migration/social_network.db
docker compose up -d auth-service user-service gateway
```

The copy preserves IDs, credential hashes, tokens, and timestamps; verifies
row counts and a SHA-256 checksum in one transaction; advances identity
sequences; and commits only after verification.

## Rollback

Before PostgreSQL-only Auth writes occur, return to the previous release and
untouched SQLite snapshot. After new registrations or sessions exist, stop
writes and reconcile both stores before any rollback.
