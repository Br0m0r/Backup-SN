# Social Network API Documentation

This directory contains comprehensive API documentation for all microservices in the social network platform.

---

## Documentation Files

### Service-Specific Documentation

- **[AUTH_SERVICE.md](./AUTH_SERVICE.md)** - Authentication & Session Management (Port 8081)
- **[USERS_SERVICE.md](./USERS_SERVICE.md)** - User Profiles, Following & Search (Port 8082)
- **[POSTS_SERVICE.md](./POSTS_SERVICE.md)** - Posts & Comments (Port 8083)
- **[GROUPS_SERVICE.md](./GROUPS_SERVICE.md)** - Groups, Events & Group Messaging (Port 8084)
- **[CHAT_SERVICE.md](./CHAT_SERVICE.md)** - Real-time Chat & Messaging (Port 8085)
- **[NOTIFICATIONS_SERVICE.md](./NOTIFICATIONS_SERVICE.md)** - Real-time Notifications (Port 8086)

---

## Quick Reference

### Service Ports

| Service | Port | Description |
|---------|------|-------------|
| Auth | 8081 | User authentication and sessions |
| Users | 8082 | User profiles and relationships |
| Posts | 8083 | Posts and comments |
| Groups | 8084 | Groups and events |
| Chat | 8085 | Real-time messaging |
| Notifications | 8086 | Real-time notifications |

---

## Common Patterns

### Authentication

Most endpoints require a Bearer token in the Authorization header:

```http
Authorization: Bearer <token>
```

Obtain tokens from:
- `POST /register` (Auth Service)
- `POST /login` (Auth Service)

---

### Response Format

#### Success Response
```json
{
  "success": true,
  "data": { ... }
}
```

#### Error Response
```json
{
  "success": false,
  "error": "Error message"
}
```

---

### WebSocket Authentication

For WebSocket connections (Chat, Notifications):

**Via Query Parameter:**
```
ws://localhost:8085/ws?token=<bearer_token>
```

**Via Header:**
```
Authorization: Bearer <token>
```

---

## Service Dependencies

```
┌─────────────┐
│   Frontend  │
└──────┬──────┘
       │
       ├──────────────┐
       │              │
┌──────▼──────┐  ┌───▼────────────┐
│   Auth      │  │  Users         │
│  (8081)     │  │  (8082)        │
└─────────────┘  └────────────────┘
       │              │
       └──────┬───────┘
              │
    ┌─────────┼──────────┬──────────────┐
    │         │          │              │
┌───▼────┐ ┌─▼─────┐ ┌──▼────────┐ ┌───▼────────────┐
│ Posts  │ │Groups │ │   Chat    │ │ Notifications  │
│ (8083) │ │(8084) │ │  (8085)   │ │    (8086)      │
└────────┘ └───────┘ └───────────┘ └────────────────┘
```

- **Auth Service**: Core service for authentication, used by all other services
- **Users Service**: Manages profiles and relationships
- **Posts, Groups, Chat**: Feature services that depend on Auth and Users
- **Notifications**: Receives events from all feature services

---

## Getting Started

### 1. Register a User
```bash
curl -X POST http://localhost:8081/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "johndoe",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe",
    "date_of_birth": "1990-01-01"
  }'
```

### 2. Login
```bash
curl -X POST http://localhost:8081/login \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "johndoe",
    "password": "password123"
  }'
```

### 3. Use Token
```bash
curl -X GET http://localhost:8082/profile \
  -H "Authorization: Bearer <your_token>"
```

---

## Rate Limiting

Rate limits are applied to prevent abuse:

| Endpoint | Limit |
|----------|-------|
| Register | 100/min per IP |
| Login | 100/min per IP |
| Follow/Unfollow | 100/min per user |
| Create Post | 100/min per user |
| Create Comment | 100/min per user |

---

## Real-Time Features

### WebSocket Services

Both Chat and Notifications services support WebSocket connections for real-time updates:

- **Chat WebSocket**: `/ws` on port 8085
- **Notifications WebSocket**: `/ws` on port 8086

### HTTP Fallback

All real-time features have HTTP fallback endpoints for clients that don't support WebSockets.

---

## Privacy & Permissions

### User Privacy
- **Public Profiles**: Anyone can follow and view content
- **Private Profiles**: Follow requests must be accepted

### Post Privacy
- **Public**: Visible to all users
- **Private**: Visible to followers only
- **Almost Private**: Visible to selected users only

### Group Permissions
- **Admin**: Create group, invite members, manage requests
- **Member**: View content, send messages, create events

---

## Error Codes

| Code | Description |
|------|-------------|
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Invalid or missing token |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource doesn't exist |
| 409 | Conflict - Duplicate resource |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error |

---

## Support & Resources

- **Full API Endpoints**: See individual service documentation files
- **Example Requests**: Included in each service documentation
- **WebSocket Examples**: See CHAT_SERVICE.md and NOTIFICATIONS_SERVICE.md

---

**Last Updated**: November 8, 2025
