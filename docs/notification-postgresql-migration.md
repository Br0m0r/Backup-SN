# Notification PostgreSQL Migration Runbook

## Scope

Notifications is the first service extracted from the shared SQLite database. It now owns:

- A private PostgreSQL database and credentials.
- Its `notifications` table and indexes.
- An independent `schema_migrations` history.
- A release-job migration executable.
- An idempotent SQLite copy command that verifies row count and SHA-256 content before committing.

The PostgreSQL schema deliberately has no foreign key to `users`. `user_id` is a stable external identity reference owned by Auth/Users, which prevents a cross-database ownership dependency.

The repository's previous development SQLite data was deliberately discarded on July 14, 2026. The existing-data procedure below is retained for other environments only; this checkout should initialize fresh databases.

## Fresh local environment

1. Copy `.env.example` to `.env` and replace the internal service token.
2. Run `docker compose up --build`.
3. Compose waits for `notifications-db`, runs `notification-migrations` once, and then starts `notification-service`.
4. Check `http://localhost:8080/api/notifications/health`. The endpoint now verifies PostgreSQL connectivity and returns 503 if the database is unavailable.

PostgreSQL is private to the Compose network; port 5432 is not published to the host. Local data is stored in the `notifications-postgres-data` named volume.

## Existing SQLite data cutover

Take a copy of `social_network.db` before beginning. Notification-producing writes must remain stopped for the entire copy and cutover window.

```powershell
Copy-Item .\social_network.db .\social_network.pre-notifications-postgres.db
docker compose stop user-service post-service group-service chat-service notification-service
docker compose up -d notifications-db
docker compose run --rm notification-migrations
docker compose run --rm --entrypoint ./copy-sqlite `
  -v "${PWD}/social_network.db:/migration/social_network.db:ro" `
  notification-migrations -sqlite-path /migration/social_network.db
docker compose up -d
```

The copy command:

1. Applies pending PostgreSQL migrations.
2. Reads SQLite rows ordered by ID.
3. Upserts them into one PostgreSQL transaction while preserving IDs and timestamps.
4. Compares the complete source and target row counts and content checksums.
5. Advances the PostgreSQL identity sequence and commits only after verification succeeds.

It is safe to rerun while writes remain stopped. It intentionally fails if PostgreSQL contains extra or different rows, because silently merging divergent notification histories would make cutover correctness ambiguous.

## Deployment procedure

For staging or production:

1. Provision a PostgreSQL database and least-privilege service role.
2. Store `DATABASE_URL` in the platform secret manager. Require TLS (`sslmode=require` or stricter) rather than the local Compose `disable` setting.
3. Configure pool variables from `.env.example` for the database connection limit allocated to this service.
4. Run the image's `./migrate` executable as a release job before deploying application replicas.
5. During the planned write pause, run `./copy-sqlite -sqlite-path <read-only snapshot>` with the target `DATABASE_URL`.
6. Record the copy row count/checksum, start one Notification replica, and exercise create/list/read/delete and WebSocket delivery.
7. Scale out only after health checks and error metrics remain clean.

Application replicas never apply schema changes at startup. Migration SQL is embedded into the release executable, checksummed after application, and serialized with a PostgreSQL advisory lock.

## Rollback

Before new PostgreSQL-only notification writes occur, rollback is simply:

1. Stop Notification and notification-producing services.
2. Restore the pre-cutover Compose configuration or application release.
3. Start services against the untouched SQLite snapshot.

After PostgreSQL accepts new writes, do not point the old service at SQLite without a reverse data migration. Preserve both the SQLite snapshot and PostgreSQL backup, stop writes, reconcile the histories, and verify counts/checksums before changing traffic.

## Remaining production work

This slice establishes ownership and a repeatable cutover, but Phase 2 is not complete until:

- PostgreSQL backups, retention, point-in-time recovery, and restore drills are configured on the chosen platform.
- The remaining services are extracted after their cross-domain SQL reads are removed.
- Object-store retention, orphan cleanup, and backup/restore procedures are exercised.
- Redis production high availability, TLS credentials, and outage exercises are configured on the chosen platform.
