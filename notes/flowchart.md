# Social Network Project - Complete Flowchart



---

## Project Phases

---

## System Architecture

```
                                    ┌─────────────────────────┐
                                    │      CLIENT SIDE        │
                                    │    (Browser/Mobile)     │
                                    └─────────────────────────┘
                                                │
                                                ▼
    ┌───────────────────────────────────────────────────────────────────────────────────────────────────────┐
    │                                  STATIC FRONTEND                                                      │
    │                            (HTML/CSS/JavaScript - Vanilla JS)                                        │
    │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐    │
    │  │   auth.js   │ │  users.js   │ │  posts.js   │ │  groups.js  │ │   chat.js   │ │  events.js  │    │
    │  │   Login/    │ │  Profile    │ │    Feed     │ │   Groups    │ │    Chat     │ │   Events    │    │
    │  │  Register   │ │   Follow    │ │  Comments   │ │   Members   │ │  Messages   │ │  Responses  │    │
    │  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘    │
    └───────────────────────────────────────┬───────────────────────────────────────────────────────────────┘
                                            │ HTTP/HTTPS + WebSocket
                                            ▼
    ┌───────────────────────────────────────────────────────────────────────────────────────────────────────┐
    │                              MICROSERVICES ARCHITECTURE                                               │
    │                             (Docker Compose Network Bridge)                                          │
    └───────────────────────────────────────┬───────────────────────────────────────────────────────────────┘
                                            │
           ┌────────────────┬───────────────┼──────────────┬────────────────┬──────────────┐
           │                │               │              │                │              │
           ▼                ▼               ▼              ▼                ▼              ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │Auth Service  │ │User Service  │ │Post Service  │ │Group Service │ │Chat Service  │ │Notification  │
    │ Port: 8081   │ │ Port: 8082   │ │ Port: 8083   │ │ Port: 8084   │ │ Port: 8085   │ │Service       │
    │              │ │              │ │              │ │              │ │              │ │Port: 8086    │
    │ • Auth       │ │ • Profiles   │ │ • Posts      │ │ • Groups     │ │ • WebSocket  │ │ • WebSocket  │
    │ • Sessions   │ │ • Following  │ │ • Comments   │ │ • Events     │ │ • Direct     │ │ • Real-time  │
    │ • Token      │ │ • Followers  │ │ • Feed       │ │ • Members    │ │   Messages   │ │   Alerts     │
    │   Verify     │ │ • Search     │ │ • Privacy    │ │ • Requests   │ │ • Group Chat │ │ • Push       │
    └──────┬───────┘ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘
           │                │               │              │                │              │
           └────────────────┴───────────────┴──────────────┴────────────────┴──────────────┘
                                            │
                                            ▼
                              ┌──────────────────────────────┐
                              │    SHARED SQLite DATABASE    │
                              │   social_network.db          │
                              │   (Volume Mount)             │
                              │                              │
                              │ • All services read/write    │
                              │ • Foreign key constraints    │
                              │ • PRAGMA journal_mode        │
                              └──────────────────────────────┘
```

---

## Database Schema Design

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     USERS       │◄──┐│     FOLLOWS     │   ┌│     POSTS       │◄──┐│    COMMENTS     │    │     GROUPS      │
├─────────────────┤   │├─────────────────┤   │├─────────────────┤   │├─────────────────┤    ├─────────────────┤
│ id (PK)         │   ││ id (PK)         │   ││ id (PK)         │   ││ id (PK)         │    │ id (PK)         │
│ username        │   ││ follower_id (FK)│───┘│ user_id (FK)    │───┘│ post_id (FK)    │    │ name            │
│ email           │   ││ following_id(FK)│    │ title           │    │ user_id (FK)    │    │ description     │
│ password_hash   │   ││ status          │    │ content         │    │ content         │    │ image_url       │
│ first_name      │   ││ created_at      │    │ image_path      │    │ image_path      │    │ creator_id (FK) │
│ last_name       │   │└─────────────────┘    │ privacy_level   │    │ created_at      │    │ created_at      │
│ date_of_birth   │   │                       │ created_at      │    └─────────────────┘    └─────────────────┘
│ avatar_path     │   │                       └─────────────────┘            │                       │
│ nickname        │   │                               │                      │                       │
│ about_me        │   │       ┌─────────────────┐    │      ┌──────────────────────┐         ┌─────────────────┐
│ is_public_      │   │       │  POST_VIEWERS   │◄───┘      │   MESSAGES           │         │ GROUP_MEMBERS   │
│   profile       │   │       ├─────────────────┤           ├──────────────────────┤         ├─────────────────┤
│ created_at      │   │       │ id (PK)         │           │ id (PK)              │         │ id (PK)         │
└─────────────────┘   │       │ post_id (FK)    │           │ sender_id (FK)       │─────────│ group_id (FK)   │
        │              │       │ user_id (FK)    │───────────│ recipient_id (FK)    │         │ user_id (FK)    │
        │              │       │ created_at      │           │ content              │         │ role            │
        │              └───────│                 │           │ is_read              │         │ status          │
        │                      └─────────────────┘           │ created_at           │         │ joined_at       │
        │                                                    └──────────────────────┘         └─────────────────┘
        │                                                            │                                 │
        │              ┌─────────────────┐            ┌──────────────────────┐            ┌─────────────────┐
        │              │    EVENTS       │            │  GROUP_MESSAGES      │            │ NOTIFICATIONS   │
        │              ├─────────────────┤            ├──────────────────────┤            ├─────────────────┤
        │              │ id (PK)         │            │ id (PK)              │            │ id (PK)         │
        │              │ group_id (FK)   │            │ group_id (FK)        │            │ user_id (FK)    │
        │              │ title           │            │ sender_id (FK)       │            │ type            │
        │              │ description     │            │ content              │            │ related_id      │
        │              │ event_time      │            │ created_at           │            │ content         │
        │              │ creator_id (FK) │            └──────────────────────┘            │ is_read         │
        │              │ created_at      │                    │                           │ created_at      │
        │              └─────────────────┘                    │                           └─────────────────┘
        │                      │                              │
        │              ┌─────────────────┐                    │          ┌─────────────────┐
        └──────────────│ EVENT_RESPONSES │────────────────────┘          │    SESSIONS     │
                       ├─────────────────┤                               ├─────────────────┤
                       │ id (PK)         │                               │ id (PK)         │
                       │ event_id (FK)   │                               │ user_id (FK)    │
                       │ user_id (FK)    │                               │ token           │
                       │ response        │                               │ expires_at      │
                       │ created_at      │                               │ created_at      │
                       └─────────────────┘                               └─────────────────┘
```

---

## API Architecture

### Core Services

```
┌──────────────────────────┐    ┌──────────────────────────┐    ┌──────────────────────────┐
│    AUTH SERVICE :8081    │    │    USER SERVICE :8082    │    │    POST SERVICE :8083    │
│                          │    │                          │    │                          │
│ POST   /register         │    │ GET    /health           │    │ GET    /health           │
│ POST   /login            │    │ GET    /profile          │    │                          │
│ POST   /logout           │    │ PUT    /profile          │    │ POST   /posts            │
│ GET    /session          │    │ GET    /profile/:id      │    │ GET    /posts            │
│ GET    /health           │    │ POST   /follow           │    │ GET    /posts/:id        │
│                          │    │ DELETE /follow           │    │ PUT    /posts/:id        │
│ Internal Endpoints:      │    │ GET    /followers        │    │ DELETE /posts/:id        │
│ GET /internal/verify-    │    │ GET    /following        │    │                          │
│     token                │    │ GET    /follow/status/:id│    │ POST   /comments         │
│ GET /internal/user/:id   │    │ GET    /follow/requests  │    │ GET    /comments         │
│                          │    │ POST   /follow/respond   │    │                          │
│ Middleware:              │    │ GET    /search           │    │ Middleware:              │
│ - Session Management     │    │ GET    /users/:id        │    │ - Auth Verification      │
│ - CORS                   │    │                          │    │ - CORS                   │
│ - Rate Limiting          │    │ Middleware:              │    │ - Rate Limiting          │
│ - Logging                │    │ - Auth Verification      │    │ - Logging                │
│                          │    │ - CORS                   │    │                          │
└──────────────────────────┘    │ - Rate Limiting          │    └──────────────────────────┘
                                │ - Logging                │
                                └──────────────────────────┘

┌──────────────────────────┐    ┌──────────────────────────┐    ┌──────────────────────────┐
│   GROUP SERVICE :8084    │    │    CHAT SERVICE :8085    │    │ NOTIFICATION SVC :8086   │
│                          │    │                          │    │                          │
│ GET    /health           │    │ GET    /health           │    │ GET    /health           │
│                          │    │                          │    │                          │
│ POST   /groups           │    │ WS     /ws               │    │ POST   /notifications    │
│ GET    /groups           │    │ GET    /chat/            │    │ GET    /notifications/   │
│ GET    /groups/:id       │    │        conversations     │    │        list              │
│ PUT    /groups/:id       │    │ GET    /chat/history/:id │    │ GET    /notifications/   │
│ POST   /groups/:id/      │    │ POST   /chat/send        │    │        unread-count      │
│        invite            │    │ POST   /chat/read/:id    │    │ PUT    /notifications/   │
│ POST   /groups/:id/      │    │ GET    /chat/unread      │    │        read/:id          │
│        request           │    │                          │    │ POST   /notifications/   │
│ GET    /groups/:id/      │    │ GET    /chat/groups/:id/ │    │        read-all          │
│        requests          │    │        history           │    │ DELETE /notifications/   │
│ POST   /groups/:id/      │    │ POST   /chat/groups/:id/ │    │        delete/:id        │
│        requests/respond  │    │        messages          │    │                          │
│ GET    /groups/:id/      │    │                          │    │ WS     /ws               │
│        members           │    │ Features:                │    │                          │
│ GET    /groups/:id/      │    │ - Direct messaging       │    │ Features:                │
│        events            │    │ - Group chat             │    │ - Real-time push         │
│ POST   /groups/:id/      │    │ - Message history        │    │ - Follow notifications   │
│        messages          │    │ - Read status            │    │ - Group invites          │
│ GET    /groups/:id/      │    │ - WebSocket connections  │    │ - Event notifications    │
│        messages          │    │                          │    │ - Message alerts         │
│                          │    │ Middleware:              │    │                          │
│ POST   /events           │    │ - Auth Verification      │    │ Middleware:              │
│ GET    /events/:id       │    │ - CORS                   │    │ - Auth Verification      │
│ POST   /events/respond   │    │ - Logging                │    │ - CORS                   │
│                          │    │                          │    │ - Logging                │
│ Middleware:              │    └──────────────────────────┘    └──────────────────────────┘
│ - Auth Verification      │
│ - CORS                   │
│ - Logging                │
└──────────────────────────┘
```

### Service Communication Pattern

```
┌─────────────┐
│   FRONTEND  │
└──────┬──────┘
       │
       ├──────────────► Auth Service (Login/Register) ──► Returns Session Token
       │
       ├──────────────► User Service (with token) ──► Validates via Auth Service
       │
       ├──────────────► Post Service (with token) ──► Validates via Auth Service
       │
       ├──────────────► Group Service (with token) ──► Validates via Auth Service
       │
       ├──────────────► Chat Service (with token + WS) ──► Validates via Auth Service
       │
       └──────────────► Notification Service (with token + WS) ──► Validates via Auth Service

Note: All services except Auth require token validation
Auth Service provides /internal/verify-token endpoint for other services
```

---

