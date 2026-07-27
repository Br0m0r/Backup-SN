# Redis Realtime State

## Scope

Redis now owns only ephemeral coordination state:

- The Gateway's per-client token buckets.
- Chat online presence, recorded separately for every service instance.
- Cross-replica Chat WebSocket fan-out.
- Cross-replica Notification WebSocket fan-out.

Messages and notifications are persisted before delivery. Redis is not a system
of record, and clients recover missed events through the existing REST APIs.

## Local development

Compose starts a private Redis 8 service and waits for `redis-cli ping` before
starting Gateway, Chat, or Notifications. Redis has no published host port and
no durable volume because all of its current state is safe to rebuild.

```powershell
Copy-Item .env.example .env
docker compose up --build
```

The relevant settings are:

- `REDIS_URL`: connection URL, including TLS, credentials, and database when required.
- `REDIS_NAMESPACE`: prefix that isolates environments sharing one Redis deployment.
- `REDIS_OPERATION_TIMEOUT`: deadline for commands on request and hub paths.
- `REDIS_PRESENCE_TTL`: expiry applied to each Chat instance's presence record.
- `REDIS_PRESENCE_REFRESH`: heartbeat interval, which must be shorter than the TTL.

`SERVICE_INSTANCE_ID` may be supplied by the deployment platform. Otherwise the
service derives an ID from its hostname and process ID.

## Distributed rate limiting

The Gateway executes one atomic Lua token-bucket update using Redis server time.
Keys contain a SHA-256 digest of the direct client address rather than the raw
address. Limits remain configurable through `GATEWAY_RATE_LIMIT_RPS` and
`GATEWAY_RATE_LIMIT_BURST`.

Gateway startup fails when its configured Redis endpoint is unavailable. If an
established Redis connection fails at runtime, requests use the existing
replica-local bucket and the Gateway logs that enforcement is degraded. This
keeps a bounded local control in place without making every API request fail.

Compose exposes the Gateway directly, so it deliberately ignores untrusted
forwarding headers. Before placing a load balancer in front, configure an exact
trusted-proxy policy and derive the limiter identity only from forwarding
headers written by that proxy.

## Presence

Each Chat instance adds its own member to a sorted set for the connected user.
The score is an expiry timestamp calculated with Redis server time. Heartbeats
refresh the member while at least one local socket remains connected; the final
local disconnect removes only that instance's member. A user therefore remains
online when another browser socket or Chat replica is still connected.

Expired members are removed during presence reads. Losing Redis causes Chat to
fall back to its local connection map, which may produce an unnecessary offline
notification but does not lose the persisted message.

## WebSocket fan-out

Chat and Notifications publish JSON envelopes to separate namespaced channels.
The envelope includes the origin instance so a publisher can deliver locally
without receiving a duplicate from its own subscription. Subscribers retry with
bounded exponential backoff after an interruption.

Redis Pub/Sub is best-effort and does not replay missed events. Correctness
therefore depends on persisting the message or notification first and having
clients reload REST history/counts after reconnecting. A future durable broker
and outbox remain required for domain events in Phase 3.

Both hubs support multiple sockets for one user on the same replica. Slow
connections with a full bounded send queue are removed rather than blocking the
hub.

## Production requirements

- Use a managed highly available Redis service with TLS and workload-specific credentials.
- Restrict network access to Gateway, Chat, and Notifications.
- Use a unique namespace per environment.
- Set memory limits and an eviction policy appropriate for ephemeral keys.
- Alert on connection failures, fallback rate limiting, subscriber retries, and command latency.
- Exercise failover while multiple Gateway and WebSocket replicas are serving traffic.
- Do not place durable application records, session secrets, or message contents outside the transient Pub/Sub envelope in this Redis deployment.
