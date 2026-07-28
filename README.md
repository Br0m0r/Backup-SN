# Social Network

A Facebook-like social network built with hybrid microservices architecture, featuring real-time communication, groups, events, and notifications.

## Features

- **Authentication**: Session-based login with secure cookie management
- **Profiles**: Public/private profiles with followers system
- **Posts**: Create posts with images, comments, and privacy controls (public/private/almost private)
- **Groups**: Create/join groups with events, chat rooms, and member management
- **Real-time Chat**: Private messaging and group chats via WebSockets
- **Notifications**: Live notifications for follow requests, group invites, and events
- **Follow System**: Send/accept follow requests with automatic acceptance for public profiles

## Tech Stack

**Frontend:**
- Vue.js 3 with Composition API
- Vite
- WebSocket for real-time features

**Backend:**
- Go microservices (Auth, Users, Posts, Groups, Chat, Notifications)
- Service-owned PostgreSQL for Notifications, Chat messages, Posts, and Groups; shared SQLite for the remaining Auth/Users identity state
- S3-compatible object storage for avatars, group images, post/comment media, and Chat attachments
- Redis for distributed Gateway rate limits, Chat presence, and WebSocket fan-out
- Gorilla WebSocket
- Docker containerization

## Prerequisites

- Docker
- Docker Compose

## How to Run

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd social-network
   ```

2. **Create local environment configuration**
   ```powershell
   Copy-Item .env.example .env
   ```
   Replace `INTERNAL_SERVICE_TOKEN` in `.env` with a random value containing at least 32 characters. The same value authenticates internal notification requests between local containers. `WEBSOCKET_ALLOWED_ORIGINS` must list the exact browser origins allowed to connect; use only the public HTTPS application origin in production.

3. **Start the application**
   ```bash
   docker compose up -d --build
   ```
   The repository intentionally contains no pre-populated development database or uploaded user data. Compose creates disposable named volumes, runs a one-shot job that applies the remaining shared SQLite schema, waits for each service-owned PostgreSQL database and applies its schema, then initializes the private MinIO media bucket before starting dependent services.

   Check startup status with:
   ```bash
   docker compose ps --all
   ```

4. **Access the application**
   - Application (frontend and APIs through the gateway): http://localhost:8080

For frontend-only development with `npm run dev`, Vite proxies `/api` HTTP and WebSocket traffic to `http://localhost:8080`. Set `GATEWAY_URL` to override that development target.

Gateway request limits can be adjusted with `GATEWAY_RATE_LIMIT_RPS`, `GATEWAY_RATE_LIMIT_BURST`, and `GATEWAY_MAX_BODY_BYTES`. The token bucket is stored atomically in Redis and shared by every Gateway replica. If Redis becomes unavailable after startup, each Gateway temporarily falls back to its replica-local bucket and logs the degraded state.

Backend containers run as UID `10001`. The transitional Auth/Users SQLite database is held in the `shared-sqlite-data` named volume rather than a tracked or host-bound file. Notifications, Chat message state, Posts, and Groups use service-owned PostgreSQL, and all uploaded media uses object storage; application containers have no database-file or upload-directory bind mounts.
   
## Default Structure

```
social-network/
├── frontend/           # Vue.js frontend
├── services/           # Go microservices
│   ├── auth/
│   ├── users/
│   ├── posts/
│   ├── groups/
│   ├── chat/
│   ├── notifications/
│   └── gateway/
├── db/                 # Database migrations
└── docker-compose.yml  # Docker orchestration
```

## Architecture Documentation

- [Microservices and Deployment Roadmap](docs/microservices-deployment-roadmap.md) - target architecture, required changes, migration phases, and production deployment preparation.
- [Current Service Boundary Inventory](docs/current-service-boundaries.md) - current table access, cross-service dependencies, and extraction order.
- [Notification PostgreSQL Migration Runbook](docs/notification-postgresql-migration.md) - schema migration, SQLite data copy, verification, cutover, and rollback.
- [Groups PostgreSQL Migration Runbook](docs/groups-postgresql-migration.md) - group/member/event schema migration, verified SQLite data copy, cutover, and rollback.
- [Media Object Storage Migration](docs/media-object-storage-migration.md) - Chat media cutover, legacy-file migration, security model, and remaining domains.
- [Redis Realtime State](docs/redis-realtime-state.md) - distributed rate limiting, presence, WebSocket fan-out, failure behavior, and production configuration.

## Stopping the Application

```bash
docker compose down
```

To remove volumes (database data):
```bash
docker compose down -v
```
