# Media Object Storage Migration

## Current status

Chat, Posts/Comments, user avatars, and group images are now moved from container-local files to S3-compatible object storage.

The repository's previous development database and uploaded files were deliberately discarded on July 14, 2026. A fresh checkout therefore does not need either legacy media cutover command below; those commands remain available only for another environment that still contains old data.

- New Chat uploads are stored under `chat/users/<user-id>/<opaque-random-name>.<validated-extension>`.
- The server detects actual file content and accepts JPEG, PNG, or GIF up to 5 MiB.
- User-provided filenames are never used as object keys.
- Delete requests may only remove keys in the authenticated user's prefix.
- Post and comment images use the equivalent `posts/users/<user-id>/...` ownership prefix.
- Avatars use `avatars/users/<user-id>/...`; replacement removes the superseded owned object and deletion clears the matching database reference.
- Group images use `groups/users/<creator-id>/...`; failed database writes remove the new object and successful replacement removes the old owned object.
- The Gateway serves GET/HEAD requests under `/media`; write methods are rejected.
- Storage API credentials and ports remain private to the Compose network.

The local development provider is MinIO. Production can use any compatible managed S3 provider and should place a CDN in front of public media where appropriate.

## Fresh local environment

Copy `.env.example` to `.env`, set non-default credentials, then run:

```powershell
docker compose up --build
```

Compose starts `media`, waits for its health endpoint, runs `media-init` to create the configured bucket and grant anonymous read-only access, and then starts Chat, Posts, and Gateway. Neither the S3 API nor the MinIO console is published to the host.

## Existing Chat image cutover

The existing SQLite rows refer to `/uploads/chat/<filename>`. Run the migration before deploying the Chat version that no longer serves its local upload directory.

```powershell
Copy-Item .\social_network.db .\social_network.pre-chat-media.db
docker compose stop chat-service
docker compose up -d media
docker compose run --rm media-init
docker compose run --rm --entrypoint ./copy-chat-media `
  -v "${PWD}/services/chat/uploads/chat:/migration/chat:ro" `
  chat-service `
  -sqlite-path /app/social_network.db `
  -uploads-dir /migration/chat
docker compose up -d chat-service gateway
```

The command selects only legacy Chat paths, verifies each local file's size and detected MIME type, uploads it with a new opaque key, and conditionally rewrites that message's `image_path`. If the database update fails, it removes the newly uploaded object. Successfully migrated rows no longer match the legacy-path query, so rerunning the command safely continues an interrupted migration.

The source files are intentionally retained. Archive or delete them only after checking historical messages through the Gateway and preserving the pre-cutover database backup.

## Existing Post and Comment image cutover

Run this migration before deploying the Posts version that no longer serves its local upload directory:

```powershell
Copy-Item .\social_network.db .\social_network.pre-post-media.db
docker compose stop post-service
docker compose up -d media
docker compose run --rm media-init
docker compose run --rm --entrypoint ./copy-post-media `
  -v "${PWD}/services/posts/uploads/posts:/migration/posts:ro" `
  post-service `
  -sqlite-path /app/social_network.db `
  -uploads-dir /migration/posts
docker compose up -d post-service gateway
```

The command handles both the `posts.image_path` and `comments.image_path` columns. Each object is uploaded under the record owner's prefix, after which a conditional SQLite update replaces the exact legacy path. Failed or concurrent reference updates trigger deletion of the newly uploaded object. Rerunning processes only references that still have legacy paths.

## Production configuration

The Chat service requires:

- `OBJECT_STORAGE_ENDPOINT`: S3 API host and optional port, without a URL scheme.
- `OBJECT_STORAGE_ACCESS_KEY` and `OBJECT_STORAGE_SECRET_KEY`: workload credentials supplied through a secret manager.
- `OBJECT_STORAGE_BUCKET`: a pre-provisioned bucket.
- `OBJECT_STORAGE_USE_TLS=true`.
- `OBJECT_STORAGE_PUBLIC_BASE_URL`: the CDN origin or public Gateway/bucket path stored in message records.

Use a service policy limited to its required key prefixes and operations. The broad local MinIO root credentials and anonymous bucket policy are development conveniences, not production policy. Private media will require authorization-aware delivery or short-lived signed URLs rather than anonymous reads.

## Remaining media operations work

All four media domains now use object storage, and the remaining upload bind mounts and static file handlers have been removed. The next operational work is to add object-retention/orphan-cleanup jobs, provision per-service least-privilege credentials, and exercise backup/restore procedures for object metadata and service databases.
