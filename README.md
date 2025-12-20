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
- SQLite database
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

2. **Run migrations** (first time only)
   ```powershell
   # Windows
   cd db
   .\migrate.ps1
   
   # Linux/Mac
   ./migrate.sh
   ```

3. **Start the application**
   ```bash
   docker-compose up --build
   ```

4. **Access the application**
   - Frontend: http://localhost:3000
   
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
│   └── notifications/
├── db/                 # Database migrations
└── docker-compose.yml  # Docker orchestration
```

## Stopping the Application

```bash
docker-compose down
```

To remove volumes (database data):
```bash
docker-compose down -v
```
