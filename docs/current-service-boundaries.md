# Current Service Boundary Inventory

## Purpose

This is the Phase 0 inventory of the application's current data access and runtime dependencies. It records where process boundaries differ from actual ownership boundaries and provides the starting point for separating the shared database.

Update this document whenever a table, synchronous dependency, or public API moves between services.

## Runtime Services

| Service | Default port | External responsibilities | Synchronous dependencies |
| --- | ---: | --- | --- |
| Gateway | 8080 | Single public origin, API routing, WebSocket upgrades, frontend proxy | All application services, Frontend, Redis |
| Auth | 8081 | Registration, login, logout, session verification | Shared SQLite database |
| Users | 8082 | Profiles, follows, user search, avatars | Auth verification; authenticated Posts and Groups reads; Notification creation; shared SQLite |
| Posts | 8083 | Posts, feed, privacy, comments, post/comment images | Auth verification; authenticated Users profile/relationship reads; Notification creation; service-owned PostgreSQL; S3-compatible media storage |
| Groups | 8084 | Groups, membership, invitations, events, group image | Auth verification; authenticated Users profile reads; Notification creation; service-owned PostgreSQL |
| Chat | 8085 | Direct/group chat REST and WebSocket delivery, attachments | Auth verification; authenticated Users profile/relationship reads; authenticated Groups membership reads; Notification creation; service-owned PostgreSQL; S3-compatible media storage; Redis |
| Notifications | 8086 | Notification inbox and WebSocket delivery | Auth verification; service-owned PostgreSQL; Redis |

The browser now defaults to relative `/api/<service>` URLs through the Gateway. Direct `VITE_*` URL overrides remain available for isolated development. In Compose, only Gateway port 8080 is published to the host; application service ports are private to the bridge network.

Private services no longer emit permissive cross-origin headers. Chat and Notification WebSockets enforce the exact origins configured through `WEBSOCKET_ALLOWED_ORIGINS`; production should contain only the public HTTPS gateway origin.

The Gateway adds browser security headers, rejects known request bodies above the configured limit, strips internal credentials, and applies a Redis-backed per-client token bucket to `/api` traffic. The limiter is shared across Gateway replicas and falls back to a replica-local bucket during a runtime Redis outage. It intentionally ignores client-supplied forwarding headers because Compose publishes the Gateway directly. Trusted-proxy handling must be added when a trusted external load balancer is introduced.

All uploaded media now uses opaque, owner-scoped keys in S3-compatible object storage; the Gateway exposes read-only `/media` access while storage credentials remain private. Repeatable migration commands remain available for legacy Chat and post/comment files. The discarded development avatars and group images required no legacy copy, and their services no longer expose static file handlers or upload bind mounts.

All Go services use `docker/backend.Dockerfile`, with the service name and port supplied as build arguments. Notifications, Chat, Posts, and Groups now have private PostgreSQL databases and embedded migration histories without cross-domain foreign keys. Posts, Groups, and Chat obtain profile/relationship data through Users contracts; Users obtains profile post lists through Posts and invite-search exclusions through Groups. Backend, Gateway, and Frontend runtime images run as non-root users, include health checks, use explicit base-image release lines, and are built in CI. Fixed Compose container names were removed so they do not block future replica scaling.

## Current Table Access

The following table is based on SQL statements in each service. "Target owner" is the ownership selected in the microservices roadmap.

| Current table | Services accessing it | Target owner | Required separation |
| --- | --- | --- | --- |
| `users` | Auth, Users | Split between Auth and Users | Posts, Groups, and Chat now use authenticated Users contracts. Split credentials from profile state. |
| `sessions` | Auth | Auth | Move with Auth credentials and expose token/JWKS contracts. |
| `follows` | Users | Users | Posts and Chat now use authenticated Users relationship contracts. |
| `posts` | Posts only, in its service-owned PostgreSQL database | Posts | Database extraction is complete; Users reads profile posts through the authenticated Posts contract. |
| `comments` | Posts only, in its service-owned PostgreSQL database | Posts | Database extraction is complete. |
| `post_viewers` | Posts only, in its service-owned PostgreSQL database | Posts | Database extraction is complete; viewer IDs are external identity references. |
| `groups` | Groups only, in its service-owned PostgreSQL database | Groups | Database extraction is complete. |
| `group_members` | Groups only, in its service-owned PostgreSQL database | Groups | Database extraction is complete. Chat uses the accepted-membership contract; Users uses the all-participants contract for invite search. |
| `events` | Groups only, in its service-owned PostgreSQL database | Groups | Database extraction is complete. |
| `event_responses` | Groups only, in its service-owned PostgreSQL database | Groups | Database extraction is complete. |
| `messages` | Chat only, in its service-owned PostgreSQL database | Chat | Database extraction is complete. External sender/recipient IDs are identity references without cross-database foreign keys. |
| `group_messages` | Chat only, in its service-owned PostgreSQL database | Chat | Database extraction is complete. Chat authorizes access through the Groups membership contract. |
| `notifications` | Notifications only, in its service-owned PostgreSQL database | Notifications | Database extraction is complete. Notification creation still uses authenticated synchronous HTTP and should move to event consumption in Phase 3. |

## Highest-Priority Ownership Violations

### Shared identity record

Auth creates and reads `users`, while Users updates the same record with profile data. The table combines credentials and profile information, so neither service can currently migrate or deploy its schema independently.

Target split:

- Auth owns account ID, email/login identifier, password hash, account status, and refresh sessions.
- Users owns the profile keyed by the stable account ID: username/display data, avatar, biography, birth date, and privacy settings.
- A successful registration emits `UserRegistered`; Users creates its profile idempotently.

### Posts depends on Users and Follows

**Resolved:** Posts no longer joins `users` or `follows`. It obtains batch profile summaries, profile search results, and accepted-following IDs through versioned Users contracts protected by `INTERNAL_SERVICE_TOKEN`. Users obtains profile post lists through the corresponding Posts contract.

Posts remains the privacy authority: it combines the Users relationship decision with its own `post_viewers` state. Event-fed profile and relationship projections remain the preferred future resilience optimization.

### Chat depends on Users, Follows, and Groups

**Resolved:** Chat no longer queries Users-owned tables. It hydrates conversation
profiles, obtains eligible contacts, and evaluates new direct-message
permissions through versioned Users contracts protected by
`INTERNAL_SERVICE_TOKEN`. Existing message history remains a Chat-owned
permission rule. Group-message authorization and recipient fan-out use the
service-authenticated Groups membership contract.

Event-fed local profile, relationship, and membership projections remain the
preferred future resilience optimization.

### Groups depends on Users

**Resolved:** Groups no longer joins `users`. Member lists, pending requests,
event creator names, and notification display names are hydrated through the
versioned Users profile contract protected by `INTERNAL_SERVICE_TOKEN`. Users
no longer reads `group_members`; its invite search obtains pending, invited,
and accepted participant IDs from
`GET /internal/v1/groups/{groupID}/participants`.

**Database extraction complete:** groups, memberships, invitations, events,
and responses now live in Groups-owned PostgreSQL with embedded migrations and
a verified SQLite copy command.

### Group message ownership

**Resolved:** Chat is the sole group-message API and SQL writer. The frontend's
history and HTTP fallback operations use Chat, and Groups no longer defines
group-message routes, handlers, service methods, models, or queries. Chat
authorizes send/history operations and resolves fan-out recipients through
Groups' `GET /internal/v1/groups/{groupID}/members[/{userID}]` contract, protected
by `INTERNAL_SERVICE_TOKEN`; it no longer reads `group_members`.

**Database extraction complete:** direct and group messages now live in
Chat-owned PostgreSQL with embedded migrations and a verified SQLite copy
command. Chat has no remaining runtime access to shared SQLite.

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
- Group chat history and HTTP fallback sends now use Chat; the duplicate Groups API and SQL writer were removed.

Still to address:

- Normalize the existing Go formatting debt in a dedicated mechanical change, then enforce `gofmt` in CI.
- Extract route construction from each `main.go` so API contracts can be integration tested without launching a process.
- Publish OpenAPI definitions and define compatibility/versioning rules.
- Standardize pagination and error codes, not only the JSON envelope.

## Extraction Order

The current access graph supports this order:

1. **Notifications:** already owns its only table; replace producers after adding event infrastructure.
2. **Media state:** move files to object storage independently of relational data.
3. **Chat messages:** database extraction and prerequisite synchronous Users/Groups contracts are complete; replace contracts with projections after event infrastructure exists.
4. **Posts:** database extraction and prerequisite synchronous contracts are complete; replace contracts with projections after event infrastructure exists.
5. **Groups:** database extraction and prerequisite synchronous contracts are complete; replace contracts with projections after event infrastructure exists.
6. **Auth and Users:** split the combined `users` record last, once stable identity events and profile contracts exist.

Each extraction must finish with database credentials that cannot access another service's data.
