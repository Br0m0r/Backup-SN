# Current Service Boundary Inventory

## Purpose

This is the Phase 0 inventory of the application's current data access and runtime dependencies. It records where process boundaries differ from actual ownership boundaries and provides the starting point for separating the shared database.

Update this document whenever a table, synchronous dependency, or public API moves between services.

## Runtime Services

| Service | Default port | External responsibilities | Synchronous dependencies |
| --- | ---: | --- | --- |
| Gateway | 8080 | Single public origin, API routing, WebSocket upgrades, frontend proxy | All application services, Frontend, Redis |
| Auth | 8081 | Registration, login, logout, session verification | Shared SQLite database |
| Users | 8082 | Profiles, follows, user search, avatars | Auth verification; Notification creation; shared SQLite |
| Posts | 8083 | Posts, feed, privacy, comments, post/comment images | Auth verification; Notification creation; shared SQLite; S3-compatible media storage |
| Groups | 8084 | Groups, membership, invitations, events, group image | Auth verification; Notification creation; shared SQLite |
| Chat | 8085 | Direct/group chat REST and WebSocket delivery, attachments | Auth verification; Notification creation; shared SQLite; S3-compatible media storage; Redis |
| Notifications | 8086 | Notification inbox and WebSocket delivery | Auth verification; service-owned PostgreSQL; Redis |

The browser now defaults to relative `/api/<service>` URLs through the Gateway. Direct `VITE_*` URL overrides remain available for isolated development. In Compose, only Gateway port 8080 is published to the host; application service ports are private to the bridge network.

Private services no longer emit permissive cross-origin headers. Chat and Notification WebSockets enforce the exact origins configured through `WEBSOCKET_ALLOWED_ORIGINS`; production should contain only the public HTTPS gateway origin.

The Gateway adds browser security headers, rejects known request bodies above the configured limit, strips internal credentials, and applies a Redis-backed per-client token bucket to `/api` traffic. The limiter is shared across Gateway replicas and falls back to a replica-local bucket during a runtime Redis outage. It intentionally ignores client-supplied forwarding headers because Compose publishes the Gateway directly. Trusted-proxy handling must be added when a trusted external load balancer is introduced.

All uploaded media now uses opaque, owner-scoped keys in S3-compatible object storage; the Gateway exposes read-only `/media` access while storage credentials remain private. Repeatable migration commands remain available for legacy Chat and post/comment files. The discarded development avatars and group images required no legacy copy, and their services no longer expose static file handlers or upload bind mounts.

All Go services use `docker/backend.Dockerfile`, with the service name and port supplied as build arguments. Notifications is the first database extraction: it uses a private PostgreSQL database, owns an embedded migration history, and no longer has a foreign key to the Users domain. The other five backend services still share SQLite. Backend, Gateway, and Frontend runtime images run as non-root users, include health checks, use explicit base-image release lines, and are built in CI. Fixed Compose container names were removed so they do not block future replica scaling.

## Current Table Access

The following table is based on SQL statements in each service. "Target owner" is the ownership selected in the microservices roadmap.

| Current table | Services accessing it | Target owner | Required separation |
| --- | --- | --- | --- |
| `users` | Auth, Users, Posts, Groups, Chat | Split between Auth and Users | Create an Auth account/credential record and a Users profile record. Replace joins in Posts, Groups, and Chat with API composition or event-fed projections. |
| `sessions` | Auth | Auth | Move with Auth credentials and expose token/JWKS contracts. |
| `follows` | Users, Posts, Chat | Users | Posts needs a feed/audience projection; Chat needs a conversation-permission contract or projection. |
| `posts` | Posts, Users | Posts | Users currently reads Posts for profile statistics; obtain this through a Posts API or event-fed counter. |
| `comments` | Posts | Posts | Move with Posts. |
| `post_viewers` | Posts | Posts | Move with Posts. External viewer IDs remain identity references, not cross-database foreign keys. |
| `groups` | Groups | Groups | Move with Groups. |
| `group_members` | Groups, Users, Chat | Groups | Users currently excludes existing members during invite search; Chat checks membership. Replace both with a Groups contract or projection. |
| `events` | Groups | Groups | Move with Groups. |
| `event_responses` | Groups | Groups | Move with Groups. |
| `messages` | Chat | Chat | Move with Chat. |
| `group_messages` | Groups and Chat | Chat | This currently has two writers and two APIs. Choose Chat as the sole message owner; Groups remains the membership authority. |
| `notifications` | Notifications only, in its service-owned PostgreSQL database | Notifications | Database extraction is complete. Notification creation still uses authenticated synchronous HTTP and should move to event consumption in Phase 3. |

## Highest-Priority Ownership Violations

### Shared identity record

Auth creates and reads `users`, while Users updates the same record with profile data. The table combines credentials and profile information, so neither service can currently migrate or deploy its schema independently.

Target split:

- Auth owns account ID, email/login identifier, password hash, account status, and refresh sessions.
- Users owns the profile keyed by the stable account ID: username/display data, avatar, biography, birth date, and privacy settings.
- A successful registration emits `UserRegistered`; Users creates its profile idempotently.

### Posts depends on Users and Follows

Posts joins `users` to render authors and joins `follows` to calculate feed visibility. After separation:

- Posts should keep an event-fed author projection for feed rendering.
- Users should publish follow changes.
- Posts should maintain the audience/read model needed for feed queries, or a future Feed service may own that projection.
- Post privacy remains enforced by Posts even if its decision uses locally projected relationship data.

### Chat depends on Users, Follows, and Groups

Chat queries profiles, relationship data, and group membership to build contacts and authorize messages. After separation:

- Display data should come from a local profile projection.
- Direct-message permission should use a versioned Users contract or local relationship projection.
- Group-message permission should use a Groups membership contract or local membership projection.
- Persisted messages must have one owner: Chat.

### Group messages have two owners

Both Groups and Chat expose group-message operations and write `group_messages`. This is the most direct violation of single-writer ownership.

Migration approach:

1. Keep the Chat API/WebSocket as the canonical message path.
2. Change the frontend's group history/send operations to Chat.
3. Have Chat authorize membership through Groups or a membership projection.
4. Remove group-message handlers and SQL from Groups.
5. Move `group_messages` into Chat's database.

### Notification creation is synchronous and publicly reachable

Users, Posts, Groups, and Chat call the Notification service synchronously through the common `notify` helper. The creation endpoint is not protected by user or service authentication.

Migration approach:

1. Protect the endpoint with service authentication as an interim safety fix. **Completed:** notification ingestion now requires `INTERNAL_SERVICE_TOKEN`, producers fail fast when it is absent, and calls have a bounded timeout.
2. Define domain events and a standard event envelope.
3. Add an outbox to each producing service.
4. Make Notifications an idempotent event consumer.
5. Remove the synchronous notification helper.

## Current Non-Database State

| State | Current location | Production target |
| --- | --- | --- |
| Avatars | S3-compatible object storage | Managed object storage/CDN |
| Post images | S3-compatible object storage | Managed object storage/CDN |
| Group images | S3-compatible object storage | Managed object storage/CDN |
| Chat attachments | S3-compatible object storage | Managed object storage/CDN |
| Chat connections/presence | Replica-local sockets plus expiring per-instance Redis presence and Pub/Sub fan-out | Persisted recovery plus managed Redis/broker presence and fan-out |
| Notification connections | Replica-local sockets plus Redis Pub/Sub fan-out | Persisted recovery plus managed Redis/broker fan-out |
| Auth validation cache | Each service's process memory | Local JWT verification; Redis only if opaque-token introspection remains |
| Rate limits | Redis-backed Gateway token bucket with replica-local degraded fallback | Managed Redis-backed distributed limits |

## Public API Contract Notes

Phase 0 keeps the existing service-specific API origins while contracts are stabilized.

Resolved during this inventory:

- Group search now uses `GET /groups?q=<query>` consistently in the frontend and backend.
- Search responses are normalized to the `{ groups: [] }` shape expected by `SuggestedGroups.vue`.
- Unused frontend methods for nonexistent group delete, group post, and deprecated group-response endpoints were removed. Group posts already use the Posts service.

Still to address:

- Normalize the existing Go formatting debt in a dedicated mechanical change, then enforce `gofmt` in CI.
- Extract route construction from each `main.go` so API contracts can be integration tested without launching a process.
- Publish OpenAPI definitions and define compatibility/versioning rules.
- Consolidate duplicate group-chat REST behavior under Chat.
- Standardize pagination and error codes, not only the JSON envelope.

## Extraction Order

The current access graph supports this order:

1. **Notifications:** already owns its only table; replace producers after adding event infrastructure.
2. **Media state:** move files to object storage independently of relational data.
3. **Chat messages:** isolate `messages`, then resolve the duplicate `group_messages` writer.
4. **Posts:** introduce profile and relationship projections before moving its tables.
5. **Groups:** introduce profile projections and a membership contract for Chat/Users.
6. **Auth and Users:** split the combined `users` record last, once stable identity events and profile contracts exist.

Each extraction must finish with database credentials that cannot access another service's data.
