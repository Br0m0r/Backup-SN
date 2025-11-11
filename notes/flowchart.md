# Social Network Project - Complete Flowchart

**Architecture Type:** True Microservices (Database per Service)  
**Communication:** HTTP APIs only (No Event Bus)  
**Status:** Migration in Progress

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
    │              │ │              │ │              │ │              │ │              │ │              │
    └──────┬───────┘ └──────┬───────┘ └──────┬───────┘ └──────┬───────┘ └──────┬───────┘ └──────┬───────┘
           │                │                │                │                │                │
           ▼                ▼                ▼                ▼                ▼                ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │   auth_      │ │   user_      │ │   post_      │ │   group_     │ │   chat_      │ │   notif_     │
    │   service.db │ │   service.db │ │   service.db │ │   service.db │ │   service.db │ │   service.db │
    │              │ │              │ │              │ │              │ │              │ │              │
    │ • users      │ │ • user_      │ │ • posts      │ │ • groups     │ │ • messages   │ │ • notif-     │
    │ • sessions   │ │   profiles   │ │ • comments   │ │ • group_     │ │ • group_     │ │   ications   │
    │              │ │ • follows    │ │ • post_      │ │   members    │ │   messages   │ │              │
    │              │ │              │ │   viewers    │ │ • events     │ │              │ │              │
    │              │ │              │ │ • user_cache │ │ • event_     │ │              │ │              │
    │              │ │              │ │   (denorm)   │ │   responses  │ │              │ │              │
    └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘
    
    Note: Services communicate via HTTP APIs only
    Each service owns and manages its own database
    No direct cross-service database access
```

---

## Database Schema Design (Per Service)

### Auth Service Database (auth_service.db)
```
┌─────────────────┐    ┌─────────────────┐
│     USERS       │    │    SESSIONS     │
├─────────────────┤    ├─────────────────┤
│ id (PK)         │◄───│ id (PK)         │
│ username        │    │ user_id (FK)    │
│ email           │    │ token           │
│ password_hash   │    │ expires_at      │
│ first_name      │    │ created_at      │
│ last_name       │    └─────────────────┘
│ date_of_birth   │
│ avatar_path     │    Owns: User authentication
│ nickname        │          Session management
│ about_me        │          Token validation
│ is_public_      │
│   profile       │
│ created_at      │
└─────────────────┘
```

### User Service Database (user_service.db)
```
┌─────────────────┐    ┌─────────────────┐
│  USER_PROFILES  │    │     FOLLOWS     │
├─────────────────┤    ├─────────────────┤
│ user_id (PK)    │◄──┐│ id (PK)         │
│ username        │   ││ follower_id (FK)│───┘
│ avatar_path     │   ││ following_id(FK)│
│ nickname        │   ││ status          │
│ about_me        │   ││ created_at      │
│ is_public       │   │└─────────────────┘
│ updated_at      │   │
└─────────────────┘   │
                      │
Owns: User profiles (synced from auth)
      Following relationships
      Privacy settings
      
Note: user_id references auth service
      (no FK constraint across services)
```

### Post Service Database (post_service.db)
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     POSTS       │◄───│    COMMENTS     │    │  POST_VIEWERS   │
├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│ id (PK)         │    │ id (PK)         │    │ id (PK)         │
│ user_id         │    │ post_id (FK)    │    │ post_id (FK)    │
│ title           │    │ user_id         │    │ user_id         │
│ content         │    │ content         │    │ created_at      │
│ image_path      │    │ image_path      │    └─────────────────┘
│ privacy_level   │    │ created_at      │
│ created_at      │    └─────────────────┘    ┌─────────────────┐
└─────────────────┘                           │   USER_CACHE    │
                                              ├─────────────────┤
Owns: Posts                                   │ user_id (PK)    │
      Comments                                │ username        │
      Privacy settings                        │ avatar_path     │
      Post viewers (almost_private)           │ updated_at      │
                                              └─────────────────┘
Note: user_id is just an integer
      User data fetched via API or cached
```

### Group Service Database (group_service.db)
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     GROUPS      │◄───│  GROUP_MEMBERS  │    │     EVENTS      │
├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│ id (PK)         │    │ id (PK)         │    │ id (PK)         │
│ name            │    │ group_id (FK)   │    │ group_id (FK)   │
│ description     │    │ user_id         │    │ title           │
│ image_url       │    │ role            │    │ description     │
│ creator_id      │    │ status          │    │ event_time      │
│ created_at      │    │ joined_at       │    │ creator_id      │
└─────────────────┘    └─────────────────┘    │ created_at      │
                                              └─────────────────┘
                                                      │
                                              ┌─────────────────┐
                                              │ EVENT_RESPONSES │
                                              ├─────────────────┤
                                              │ id (PK)         │
                                              │ event_id (FK)   │
                                              │ user_id         │
                                              │ response        │
                                              │ created_at      │
                                              └─────────────────┘

Owns: Groups, Members, Events, Event Responses
Note: user_id and creator_id reference auth service (no FK)
```

### Chat Service Database (chat_service.db)
```
┌──────────────────────┐    ┌──────────────────────┐
│      MESSAGES        │    │   GROUP_MESSAGES     │
├──────────────────────┤    ├──────────────────────┤
│ id (PK)              │    │ id (PK)              │
│ sender_id            │    │ group_id             │
│ recipient_id         │    │ sender_id            │
│ content              │    │ content              │
│ is_read              │    │ created_at           │
│ created_at           │    └──────────────────────┘
└──────────────────────┘

Owns: Direct messages
      Group messages
      
Note: sender_id, recipient_id, group_id 
      are just integers (no FK constraints)
```

### Notification Service Database (notif_service.db)
```
┌─────────────────┐
│ NOTIFICATIONS   │
├─────────────────┤
│ id (PK)         │
│ user_id         │
│ type            │
│ related_id      │
│ content         │
│ is_read         │
│ created_at      │
└─────────────────┘

Owns: All notifications

Note: user_id is just an integer (no FK)
      related_id references entities in other services
```

**Key Changes from Shared DB:**
- No foreign key constraints across services
- User data denormalized/cached where needed
- Each service owns its data exclusively
- Inter-service references are just integers

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

### Service Communication Pattern (True Microservices - HTTP Only)

```
┌─────────────┐
│   FRONTEND  │
└──────┬──────┘
       │
       ├──────────────► Auth Service (Login/Register) ──► Returns Session Token
       │
       ├──────────────► User Service (with token) ──► Validates via Auth Service API
       │                                            ──► Fetches user data via Auth Service API
       │
       ├──────────────► Post Service (with token) ──► Validates via Auth Service API
       │                                           ──► Fetches user info via Auth/User Service API
       │
       ├──────────────► Group Service (with token) ──► Validates via Auth Service API
       │                                            ──► Fetches user info via Auth Service API
       │
       ├──────────────► Chat Service (with token + WS) ──► Validates via Auth Service API
       │                                                ──► Fetches user info via Auth Service API
       │
       └──────────────► Notification Service (with token + WS) ──► Validates via Auth Service API


Service-to-Service Communication (Internal APIs):
┌───────────────────────────────────────────────────────────────────────────┐
│                                                                           │
│  Auth Service provides:                                                  │
│    GET /internal/verify-token     - Token validation                    │
│    GET /internal/users/:id        - User basic info                     │
│    GET /internal/users/batch      - Multiple users at once              │
│    POST /internal/users/deleted   - Notify when user deleted            │
│                                                                           │
│  User Service provides:                                                  │
│    GET /internal/users/:id        - Extended profile info               │
│    GET /internal/followers/:id    - User's followers                    │
│    GET /internal/following/:id    - User's following list               │
│                                                                           │
│  Post Service provides:                                                  │
│    GET /internal/posts/user/:id   - Posts by user (for deletion)        │
│    POST /internal/users/deleted   - Handle user deletion cascade        │
│                                                                           │
│  Group Service provides:                                                 │
│    GET /internal/groups/user/:id  - Groups user is in                   │
│    POST /internal/users/deleted   - Handle user deletion cascade        │
│                                                                           │
│  Chat Service provides:                                                  │
│    POST /internal/users/deleted   - Handle user deletion cascade        │
│                                                                           │
│  Notification Service provides:                                          │
│    POST /internal/notify          - Create notification from any service│
│    POST /internal/users/deleted   - Handle user deletion cascade        │
│                                                                           │
└───────────────────────────────────────────────────────────────────────────┘

Data Flow Example: Creating a Post
1. Frontend → POST /posts (to Post Service)
2. Post Service → GET /internal/verify-token (to Auth Service)
3. Post Service → Saves post in its own DB
4. Post Service → POST /internal/notify (to Notification Service) [optional]
5. Post Service → Returns response to Frontend

Data Flow Example: Viewing a Post
1. Frontend → GET /posts/:id (to Post Service)
2. Post Service → Retrieves post from its own DB (has user_id)
3. Post Service → GET /internal/users/:id (to Auth Service) [for username/avatar]
4. Post Service → Combines data and returns to Frontend

Data Flow Example: User Deletion
1. Frontend → DELETE /users/:id (to Auth Service)
2. Auth Service → Deletes from its own DB (users, sessions)
3. Auth Service → POST /internal/users/deleted (to ALL other services)
4. Each service deletes related data from their own DBs
5. Auth Service → Returns success to Frontend
```

---

## Migration Checklist

### Phase 1: Database Separation ✅ COMPLETED
- [x] Create 6 separate SQLite databases
- [x] Split tables into appropriate service databases
- [x] Remove foreign key constraints across services
- [x] Add user_cache tables where needed (Post, Group services)
- [x] Update docker-compose.yml with separate volumes

### Phase 2: Internal API Endpoints
- [ ] Auth Service: Add `/internal/users/:id` endpoint
- [ ] Auth Service: Add `/internal/users/batch` endpoint
- [ ] Auth Service: Keep `/internal/verify-token` endpoint
- [ ] Add `/internal/users/deleted` handlers in all services
- [ ] User Service: Add internal profile endpoints
- [ ] Test all internal APIs

### Phase 3: Update Service Logic
- [ ] Remove all cross-service SQL JOINs
- [ ] Replace with HTTP API calls
- [ ] Add caching for frequently accessed data
- [ ] Add retry logic for failed API calls
- [ ] Add timeout handling

### Phase 4: Docker Configuration
- [ ] Update docker-compose.yml with separate volumes
- [ ] Test container startup
- [ ] Verify network connectivity between services
- [ ] Test service-to-service API calls

### Phase 5: Testing
- [ ] Integration tests for service communication
- [ ] Test user deletion cascade
- [ ] Test post creation with user lookup
- [ ] Load testing for API calls
- [ ] Failure scenario testing (service down)

---

## Known Trade-offs

**Pros:**
- True service independence
- Can scale services separately
- Can deploy independently
- Clear ownership boundaries

**Cons:**
- Increased latency (HTTP calls)
- More complex error handling
- No ACID transactions across services
- Eventual consistency challenges
- More API endpoints to maintain

---

