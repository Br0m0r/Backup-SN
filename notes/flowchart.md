# Social Network Project - Complete Flowchart



---

## Project Phases

### Phase Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   PHASE 1       │───▶│   PHASE 2       │───▶│   PHASE 3       │───▶│   PHASE 4       │───▶│   PHASE 5       │
│ PLANNING        │    │ INFRASTRUCTURE  │    │ CORE SERVICES   │    │ FRONTEND        │    │ INTEGRATION     │
│ (Week 1)        │    │ (Week 2-3)      │    │ (Week 4-8)      │    │ (Week 6-10)     │    │ (Week 11-12)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
        │                        │                        │                        │                        │
        ▼                        ▼                        ▼                        ▼                        ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│• Requirements   │    │• Docker Setup   │    │• Auth Service   │    │• Component      │    │• API Integration│
│• Architecture   │    │• Database       │    │• User Service   │    │  Library        │    │• WebSocket      │
│• DB Design      │    │• Migrations     │    │• Post Service   │    │• Pages          │    │• Testing        │
│• API Specs      │    │• CI/CD          │    │• Group Service  │    │• State Mgmt     │    │• Deployment     │
│• Team Roles     │    │• Basic Auth     │    │• Chat Service   │    │• Responsive     │    │• Monitoring     │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
```

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
    │                                  FRONTEND CONTAINER                                                   │
    │                                     (Next.js)                                                        │
    │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐    │
    │  │    Auth     │ │   Profile   │ │   Posts     │ │   Groups    │ │    Chat     │ │Notifications│    │
    │  │   Pages     │ │    Page     │ │    Feed     │ │   Pages     │ │ Interface   │ │   Center    │    │
    │  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘    │
    └───────────────────────────────────────┬───────────────────────────────────────────────────────────────┘
                                            │ HTTP/HTTPS + WebSocket
                                            ▼
    ┌───────────────────────────────────────────────────────────────────────────────────────────────────────┐
    │                                 LOAD BALANCER / GATEWAY                                               │
    │                                      (Optional)                                                      │
    └───────────────────────────────────────┬───────────────────────────────────────────────────────────────┘
                                            │
                          ┌─────────────────┼─────────────────┐
                          │                 │                 │
                          ▼                 ▼                 ▼
    ┌─────────────────────────┐ ┌─────────────────────────┐ ┌─────────────────────────┐
    │   BACKEND CONTAINER     │ │   WEBSOCKET CONTAINER   │ │   DATABASE CONTAINER    │
    │      (Go Server)        │ │      (Go WebSocket)     │ │       (SQLite)          │
    │                         │ │                         │ │                         │
    │ ┌─────────────────────┐ │ │ ┌─────────────────────┐ │ │ ┌─────────────────────┐ │
    │ │   Microservices     │ │ │ │   Real-time Chat    │ │ │ │    Database         │ │
    │ │                     │ │ │ │   Notifications     │ │ │ │    Migrations       │ │
    │ │ • Auth Service      │ │ │ │   Live Updates      │ │ │ │    Data Storage     │ │
    │ │ • User Service      │ │ │ │                     │ │ │ │                     │ │
    │ │ • Post Service      │ │ │ └─────────────────────┘ │ │ └─────────────────────┘ │
    │ │ • Group Service     │ │ │                         │ │                         │
    │ │ • Notification Svc  │ │ └─────────────────────────┘ └─────────────────────────┘
    │ └─────────────────────┘ │
    └─────────────────────────┘
```

---

## Database Schema Design

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     USERS       │◄──┐│     FOLLOWS     │   ┌│     POSTS       │◄──┐│    COMMENTS     │    │     GROUPS      │
├─────────────────┤   │├─────────────────┤   │├─────────────────┤   │├─────────────────┤    ├─────────────────┤
│ id (PK)         │   ││ id (PK)         │   ││ id (PK)         │   ││ id (PK)         │    │ id (PK)         │
│ email           │   ││ follower_id (FK)│───┘│ user_id (FK)    │───┘│ post_id (FK)    │    │ name            │
│ password_hash   │   ││ following_id(FK)│    │ content         │    │ user_id (FK)    │    │ description     │
│ first_name      │   ││ status          │    │ image_path      │    │ content         │    │ creator_id (FK) │
│ last_name       │   ││ created_at      │    │ privacy_level   │    │ image_path      │    │ created_at      │
│ dob             │   │└─────────────────┘    │ created_at      │    │ created_at      │    │                 │
│ avatar_path     │   │                       └─────────────────┘    └─────────────────┘    └─────────────────┘
│ nickname        │   │                               │                       │                       │
│ about_me        │   │       ┌─────────────────┐    │       ┌─────────────────┐               ┌─────────────────┐
│ is_public       │   │       │  POST_VIEWERS   │◄───┘       │   MESSAGES      │               │ GROUP_MEMBERS   │
│ created_at      │   │       ├─────────────────┤            ├─────────────────┤               ├─────────────────┤
└─────────────────┘   │       │ id (PK)         │            │ id (PK)         │               │ id (PK)         │
        │              │       │ post_id (FK)    │            │ sender_id (FK)  │───────────────│ group_id (FK)   │
        │              │       │ user_id (FK)    │────────────│ recipient_id(FK)│               │ user_id (FK)    │
        │              └───────│ created_at      │            │ content         │               │ role            │
        │                      └─────────────────┘            │ created_at      │               │ status          │
        │                                                     │ is_read         │               │ joined_at       │
        │                                                     └─────────────────┘               └─────────────────┘
        │                                                             │                                   │
        │              ┌─────────────────┐            ┌─────────────────┐               ┌─────────────────┐
        │              │    EVENTS       │            │ GROUP_MESSAGES  │               │ NOTIFICATIONS   │
        │              ├─────────────────┤            ├─────────────────┤               ├─────────────────┤
        │              │ id (PK)         │            │ id (PK)         │               │ id (PK)         │
        │              │ group_id (FK)   │            │ group_id (FK)   │               │ user_id (FK)    │
        │              │ title           │            │ sender_id (FK)  │               │ type            │
        │              │ description     │            │ content         │               │ related_id      │
        │              │ event_time      │            │ created_at      │               │ content         │
        │              │ creator_id (FK) │            └─────────────────┘               │ created_at      │
        │              │ created_at      │                    │                         │ is_read         │
        │              └─────────────────┘                    │                         └─────────────────┘
        │                      │                             │
        │              ┌─────────────────┐                   │
        └──────────────│ EVENT_RESPONSES │───────────────────┘
                       ├─────────────────┤
                       │ id (PK)         │
                       │ event_id (FK)   │
                       │ user_id (FK)    │
                       │ response        │
                       │ created_at      │
                       └─────────────────┘
```

---

## API Architecture

### Core Services

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  AUTH SERVICE   │    │  USER SERVICE   │    │  POST SERVICE   │    │ GROUP SERVICE   │    │ NOTIFICATION    │
│                 │    │                 │    │                 │    │                 │    │   SERVICE       │
│ POST /register  │    │ GET /profile    │    │ POST /posts     │    │ POST /groups    │    │ GET /notifications│
│ POST /login     │    │ PUT /profile    │    │ GET /posts      │    │ GET /groups     │    │ PUT /notifications│
│ POST /logout    │    │ POST /follow    │    │ GET /posts/:id  │    │ GET /groups/:id │    │ POST /notify    │
│ GET /session    │    │ DELETE /follow  │    │ PUT /posts/:id  │    │ PUT /groups/:id │    │ WebSocket       │
│ Middleware:     │    │ GET /followers  │    │ DELETE /posts/:id│    │ POST /invite    │    │ /ws/notifications│
│ - Session       │    │ GET /following  │    │ POST /comments  │    │ POST /request   │    │                 │
│ - CORS          │    │ GET /search     │    │ GET /comments   │    │ POST /events    │    │                 │
│ - Rate Limit    │    │                 │    │ File Upload     │    │ GET /events     │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Chat Service

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                      CHAT SERVICE                                                              │
│                                                                                                                 │
│  WebSocket Endpoints:                              │  HTTP Endpoints:                                          │
│  • /ws/chat/:userId                               │  • GET /messages/:conversationId                          │
│  • /ws/group-chat/:groupId                        │  • POST /messages                                         │
│                                                   │  • GET /conversations                                     │
│  Real-time Features:                              │  • PUT /messages/:id/read                                 │
│  • Direct messaging                               │                                                            │
│  • Group chat                                     │  Features:                                                │
│  • Typing indicators                              │  • Message persistence                                   │
│  • Online status                                  │  • Emoji support                                         │
│  • Message delivery status                        │  • File sharing                                          │
└─────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Development Workflow

### Week 1: Project Setup

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Day 1-2         │───▶│ Day 3-4         │───▶│ Day 5-6         │───▶│ Day 7           │
│ • Team Meeting  │    │ • Docker Setup  │    │ • Database      │    │ • Review &      │
│ • Requirements  │    │ • Repository    │    │   Design        │    │   Planning      │
│ • Architecture  │    │ • Basic         │    │ • ER Diagram    │    │ • Next Week     │
│ • Tool Selection│    │   Structure     │    │ • API Specs     │    │   Assignment    │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Week 2-3: Infrastructure

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Database Setup  │───▶│ Auth System     │───▶│ Basic API       │
│ • Migrations    │    │ • Registration  │    │ • Routing       │
│ • Schema        │    │ • Login         │    │ • Middleware    │
│ • Connections   │    │ • Sessions      │    │ • Testing       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Week 4-8: Core Services (Parallel Development)

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ User Service    │    │ Post Service    │    │ Group Service   │    │ Chat Service    │
│ Week 4          │    │ Week 5          │    │ Week 6          │    │ Week 7-8        │
│ • Profile CRUD  │    │ • Post CRUD     │    │ • Group CRUD    │    │ • WebSocket     │
│ • Follow System │    │ • Comments      │    │ • Members       │    │ • Messages      │
│ • Privacy       │    │ • Media Upload  │    │ • Events        │    │ • Notifications │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Week 6-10: Frontend (Overlapping with Backend)

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Component Lib   │───▶│ Auth Pages      │───▶│ Main Features   │───▶│ Real-time UI    │
│ Week 6          │    │ Week 7          │    │ Week 8-9        │    │ Week 10         │
│ • Design System │    │ • Login/Signup  │    │ • Feed          │    │ • Chat Interface│
│ • Layout        │    │ • Profile       │    │ • Groups        │    │ • Notifications │
│ • Navigation    │    │ • Basic UI      │    │ • Responsive    │    │ • Polish        │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Week 11-12: Integration & Testing

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Integration     │───▶│ Testing         │───▶│ Deployment      │
│ • API Connect   │    │ • Unit Tests    │    │ • Production    │
│ • WebSocket     │    │ • Integration   │    │ • Monitoring    │
│ • Bug Fixes     │    │ • E2E Testing   │    │ • Documentation │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

---

## Team Organization

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ FRONTEND TEAM   │    │ BACKEND TEAM    │    │ DATABASE TEAM   │    │ DEVOPS TEAM     │
│ (2-3 people)    │    │ (2-3 people)    │    │ (1-2 people)    │    │ (1-2 people)    │
├─────────────────┤    ├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│ • UI/UX Design  │    │ • API Development│    │ • Schema Design │    │ • Docker Setup  │
│ • Component Dev │    │ • Business Logic│    │ • Migrations    │    │ • CI/CD Pipeline│
│ • State Mgmt    │    │ • WebSocket     │    │ • Optimization  │    │ • Testing       │
│ • Integration   │    │ • Authentication│    │ • Backup        │    │ • Deployment    │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
```

---

## Docker Structure

```
project-root/
├── docker-compose.yml
├── frontend/
│   ├── Dockerfile
│   ├── package.json
│   ├── next.config.js
│   └── src/
├── backend/
│   ├── Dockerfile
│   ├── go.mod
│   ├── main.go
│   └── pkg/
│       ├── auth/
│       ├── users/
│       ├── posts/
│       ├── groups/
│       ├── chat/
│       └── db/
│           └── migrations/
│               └── sqlite/
├── websocket/
│   ├── Dockerfile
│   └── main.go
└── database/
    └── init.sql
```

---

## Testing Strategy

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   UNIT TESTS    │    │INTEGRATION TESTS│    │   E2E TESTS     │    │PERFORMANCE TESTS│
├─────────────────┤    ├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│ • Auth Functions│    │ • API Endpoints │    │ • User Flows    │    │ • Load Testing  │
│ • Business Logic│    │ • Database Ops  │    │ • UI Interactions│    │ • Memory Usage  │
│ • Utilities     │    │ • WebSocket     │    │ • Cross-browser │    │ • Database      │
│ • Components    │    │ • File Upload   │    │ • Mobile        │    │ • Scalability   │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
```

---

## Deployment Pipeline

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   COMMIT    │───▶│    BUILD    │───▶│    TEST     │───▶│   DEPLOY    │───▶│   MONITOR   │───▶│  MAINTAIN   │
├─────────────┤    ├─────────────┤    ├─────────────┤    ├─────────────┤    ├─────────────┤    ├─────────────┤
│ • Git Push  │    │ • Docker    │    │ • Unit      │    │ • Staging   │    │ • Logs      │    │ • Updates   │
│ • PR Review │    │   Build     │    │ • Integration│    │ • Production│    │ • Metrics   │    │ • Bug Fixes │
│ • Code      │    │ • Dependencies│    │ • E2E       │    │ • Rollback  │    │ • Alerts    │    │ • Features  │
│   Quality   │    │ • Compile   │    │ • Security  │    │   Ready     │    │ • Health    │    │ • Security  │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```