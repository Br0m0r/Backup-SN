# Microservices and Deployment Roadmap

## Purpose

This document describes how to evolve the current application into independently deployable services and prepare it for production deployment.

The existing codebase already has six separate Go processes, Dockerfiles, and service-specific HTTP APIs. However, it is not yet a fully decoupled microservice system because all services share one SQLite database, several services query data owned by other domains, uploaded files live on container filesystems, and real-time connection state exists only in process memory.

The goal is not merely to create more containers. A service should eventually be able to:

- Own and migrate its data without coordinating a release with every other service.
- Be deployed, scaled, and rolled back independently.
- Communicate through explicit, versioned contracts rather than shared tables.
- Tolerate temporary failures of other services.
- Run multiple replicas without losing correctness or real-time delivery.
- Be observed and operated in production.

## Current Architecture

```text
Browser
  |-- Auth API          :8081
  |-- Users API         :8082
  |-- Posts API         :8083
  |-- Groups API        :8084
  |-- Chat API/WS       :8085
  `-- Notifications/WS  :8086

All Go services --> one shared social_network.db
Uploads         --> bind-mounted local directories
Service auth    --> synchronous calls to Auth, cached in each process
Notifications  --> synchronous HTTP calls to Notification service
WebSockets      --> in-memory connection maps
```

This is best described as a distributed application with a shared database. It is a useful starting point, but its failure and deployment boundaries do not yet match its process boundaries.

## Target Architecture

```text
                          +------------------+
Browser / mobile client ->| API gateway/BFF  |
                          +--------+---------+
                                   |
             +----------+----------+----------+----------+
             |          |          |          |          |
           Auth       Users      Posts      Groups      Chat
             |          |          |          |          |
          auth DB    users DB   posts DB   groups DB   chat DB
             |          |          |          |          |
             +----------+----- event broker --+----------+
                                   |
                            Notifications
                                   |
                            notifications DB

Shared infrastructure only:
- Event broker for domain events
- Redis for ephemeral presence, rate limits, and WebSocket fan-out
- Object storage for user-uploaded media
- Central logs, metrics, and traces
```

The databases may initially be separate PostgreSQL databases on one managed PostgreSQL server. Physical database servers do not need to be split on day one; ownership and credentials are the important boundaries.

## Proposed Service Ownership

| Service | Owns | May expose or publish |
| --- | --- | --- |
| Auth | Credentials, sessions/refresh tokens, identity lifecycle | Token/JWKS endpoints; `UserRegistered`, `UserDisabled` |
| Users | Profiles, privacy settings, follows and follow requests | Profile API; `ProfileUpdated`, `FollowRequested`, `FollowAccepted` |
| Posts | Posts, comments, post viewers, post privacy rules | Post/feed API; `PostCreated`, `CommentCreated` |
| Groups | Groups, memberships, invitations, events and responses | Group API; `MemberJoined`, `GroupUpdated`, `EventCreated` |
| Chat | Direct and group messages, read state | Chat API/WebSocket; `MessageCreated` when needed |
| Notifications | Notification inbox and delivery preferences | Notification API/WebSocket; consumes domain events |

Cross-service database foreign keys must disappear. References such as `post.user_id` remain as stable identity values, but another service's database is responsible for validating and describing that identity.

### Data needed from another service

Use one of these patterns deliberately:

1. **Request-time composition:** the gateway or calling service requests the required data from its owner. Use this for fresh data on low-volume paths.
2. **Event-fed local projection:** a service stores a small local read model, such as author display name and avatar, updated from events. Use this for feeds and other high-volume reads.
3. **Immutable snapshot:** store the display data that was correct when an item was created. Use only when historical presentation should not change.

Do not replace SQL joins with long chains of synchronous service calls. That creates fragile distributed joins and poor latency.

## Required Changes

### 1. Establish contracts and service boundaries

- Inventory every SQL query and assign every table to exactly one service.
- Remove queries that read or mutate another service's tables.
- Define versioned HTTP contracts, preferably with OpenAPI.
- Standardize error envelopes, pagination, timestamps, IDs, and validation rules.
- Introduce generated or contract-tested clients for internal HTTP calls.
- Keep shared Go code limited to infrastructure helpers. Do not share domain models or database packages across services.
- Give every service its own migration directory and migration command.

Completion criteria:

- Every table has one owning service.
- No service account can access another service's database.
- API and event changes have compatibility rules.

### 2. Replace shared SQLite with service-owned PostgreSQL data

SQLite is appropriate for local or single-process applications, but six processes with connection pools writing one file are a production bottleneck and operational risk.

- Move to PostgreSQL.
- Start with one PostgreSQL cluster and a separate database, role, and migration history per service.
- Configure services with `DATABASE_URL`, pool limits, connection timeouts, and TLS settings.
- Run migrations as a release job, not independently from every application replica at startup.
- Add backup, point-in-time recovery, restore testing, and retention policies.
- Avoid distributed transactions. Operations spanning services should use events and compensating actions.

Suggested migration sequence:

1. Create service-owned PostgreSQL schemas/databases and roles.
2. Copy existing data with repeatable migration scripts and verify row counts/checksums.
3. Temporarily stop writes for the domain being moved or capture changes through an outbox/change log.
4. Switch one service to its new database.
5. Observe and verify it before removing the old tables.

Avoid long-lived dual writes from application code; they are difficult to make reliable.

### 3. Redesign authentication and service authorization

The current five-minute token cache means a revoked session may remain accepted by each service until its local entry expires. It also makes Auth availability part of a cache miss on every service.

Recommended design:

- Auth issues short-lived, asymmetrically signed access tokens.
- Services validate access tokens locally using Auth's public keys/JWKS.
- Auth stores and rotates longer-lived refresh tokens securely.
- Include only stable authorization claims in access tokens; retrieve rapidly changing permissions from their owner.
- Rotate signing keys and support overlapping old/new keys.
- Use HttpOnly, Secure, SameSite cookies for browser refresh/session credentials where practical. Avoid placing long-lived credentials in `localStorage`.
- Use workload identity, mTLS, or short-lived service credentials for internal calls.
- Protect internal endpoints. The current notification creation endpoint must not accept unauthenticated public traffic.

The gateway should authenticate external requests, but each service must still enforce authorization for its own resources.

### 4. Add an API gateway or backend-for-frontend

The frontend currently knows every service URL and calls every exposed port directly.

- Put a gateway or ingress in front of the APIs.
- Expose one public origin, for example `/api/auth`, `/api/users`, `/api/posts`, and `/ws/chat`.
- Keep databases, brokers, Redis, and service ports on private networks.
- Centralize TLS termination, request IDs, coarse rate limits, body-size limits, and CORS policy.
- Keep domain authorization and input validation inside services.
- Use the gateway/BFF to compose UI-specific responses where appropriate.

This simplifies frontend configuration and prevents internal topology from becoming a public contract.

### 5. Introduce reliable asynchronous events

Notifications and read-model updates should not block user-facing requests on another service's availability.

- Add a broker such as NATS, RabbitMQ, Kafka, or a cloud-managed equivalent. Choose based on operational needs; a small system generally does not need Kafka initially.
- Define versioned event envelopes containing event ID, type, version, timestamp, producer, correlation ID, and payload.
- Publish facts in past tense, such as `FollowRequested` or `CommentCreated`.
- Use the transactional outbox pattern so a database change and its event cannot diverge.
- Make consumers idempotent and record processed event IDs.
- Configure retries, exponential backoff, dead-letter handling, and replay procedures.
- Do not put secrets or unnecessary personal data in events.

The Notification service should consume events rather than requiring every domain service to call its open HTTP endpoint.

### 6. Externalize uploaded media

Local upload directories prevent safe rescheduling and horizontal scaling.

- Store avatars, post images, group images, and chat attachments in S3-compatible object storage.
- Use opaque object keys rather than user-provided filenames.
- Prefer short-lived pre-signed uploads or a dedicated media endpoint.
- Store only object keys and metadata in service databases.
- Serve public media through a CDN; use signed URLs for private media.
- Validate size and MIME type, generate safe variants, remove metadata where appropriate, and add malware scanning if the threat model requires it.
- Define orphan cleanup and retention behavior.

MinIO can provide a compatible local-development environment.

### 7. Make WebSockets horizontally scalable

Chat and Notification hubs currently keep a single in-memory map of connected users. A second replica would not know about connections on the first.

- Use Redis or broker-backed pub/sub for cross-replica delivery.
- Store ephemeral presence with expiry in Redis.
- Persist messages/notifications before attempting delivery.
- Treat WebSocket delivery as an optimization; clients must recover missed data through REST cursors or sequence numbers after reconnecting.
- Implement bounded queues, slow-client handling, reconnect backoff, heartbeat settings, and connection limits.
- Validate allowed WebSocket origins instead of accepting all origins.
- Sticky sessions may reduce churn but must not be required for correctness.

### 8. Production configuration and resilience

- Use typed per-service configuration loaded from environment variables or a secrets manager.
- Separate development, test, staging, and production configuration.
- Never bake secrets into images or commit them to the repository.
- Add explicit connect/read/write/idle timeouts to HTTP servers and clients.
- Add bounded retries only for safe or idempotent operations, with jitter and deadlines.
- Add idempotency keys for retryable create operations.
- Handle graceful shutdown: stop accepting traffic, drain requests/connections, then close resources.
- Provide `/livez` and `/readyz`; readiness should reflect whether a replica can safely serve traffic.
- Set CPU/memory requests and limits and bound database pools relative to replica count.
- Use UTC internally and ISO-8601 timestamps at boundaries.

### 9. Observability and operations

- Emit structured JSON logs with service, environment, request ID, trace ID, user ID where appropriate, route, status, and latency.
- Propagate correlation and trace context across HTTP and events.
- Add OpenTelemetry traces and metrics.
- Track request rate, error rate, latency, saturation, database pool health, broker lag, WebSocket connections, event retries, and dead-letter counts.
- Centralize logs and define dashboards and alerts around user-visible symptoms.
- Redact credentials, tokens, message contents, and other sensitive fields.
- Write runbooks for database restore, failed migrations, dead-letter replay, signing-key rotation, and rollback.

### 10. Testing and delivery safety

There is no visible automated test suite covering the current architecture. Decoupling without a safety net is high risk.

- Add unit tests for domain rules and authorization.
- Add repository tests against real PostgreSQL containers.
- Add API integration and consumer-driven contract tests.
- Add event schema compatibility tests and idempotency tests.
- Add end-to-end tests for registration, login, follow flow, post privacy, groups, chat, and notifications.
- Run race detection for concurrent Go code, especially WebSocket hubs.
- Add static analysis, dependency scanning, image scanning, and secret scanning.
- Test backup restoration and deployment rollback, not only application behavior.

## Deployment Foundation

### Container images

Update every Dockerfile to:

- Use a consistent, pinned Go builder and minimal pinned runtime image.
- Build a statically linked binary where dependencies allow; moving away from SQLite removes the current CGO requirement.
- Run as a non-root user with a read-only root filesystem.
- Copy only the service binary and required certificates/configuration.
- Include build version/commit metadata and produce multi-architecture images if needed.
- Add `.dockerignore` and keep build contexts small.
- Do not store databases or uploads inside application containers.

### Local development

Create a development Compose stack containing:

- Gateway
- Six services
- Service-owned PostgreSQL databases/roles
- Event broker
- Redis
- MinIO
- Optional local observability stack

Provide checked-in example environment files with safe placeholders, seed commands, migration commands, and one command to start the stack.

### CI/CD

For each change:

1. Format, lint, test, and run security checks.
2. Build immutable images tagged with the commit SHA.
3. Generate an SBOM and scan the image.
4. Push to a container registry.
5. Deploy to staging and run smoke/contract tests.
6. Require approval or automated policy before production.
7. Run backward-compatible migrations before switching application traffic.
8. Use rolling or canary deployment with automated rollback signals.

Services should deploy independently; avoid rebuilding and redeploying all services for one domain change.

### Hosting strategy

Start with a managed container platform plus managed PostgreSQL, object storage, Redis, and broker services. Kubernetes is justified only when its operational benefits outweigh its substantial complexity. The architecture above works with managed container services, Kubernetes, or a smaller VM-based platform.

Use infrastructure as code for networks, databases, identities, secrets, DNS, TLS, storage, monitoring, and application services. Maintain at least staging and production environments.

## Phased Implementation Plan

### Phase 0: Stabilize and measure

- [ ] Add critical-path tests and baseline performance measurements.
- [x] Document current APIs, SQL ownership, and service dependencies.
- [x] Add request IDs, structured HTTP logging, server timeouts, and graceful shutdown.
- [x] Remove known API contract mismatches between frontend clients and backend routes.
- [x] Add CI for Go and frontend builds/tests.
- [ ] Upgrade Vite/esbuild through a tested major-version migration, then raise the CI audit threshold from critical to high.

Exit condition: current behavior is repeatably testable and failures can be diagnosed.

### Phase 1: Create a deployable edge

- [x] Protect notification ingestion with an interim internal service credential and bounded producer timeouts.
- [x] Remove permissive service CORS and enforce an explicit WebSocket origin allowlist.
- [x] Add a gateway/ingress and route all frontend traffic through one origin.
- [x] Make ports and URLs fully configurable.
- [x] Separate public and private networks/endpoints in the Compose deployment.
- [x] Harden CORS, WebSocket origins, edge rate limiting, request sizes, and interim service-to-service authentication.
- [x] Produce hardened, consistent non-root container images with health checks and CI builds.

Exit condition: the existing shared-database system can be deployed safely as an interim release.

### Phase 2: Externalize state

- [x] Move uploads to object storage.
  - [x] Move new and legacy Chat images to S3-compatible storage with user-scoped deletion.
  - [x] Move new and legacy post/comment images to S3-compatible storage with user-scoped deletion.
  - [x] Move avatars and group images with detected media types, opaque keys, durable reference updates, and safe replacement/deletion.
- [x] Add Redis for rate-limit/presence state and cross-replica WebSocket fan-out.
- [ ] Move SQLite data into PostgreSQL while initially preserving the logical schema if necessary.
  - [x] Extract Notifications into its own PostgreSQL database, migration history, release job, and verified SQLite copy command.
  - [x] Extract Chat message state into its own PostgreSQL database and replace its identity/follow reads with authenticated Users contracts.
  - [x] Extract Posts into its own PostgreSQL database after replacing Users/Follow joins with authenticated contracts.
  - [x] Make clean interim environments deterministic with a disposable SQLite volume and one-shot schema migration job.
  - [x] Extract Groups into its own PostgreSQL database after replacing cross-domain reads with authenticated contracts.
  - [ ] Extract Users and Auth in dependency order.
- [ ] Establish backups and verify restoration.

Exit condition: application containers are disposable and major services can run multiple replicas.

### Phase 3: Establish asynchronous communication

- [ ] Add the broker and event envelope.
- [ ] Implement an outbox publisher and idempotent consumer library/pattern.
- [ ] Convert notifications to domain-event consumers.
- [ ] Add retries, dead-letter handling, metrics, and replay tooling.

Exit condition: core writes do not fail merely because Notification service is unavailable.

### Phase 4: Split data ownership

- [x] Give Notifications its own database first; it has a relatively clear boundary.
- [x] Split Chat messages and delivery/read state.
  - [x] Make Chat the sole API and SQL owner for group messages.
  - [x] Replace Chat's direct `group_members` read with a versioned, service-authenticated Groups contract.
  - [x] Move direct and group messages into Chat-owned PostgreSQL with migrations and a verified copy command.
- [x] Split Posts and its privacy/read models.
- [x] Split Groups, membership, invitations, and events.
- [ ] Split Users/profile/follow data from Auth credentials and sessions.
- [ ] Remove all remaining cross-service database access and foreign keys.

Exit condition: every service uses credentials that grant access only to its own data.

### Phase 5: Independent production operations

- [ ] Deploy services independently with canary/rolling releases.
- [ ] Add tracing, dashboards, SLOs, alerts, and runbooks.
- [ ] Perform load, failure, restore, and rollback exercises.
- [ ] Add autoscaling policies based on suitable service metrics.
- [ ] Review privacy, retention, audit, and incident-response requirements.

Exit condition: each service can be deployed, scaled, restored, and rolled back without coordinating a full-system release.

## Recommended First Milestone

Do not begin by splitting all databases simultaneously. The first useful milestone is a production-capable version of the existing system:

1. Add tests and standard operational behavior.
2. Put one gateway in front of the services.
3. Move from shared SQLite to PostgreSQL.
4. Move files to object storage.
5. Make WebSocket delivery multi-replica safe.
6. Deploy this interim architecture to staging.

After that foundation works, introduce the broker and separate one domain at a time. This provides deployable value early and keeps data migration risk manageable.

Implementation status: Notifications, Chat, Posts, and Groups have completed data extraction. See their PostgreSQL migration runbooks for cutover procedures. Posts, Groups, and Chat consume versioned authenticated Users reads rather than accessing Users-owned tables. Redis shares Gateway rate limits, Chat presence, and Chat/Notification WebSocket fan-out across replicas; see [Redis Realtime State](redis-realtime-state.md). Only Auth and Users retain the shared SQLite database while their combined identity record is split. For clean interim environments, Compose owns that remaining SQLite state in a named volume and applies migrations through a one-shot release job before Auth starts.

## Decisions to Record Before Implementation

Create architecture decision records under `docs/adr/` for at least:

- Service boundaries and data ownership.
- PostgreSQL topology and migration approach.
- Gateway/BFF choice.
- Authentication/token model.
- Broker and event-delivery guarantees.
- Object storage and private-media access.
- Redis/presence and WebSocket fan-out design.
- Hosting platform and infrastructure-as-code tool.
- Observability platform and retention.

Each decision should record context, chosen option, alternatives, consequences, owner, and date. Tool and cloud-provider choices should follow these architectural decisions rather than define the architecture themselves.
