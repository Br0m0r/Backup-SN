# Notification Service - Complete Guide

**A comprehensive walkthrough of the WebSocket-based real-time notification system**

---

## Table of Contents

1. [Overview - The Big Picture](#overview---the-big-picture)
2. [Why Do We Need a Notification Service?](#why-do-we-need-a-notification-service)
3. [The NotificationHub - Central Broadcaster](#the-notificationhub---central-broadcaster)
4. [The NotificationClient - Individual Connections](#the-notificationclient---individual-connections)
5. [Notification Types - What Can Be Notified?](#notification-types---what-can-be-notified)
6. [How Other Services Create Notifications](#how-other-services-create-notifications)
7. [The Complete Flow - User Gets Notified](#the-complete-flow---user-gets-notified)
8. [REST API vs WebSocket - Two Ways to Get Notifications](#rest-api-vs-websocket---two-ways-to-get-notifications)
9. [Database Schema and Queries](#database-schema-and-queries)
10. [Frontend Integration](#frontend-integration)
11. [Service-to-Service Communication](#service-to-service-communication)
12. [Security Considerations](#security-considerations)

---

## Overview - The Big Picture

### What Does This Service Do?

The notification service alerts users about important events in real-time:
- Someone followed you
- You got a group invitation
- New message arrived
- Someone commented on your post
- Event reminders

### Architecture at a Glance

```
┌─────────────────────────────────────────────────────────────────┐
│                      Notification Ecosystem                      │
└─────────────────────────────────────────────────────────────────┘

   User Service        Groups Service       Chat Service
   (port 8082)         (port 8084)         (port 8085)
        ↓                   ↓                   ↓
        └───────────────────┴───────────────────┘
                            ↓
                  ┌──────────────────────┐
                  │ Notification Service │  ← Creates notifications
                  │    (port 8086)       │
                  └──────────────────────┘
                            ↓
                  ┌──────────────────────┐
                  │   NotificationHub    │  ← Manages WebSocket clients
                  │  [Connected Users]   │
                  └──────────────────────┘
                            ↓
                    ┌───────┴────────┐
                    ↓                ↓
            Browser (User A)    Browser (User B)
            🔔 3 unread         🔔 1 unread
```

### Key Components

1. **NotificationHub**: Manages all active WebSocket connections
2. **NotificationClient**: Represents each user's WebSocket connection
3. **REST API**: Allows other services to create notifications
4. **Database**: Stores notification history
5. **WebSocket**: Pushes real-time notifications to connected users

[Back to Top](#table-of-contents)

---

## Why Do We Need a Notification Service?

### The Problem Without It

Imagine this scenario:
```
User A follows User B
     ↓
User B has to refresh the page to see they have a new follower
     ↓
Bad user experience! 😞
```

### The Solution

```
User A follows User B
     ↓
User Service calls Notification Service
     ↓
Notification Service:
  1. Saves notification to database
  2. Checks if User B is online (connected via WebSocket)
  3. If online: Instantly pushes notification via WebSocket
  4. If offline: User B will see it when they next login
     ↓
User B sees notification bell 🔔 light up INSTANTLY
     ↓
Great user experience! 😊
```

### Centralized Notification Logic

Instead of every service implementing its own notification system:

**❌ Without Notification Service:**
```
User Service → Has notification code
Groups Service → Has duplicate notification code
Posts Service → Has duplicate notification code
Chat Service → Has duplicate notification code
```
Result: Code duplication, inconsistent notifications

**✅ With Notification Service:**
```
User Service ────┐
Groups Service ──┼─→ Notification Service (single source of truth)
Posts Service ───┤
Chat Service ────┘
```
Result: Clean, consistent, maintainable

[Back to Top](#table-of-contents)

---

## The NotificationHub - Central Broadcaster

### What is the Hub?

Think of the Hub as a **switchboard operator** in an old telephone exchange. It knows:
- Who is currently online (connected clients)
- Where to route notifications
- How to broadcast messages

### Hub Structure

```go
type NotificationHub struct {
    clients    map[int]*NotificationClient  // userID → Client connection
    broadcast  chan *models.Notification    // Channel for notifications to send
    register   chan *NotificationClient     // New users connecting
    unregister chan *NotificationClient     // Users disconnecting
    mu         sync.RWMutex                 // Thread-safe access to clients map
    database   *sql.DB                      // Database connection
}
```

Let's break this down:

#### 1. **clients** - The Connection Registry
```go
clients map[int]*NotificationClient
```
This is like a phonebook:
```
User ID | WebSocket Connection
--------|----------------------
   1    | *NotificationClient (John's connection)
   2    | *NotificationClient (Sarah's connection)
   5    | *NotificationClient (Tom's connection)
```

When a notification arrives for User 2, the Hub looks up `clients[2]` and sends it through Sarah's WebSocket.

#### 2. **broadcast** - The Notification Queue
```go
broadcast chan *models.Notification
```
This is a **channel** (Go's way of sending data between goroutines):
```
[Notification 1] → [Notification 2] → [Notification 3] → ...
```
Think of it as a conveyor belt carrying notifications to be delivered.

#### 3. **register** and **unregister** - Connection Management
```go
register   chan *NotificationClient
unregister chan *NotificationClient
```
These channels handle users connecting/disconnecting:
- **register**: "Hey Hub, User 5 just connected!"
- **unregister**: "Hey Hub, User 2 just disconnected!"

#### 4. **mu** - The Traffic Cop
```go
mu sync.RWMutex
```
This ensures **thread-safety**. Multiple goroutines access `clients` simultaneously:
- Reading: "Is User 5 online?"
- Writing: "Add User 3 to connections"

The mutex (mutual exclusion lock) prevents race conditions.

### The Hub's Main Loop

```go
func (h *NotificationHub) Run() {
    for {
        select {
        case client := <-h.register:
            // New user connected
            h.mu.Lock()
            h.clients[client.userID] = client
            h.mu.Unlock()
            
        case client := <-h.unregister:
            // User disconnected
            h.mu.Lock()
            delete(h.clients, client.userID)
            close(client.send)
            h.mu.Unlock()
            
        case notification := <-h.broadcast:
            // New notification to deliver
            h.mu.RLock()
            if client, ok := h.clients[notification.UserID]; ok {
                // User is online, send it!
                client.send <- marshalled_notification
            }
            h.mu.RUnlock()
        }
    }
}
```

**What's happening here?**

This is an **infinite loop** using Go's `select` statement (like a switch for channels):

1. **When someone connects** (`client := <-h.register`):
   - Lock the map (thread-safety)
   - Add them to `clients`
   - Unlock the map
   - Log it: "Client registered: user 5"

2. **When someone disconnects** (`client := <-h.unregister`):
   - Lock the map
   - Remove them from `clients`
   - Close their send channel (no more messages)
   - Unlock the map

3. **When a notification arrives** (`notification := <-h.broadcast`):
   - Lock the map for reading (RLock - allows multiple readers)
   - Check if user is online: `clients[notification.UserID]`
   - If online, send notification through their WebSocket
   - Unlock the map

### Visualizing the Hub

```
┌─────────────────────────────────────────────────────────┐
│                   NotificationHub                        │
│                                                          │
│  Clients Map:                                            │
│  ┌─────────────────────────────────────┐                │
│  │ UserID | WebSocket Connection       │                │
│  ├─────────────────────────────────────┤                │
│  │   1    │ Client{conn, send chan}    │                │
│  │   3    │ Client{conn, send chan}    │                │
│  │   7    │ Client{conn, send chan}    │                │
│  └─────────────────────────────────────┘                │
│                                                          │
│  Channels:                                               │
│  register   ──►  [Client 3] ──► Add to map               │
│  unregister ──►  [Client 1] ──► Remove from map          │
│  broadcast  ──►  [Notif for User 3] ──► Send via WS     │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

[Back to Top](#table-of-contents)

---

## The NotificationClient - Individual Connections

### What is a Client?

A **NotificationClient** represents **one user's WebSocket connection** to the notification service.

```go
type NotificationClient struct {
    hub    *NotificationHub        // Reference back to the Hub
    conn   *websocket.Conn         // The actual WebSocket connection
    send   chan []byte             // Channel for outgoing notifications
    userID int                     // Which user this connection belongs to
}
```

### Client Components

#### 1. **hub** - Link to the Hub
```go
hub *NotificationHub
```
The client needs to know about the Hub so it can:
- Unregister itself when disconnecting
- Access the database if needed

#### 2. **conn** - The WebSocket Connection
```go
conn *websocket.Conn
```
This is the actual TCP connection to the user's browser. Through this:
- Server **sends** notifications to browser
- Browser **receives** notifications (mostly one-way communication)

#### 3. **send** - The Outgoing Queue
```go
send chan []byte
```
This channel holds notifications waiting to be sent:
```
[Notification bytes 1] → [Notification bytes 2] → ...
```

Why a channel instead of sending directly?
- **Non-blocking**: If the connection is slow, we don't block
- **Buffered**: Can queue up multiple notifications
- **Thread-safe**: Multiple goroutines can push to it safely

#### 4. **userID** - Identity
```go
userID int
```
Simple: who is this connection for?

### Client's Two Goroutines

Each client runs **two concurrent goroutines**:

#### **readPump** - Listening for Pings
```go
func (c *NotificationClient) readPump() {
    defer func() {
        c.hub.unregister <- c
        c.conn.Close()
    }()
    
    c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
        return nil
    })
    
    for {
        _, _, err := c.conn.ReadMessage()
        if err != nil {
            break  // Connection died
        }
    }
}
```

**What's this doing?**

Unlike chat (which is bidirectional), notifications are mostly **one-way** (server → client). But the client still needs to:

1. **Keep connection alive**: WebSocket sends "ping" messages every 54 seconds, client responds with "pong"
2. **Detect disconnections**: If `ReadMessage()` returns error, connection is dead
3. **Clean up**: When loop exits, unregister from Hub and close connection

Think of it as a **heartbeat monitor** 💓

#### **writePump** - Sending Notifications
```go
func (c *NotificationClient) writePump() {
    ticker := time.NewTicker(54 * time.Second)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()
    
    for {
        select {
        case message, ok := <-c.send:
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            
            // Send the notification
            w, _ := c.conn.NextWriter(websocket.TextMessage)
            w.Write(message)
            
            // Batch: if more notifications waiting, send them too
            n := len(c.send)
            for i := 0; i < n; i++ {
                w.Write([]byte{'\n'})
                w.Write(<-c.send)
            }
            
            w.Close()
            
        case <-ticker.C:
            // Send ping to keep connection alive
            c.conn.WriteMessage(websocket.PingMessage, nil)
        }
    }
}
```

**What's this doing?**

1. **Waits for notifications**: `<-c.send` blocks until notification arrives
2. **Sends the notification**: Writes JSON to WebSocket
3. **Batch optimization**: If multiple notifications queued, send them all at once
4. **Sends pings**: Every 54 seconds, sends ping to keep connection alive

### Client Lifecycle

```
1. User logs in
   ↓
2. Browser: new WebSocket('ws://localhost:8086/ws?token=xxx')
   ↓
3. Server: Upgrade HTTP → WebSocket
   ↓
4. Create NotificationClient{userID: 5, conn: websocket, ...}
   ↓
5. Register with Hub: hub.register <- client
   ↓
6. Start goroutines:
   - go client.readPump()   (listen for pings/disconnects)
   - go client.writePump()  (send notifications)
   ↓
7. [Client is now active and can receive notifications]
   ↓
8. Notification arrives:
   Hub → client.send <- notification_bytes
   ↓
9. writePump sends it through WebSocket
   ↓
10. Browser receives and displays notification
    ↓
11. User logs out or closes browser
    ↓
12. readPump detects disconnect
    ↓
13. Unregister: hub.unregister <- client
    ↓
14. Hub removes from clients map
    ↓
15. Both goroutines exit, resources freed
```

[Back to Top](#table-of-contents)

---

## Notification Types - What Can Be Notified?

### The 8 Notification Types

```go
const (
    TypeFollow        = "follow"          // Someone followed you
    TypeFollowRequest = "follow_request"  // Follow request received
    TypeGroupInvite   = "group_invite"    // Invited to group
    TypeGroupRequest  = "group_request"   // Join request for your group
    TypeEvent         = "event"           // Event created/updated
    TypeMessage       = "message"         // New message
    TypeComment       = "comment"         // Comment on your post
    TypePost          = "post"            // New post from followed user
)
```

### Notification Data Structure

```go
type Notification struct {
    ID        int       `json:"id"`          // Unique notification ID
    UserID    int       `json:"user_id"`     // Who receives this notification
    Type      string    `json:"type"`        // One of the 8 types above
    RelatedID int       `json:"related_id"`  // ID of the related entity
    Content   string    `json:"content"`     // Human-readable message
    IsRead    bool      `json:"is_read"`     // Has user seen it?
    CreatedAt time.Time `json:"created_at"`  // When it was created
}
```

### RelatedID - The Context Pointer

**RelatedID** points to the relevant entity:

| Type | RelatedID Points To | Example |
|------|---------------------|---------|
| `follow` | Follower's user ID | User 5 followed you → `related_id: 5` |
| `follow_request` | Requester's user ID | User 3 wants to follow → `related_id: 3` |
| `group_invite` | Group ID | Invited to group 7 → `related_id: 7` |
| `group_request` | Group ID | Join request for group 2 → `related_id: 2` |
| `event` | Event ID | Event 12 created → `related_id: 12` |
| `message` | Message/Conversation ID | New message 45 → `related_id: 45` |
| `comment` | Comment ID | Comment 89 on post → `related_id: 89` |
| `post` | Post ID | New post 34 → `related_id: 34` |

**Why is this useful?**

When user clicks notification, frontend can:
```javascript
if (notification.type === 'follow_request') {
    // Navigate to /users/requests
    showFollowRequest(notification.related_id);
}
else if (notification.type === 'group_invite') {
    // Navigate to /groups/123
    navigateToGroup(notification.related_id);
}
else if (notification.type === 'message') {
    // Open chat with sender
    openChat(notification.related_id);
}
```

### Content - The Human Message

**Content** is what users actually see:

```javascript
{
    "type": "follow_request",
    "content": "Tom sent you a follow request"  // ← User sees this
}

{
    "type": "group_invite",
    "content": "Sarah invited you to join Gaming Club"
}

{
    "type": "comment",
    "content": "Mike commented on your post: 'Great photo!'"
}
```

[Back to Top](#table-of-contents)

---

## How Other Services Create Notifications

### Service-to-Service Communication

When something happens in another service, it calls the Notification Service via HTTP:

```
┌─────────────────┐
│  User Service   │  User A follows User B
│  (port 8082)    │
└────────┬────────┘
         │ HTTP POST
         ↓
┌─────────────────────────┐
│  Notification Service   │
│      (port 8086)        │
│  POST /notifications    │
└─────────────────────────┘
```

### Example: User Service Creates Follow Request Notification

```go
// In User Service: services/users/handlers/follow.go

func (h *UserHandlers) SendFollowRequest(w http.ResponseWriter, r *http.Request) {
    // ... validate request, save follow request to database ...
    
    // Create notification
    notificationData := map[string]interface{}{
        "user_id":    targetUserID,                    // User B (receiver)
        "type":       "follow_request",
        "related_id": followerID,                      // User A (sender)
        "content":    fmt.Sprintf("%s sent you a follow request", followerName),
    }
    
    // Call Notification Service
    jsonData, _ := json.Marshal(notificationData)
    resp, err := http.Post(
        "http://notification-service:8086/notifications",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    
    if err != nil {
        log.Printf("Failed to create notification: %v", err)
        // Don't fail the request, just log the error
    }
    
    // ... return success response ...
}
```

### What Happens Inside Notification Service?

```go
// In Notification Service: handlers/notification.go

func (h *NotificationHandlers) CreateNotification(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request body
    var req CreateNotificationRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // 2. Validate notification type
    if !validTypes[req.Type] {
        return error
    }
    
    // 3. Save to database
    notification := db.CreateNotification(h.database, &req)
    
    // 4. Broadcast to WebSocket (if user is online)
    h.hub.BroadcastNotification(notification)
    
    // 5. Return success
    return notification
}
```

### The Broadcast Flow

```go
func (h *NotificationHub) BroadcastNotification(notification *Notification) {
    h.broadcast <- notification  // Push to Hub's broadcast channel
}
```

Then in Hub's Run() loop:
```go
case notification := <-h.broadcast:
    h.mu.RLock()
    if client, ok := h.clients[notification.UserID]; ok {
        // User is online! Send it via WebSocket
        data, _ := json.Marshal(notification)
        client.send <- data
    } else {
        // User is offline, notification is already in database
        // They'll see it when they next login
    }
    h.mu.RUnlock()
```

### Complete Flow Diagram

```
User A Follows User B
     ↓
┌────────────────────────────────────────────────────────────┐
│ User Service: Save follow request to database              │
└────────────────────────────────────────────────────────────┘
     ↓
     HTTP POST to http://notification-service:8086/notifications
     {
         "user_id": 2,           // User B
         "type": "follow_request",
         "related_id": 1,        // User A
         "content": "User A sent you a follow request"
     }
     ↓
┌────────────────────────────────────────────────────────────┐
│ Notification Service:                                       │
│ 1. Validate request ✓                                       │
│ 2. Save to notifications table                              │
│ 3. Check if User B is online (connected via WebSocket)     │
│ 4. If online: Send via WebSocket → Browser shows 🔔        │
│ 5. If offline: User B will see it when they login          │
└────────────────────────────────────────────────────────────┘
```

[Back to Top](#table-of-contents)

---

## The Complete Flow - User Gets Notified

### Scenario: Sarah Invites Tom to Join "Gaming Club"

Let's trace the **complete end-to-end flow**:

#### Step 1: Tom Logs In
```
1. Browser: POST /login (username, password)
   ↓
2. Auth Service: Validates, creates session token
   ↓
3. Browser receives token, stores in localStorage
   ↓
4. Browser: new WebSocket('ws://localhost:8086/ws?token=xxx')
   ↓
5. Notification Service:
   - Validates token with Auth Service
   - Creates NotificationClient{userID: 5}  // Tom is User 5
   - Registers: hub.register <- client
   ↓
6. Tom's browser is now connected and listening
```

#### Step 2: Sarah Invites Tom to Group
```
1. Sarah (User 3) clicks "Invite Tom" in Gaming Club (Group 7)
   ↓
2. Browser: POST /groups/7/invite {user_id: 5}
   ↓
3. Groups Service:
   - Validates Sarah has permission
   - Saves invitation to database
   - Calls Notification Service ↓
```

#### Step 3: Notification Service Creates Notification
```
Groups Service → POST http://notification-service:8086/notifications
Body: {
    "user_id": 5,                                           // Tom
    "type": "group_invite",
    "related_id": 7,                                        // Gaming Club
    "content": "Sarah invited you to join Gaming Club"
}
     ↓
Notification Service Handler:
1. Validates type "group_invite" ✓
2. Inserts into database:
   INSERT INTO notifications (user_id, type, related_id, content)
   VALUES (5, 'group_invite', 7, 'Sarah invited you...')
   
3. Gets created notification with ID (e.g., ID: 123)
4. Calls: h.hub.BroadcastNotification(notification)
```

#### Step 4: Hub Broadcasts to Tom
```
Hub.BroadcastNotification(notification):
     ↓
h.broadcast <- notification  // Push to broadcast channel
     ↓
Hub's Run() loop receives it:
     ↓
case notification := <-h.broadcast:
    h.mu.RLock()
    client, ok := h.clients[5]  // Look up Tom's connection
    if ok {
        // Tom is online!
        data := json.Marshal(notification)  // Convert to JSON bytes
        client.send <- data                  // Push to Tom's send channel
    }
    h.mu.RUnlock()
```

#### Step 5: Client Sends to Tom's Browser
```
Tom's NotificationClient.writePump():
     ↓
case message := <-c.send:  // Receives notification bytes
     ↓
w, _ := c.conn.NextWriter(websocket.TextMessage)
w.Write(message)  // Send through WebSocket
w.Close()
     ↓
Tom's Browser receives:
{
    "type": "notification",
    "notification": {
        "id": 123,
        "user_id": 5,
        "type": "group_invite",
        "related_id": 7,
        "content": "Sarah invited you to join Gaming Club",
        "is_read": false,
        "created_at": "2025-10-17T10:30:00Z"
    }
}
```

#### Step 6: Browser Updates UI
```javascript
notificationWS.onmessage = (event) => {
    const data = JSON.parse(event.data);
    const notification = data.notification;
    
    // Update notification count
    updateNotificationBadge();  // 🔔 3 → 🔔 4
    
    // Add to notification dropdown
    addNotificationToList(notification);
    
    // Show browser notification if permitted
    if (Notification.permission === "granted") {
        new Notification("Sarah invited you to join Gaming Club");
    }
    
    // Play sound (optional)
    notificationSound.play();
};
```

### Timeline Visualization

```
Time: 0ms
├─ Sarah clicks "Invite Tom"
│
Time: 10ms
├─ Groups Service validates & saves to DB
│
Time: 20ms
├─ Groups Service → POST to Notification Service
│
Time: 25ms
├─ Notification Service saves to database
│  INSERT INTO notifications...
│
Time: 27ms
├─ Notification Service broadcasts to Hub
│  hub.broadcast <- notification
│
Time: 28ms
├─ Hub checks if Tom is online
│  client, ok := clients[5]  ✓ Tom is connected!
│
Time: 29ms
├─ Hub pushes to Tom's client send channel
│  client.send <- notification_data
│
Time: 30ms
├─ Client's writePump sends via WebSocket
│  conn.Write(notification_data)
│
Time: 35ms (network latency)
├─ Tom's browser receives notification
│
Time: 36ms
└─ Browser updates UI: 🔔 badge lights up!

Total: 36ms from action to notification!
```

### What if Tom Was Offline?

```
Step 4 (modified):
     ↓
Hub.BroadcastNotification(notification):
     ↓
case notification := <-h.broadcast:
    h.mu.RLock()
    client, ok := h.clients[5]  // Look up Tom
    if !ok {
        // Tom is NOT online, skip WebSocket
        // Notification is already in database though!
    }
    h.mu.RUnlock()
     ↓
Later when Tom logs in:
     ↓
Browser: GET /notifications/list?unread=true
     ↓
Notification Service returns all unread notifications
     ↓
Browser displays: "🔔 You have 4 unread notifications"
```

[Back to Top](#table-of-contents)

---

## REST API vs WebSocket - Two Ways to Get Notifications

### Why Both?

| Method | Use Case | When |
|--------|----------|------|
| **WebSocket** | Real-time push | User is actively browsing |
| **REST API** | Fetch history | User just logged in, or polling |

### REST API Endpoints

#### 1. Get All Notifications
```http
GET /notifications/list?limit=20&offset=0&unread=false
Authorization: Bearer <token>
```

**Use case**: Display notification history page

**Response**:
```json
{
    "success": true,
    "data": {
        "notifications": [
            {
                "id": 123,
                "user_id": 5,
                "type": "group_invite",
                "related_id": 7,
                "content": "Sarah invited you to join Gaming Club",
                "is_read": false,
                "created_at": "2025-10-17T10:30:00Z"
            },
            {
                "id": 122,
                "type": "follow_request",
                "content": "Tom sent you a follow request",
                "is_read": true,
                "created_at": "2025-10-17T09:15:00Z"
            }
        ],
        "count": 2
    }
}
```

#### 2. Get Unread Count
```http
GET /notifications/unread-count
Authorization: Bearer <token>
```

**Use case**: Show badge on notification bell

**Response**:
```json
{
    "success": true,
    "data": {
        "unread_count": 5
    }
}
```

#### 3. Mark as Read
```http
PUT /notifications/read/123
Authorization: Bearer <token>
```

**Use case**: User clicks notification, mark it as read

#### 4. Mark All as Read
```http
POST /notifications/read-all
Authorization: Bearer <token>
```

**Use case**: "Mark all as read" button

#### 5. Delete Notification
```http
DELETE /notifications/delete/123
Authorization: Bearer <token>
```

**Use case**: User dismisses notification

### WebSocket Connection

```javascript
const ws = new WebSocket('ws://localhost:8086/ws?token=' + sessionToken);

ws.onopen = () => {
    console.log('Connected to notifications');
};

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    if (data.type === 'notification') {
        handleNewNotification(data.notification);
    }
};

ws.onclose = () => {
    console.log('Disconnected from notifications');
    // Implement reconnection logic
};
```

### Hybrid Approach (Best Practice)

```javascript
// On page load
async function initializeNotifications() {
    // 1. Fetch existing notifications via REST
    const response = await fetch('/notifications/list?unread=true', {
        headers: { 'Authorization': 'Bearer ' + token }
    });
    const data = await response.json();
    displayNotifications(data.notifications);
    updateBadge(data.count);
    
    // 2. Connect WebSocket for real-time updates
    connectNotificationWebSocket();
}

// WebSocket receives new notification
function handleNewNotification(notification) {
    // Add to existing list
    prependNotification(notification);
    
    // Update badge count
    incrementBadge();
    
    // Show browser notification
    showBrowserNotification(notification.content);
}
```

[Back to Top](#table-of-contents)

---

## Database Schema and Queries

### Notifications Table

```sql
CREATE TABLE notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,                -- Recipient
    type TEXT CHECK (type IN (                -- Notification type
        'follow', 'follow_request', 
        'group_invite', 'group_request', 
        'event', 'message', 'comment', 'post'
    )) NOT NULL,
    related_id INTEGER,                       -- Context pointer
    content TEXT NOT NULL,                    -- Human-readable message
    is_read BOOLEAN NOT NULL DEFAULT 0,       -- Read status
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Indexes for fast queries
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_is_read ON notifications(is_read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);
```

### Key Queries

#### Create Notification
```go
func CreateNotification(db *sql.DB, req *CreateNotificationRequest) (*Notification, error) {
    query := `
        INSERT INTO notifications (user_id, type, related_id, content)
        VALUES (?, ?, ?, ?)
    `
    result, _ := db.Exec(query, req.UserID, req.Type, req.RelatedID, req.Content)
    id, _ := result.LastInsertId()
    return GetNotificationByID(db, int(id))
}
```

#### Get User Notifications (with Pagination)
```go
func GetUserNotifications(db *sql.DB, userID, limit, offset int) ([]Notification, error) {
    query := `
        SELECT id, user_id, type, related_id, content, is_read, created_at
        FROM notifications
        WHERE user_id = ?
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?
    `
    rows, _ := db.Query(query, userID, limit, offset)
    // ... scan rows into notifications slice
}
```

#### Get Unread Notifications
```go
func GetUnreadNotifications(db *sql.DB, userID int) ([]Notification, error) {
    query := `
        SELECT id, user_id, type, related_id, content, is_read, created_at
        FROM notifications
        WHERE user_id = ? AND is_read = 0
        ORDER BY created_at DESC
    `
    rows, _ := db.Query(query, userID)
    // ... scan rows
}
```

#### Get Unread Count
```go
func GetUnreadCount(db *sql.DB, userID int) (int, error) {
    query := `SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = 0`
    var count int
    db.QueryRow(query, userID).Scan(&count)
    return count, nil
}
```

#### Mark as Read
```go
func MarkAsRead(db *sql.DB, notificationID, userID int) error {
    query := `UPDATE notifications SET is_read = 1 WHERE id = ? AND user_id = ?`
    _, err := db.Exec(query, notificationID, userID)
    return err
}
```

#### Mark All as Read
```go
func MarkAllAsRead(db *sql.DB, userID int) error {
    query := `UPDATE notifications SET is_read = 1 WHERE user_id = ? AND is_read = 0`
    _, err := db.Exec(query, userID)
    return err
}
```

[Back to Top](#table-of-contents)

---

## Frontend Integration

### Complete Frontend Module

```javascript
// frontend/js/notifications.js

const NOTIFICATION_URL = 'http://localhost:8086';
const NOTIFICATION_WS_URL = 'ws://localhost:8086/ws';

let notificationWebSocket = null;
let reconnectAttempts = 0;
const MAX_RECONNECT_ATTEMPTS = 5;

// Connect WebSocket
function connectNotificationWebSocket() {
    const token = AppState.getToken();
    if (!token) return;
    
    console.log('Connecting to notification WebSocket...');
    notificationWebSocket = new WebSocket(`${NOTIFICATION_WS_URL}?token=${token}`);
    
    notificationWebSocket.onopen = () => {
        console.log('✅ Notification WebSocket connected');
        reconnectAttempts = 0;
        updateConnectionStatus(true);
    };
    
    notificationWebSocket.onmessage = (event) => {
        const data = JSON.parse(event.data);
        if (data.type === 'notification') {
            handleNewNotification(data.notification);
        }
    };
    
    notificationWebSocket.onerror = (error) => {
        console.error('❌ Notification WebSocket error:', error);
    };
    
    notificationWebSocket.onclose = () => {
        console.log('🔌 Notification WebSocket closed');
        updateConnectionStatus(false);
        
        // Attempt reconnection
        if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
            reconnectAttempts++;
            setTimeout(() => {
                console.log(`Reconnecting... attempt ${reconnectAttempts}/${MAX_RECONNECT_ATTEMPTS}`);
                connectNotificationWebSocket();
            }, 2000 * reconnectAttempts);
        }
    };
}

// Disconnect WebSocket
function disconnectNotificationWebSocket() {
    if (notificationWebSocket) {
        notificationWebSocket.close();
        notificationWebSocket = null;
    }
}

// Handle incoming notification
function handleNewNotification(notification) {
    console.log('New notification:', notification);
    
    // Update badge count
    updateNotificationBadge();
    
    // Add to dropdown list
    prependNotificationToList(notification);
    
    // Show browser notification
    if (Notification.permission === "granted") {
        new Notification(notification.content, {
            icon: '/icon.png',
            tag: 'notification-' + notification.id
        });
    }
    
    // Play sound (optional)
    playNotificationSound();
}

// Fetch all notifications
async function loadNotifications() {
    const response = await Utils.apiCall(
        `${NOTIFICATION_URL}/notifications/list?limit=20`,
        'GET',
        null,
        true
    );
    
    if (response.ok && response.data.success) {
        displayNotifications(response.data.data.notifications);
    }
}

// Get unread count
async function loadUnreadCount() {
    const response = await Utils.apiCall(
        `${NOTIFICATION_URL}/notifications/unread-count`,
        'GET',
        null,
        true
    );
    
    if (response.ok && response.data.success) {
        const count = response.data.data.unread_count;
        updateNotificationBadge(count);
    }
}

// Mark notification as read
async function markAsRead(notificationId) {
    await Utils.apiCall(
        `${NOTIFICATION_URL}/notifications/read/${notificationId}`,
        'PUT',
        null,
        true
    );
    updateNotificationBadge();
}

// Mark all as read
async function markAllAsRead() {
    await Utils.apiCall(
        `${NOTIFICATION_URL}/notifications/read-all`,
        'POST',
        null,
        true
    );
    updateNotificationBadge();
}

// UI Functions
function updateNotificationBadge(count) {
    const badge = document.getElementById('notificationBadge');
    if (count === undefined) {
        loadUnreadCount(); // Refresh count
    } else {
        badge.textContent = count;
        badge.style.display = count > 0 ? 'inline' : 'none';
    }
}

function displayNotifications(notifications) {
    const container = document.getElementById('notificationList');
    container.innerHTML = '';
    
    notifications.forEach(notif => {
        const el = createNotificationElement(notif);
        container.appendChild(el);
    });
}

function createNotificationElement(notif) {
    const div = document.createElement('div');
    div.className = 'notification-item ' + (notif.is_read ? '' : 'unread');
    div.onclick = () => handleNotificationClick(notif);
    
    const icons = {
        'follow': '👥',
        'follow_request': '👤',
        'group_invite': '📨',
        'group_request': '🤝',
        'event': '📅',
        'message': '💬',
        'comment': '💭',
        'post': '📝'
    };
    
    div.innerHTML = `
        <div class="notification-icon">${icons[notif.type] || '🔔'}</div>
        <div class="notification-content">
            <p>${Utils.escapeHtml(notif.content)}</p>
            <span class="notification-time">${formatTime(notif.created_at)}</span>
        </div>
    `;
    
    return div;
}

function handleNotificationClick(notification) {
    // Mark as read
    if (!notification.is_read) {
        markAsRead(notification.id);
    }
    
    // Navigate based on type
    switch (notification.type) {
        case 'follow_request':
            Utils.switchMainTab('users');
            window.switchUserSubTab('requests');
            break;
        case 'group_invite':
            Utils.switchMainTab('groups');
            Groups.viewGroup(notification.related_id);
            break;
        case 'message':
            Utils.switchMainTab('chat');
            Chat.openPrivateChat(notification.related_id);
            break;
        // ... handle other types
    }
}

// Listen for login/logout events
document.addEventListener('userLoggedIn', () => {
    connectNotificationWebSocket();
    loadUnreadCount();
});

document.addEventListener('userLoggedOut', () => {
    disconnectNotificationWebSocket();
});

// Export
window.Notifications = {
    connectNotificationWebSocket,
    disconnectNotificationWebSocket,
    loadNotifications,
    loadUnreadCount,
    markAsRead,
    markAllAsRead
};
```

[Back to Top](#table-of-contents)

---

## Service-to-Service Communication

### How Services Call Notification Service

Every service that creates notifications follows this pattern:

```go
// Helper function to create notification
func createNotification(userID int, notifType, content string, relatedID int) {
    notificationData := map[string]interface{}{
        "user_id":    userID,
        "type":       notifType,
        "related_id": relatedID,
        "content":    content,
    }
    
    jsonData, err := json.Marshal(notificationData)
    if err != nil {
        log.Printf("Failed to marshal notification: %v", err)
        return
    }
    
    notificationServiceURL := os.Getenv("NOTIFICATION_SERVICE_URL")
    if notificationServiceURL == "" {
        notificationServiceURL = "http://notification-service:8086"
    }
    
    resp, err := http.Post(
        notificationServiceURL+"/notifications",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    
    if err != nil {
        log.Printf("Failed to create notification: %v", err)
        // Don't fail the main request, just log the error
        return
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        log.Printf("Notification service returned status: %d", resp.StatusCode)
    }
}
```

### Examples from Each Service

#### User Service - Follow Request
```go
// services/users/handlers/follow.go
func (h *UserHandlers) SendFollowRequest(w http.ResponseWriter, r *http.Request) {
    // ... save follow request ...
    
    createNotification(
        targetUserID,
        "follow_request",
        fmt.Sprintf("%s sent you a follow request", requesterName),
        requesterID,
    )
}
```

#### Groups Service - Group Invitation
```go
// services/groups/handlers/group.go
func (h *GroupHandlers) InviteToGroup(w http.ResponseWriter, r *http.Request) {
    // ... save invitation ...
    
    createNotification(
        invitedUserID,
        "group_invite",
        fmt.Sprintf("%s invited you to join %s", inviterName, groupName),
        groupID,
    )
}
```

#### Chat Service - New Message
```go
// services/chat/handlers/chat.go
func (h *ChatHandlers) SaveMessage(msg *Message) {
    // ... save message to database ...
    
    // Only notify if receiver is not in active chat
    if !h.hub.IsUserOnline(msg.ReceiverID) {
        createNotification(
            msg.ReceiverID,
            "message",
            fmt.Sprintf("New message from %s", senderName),
            msg.ID,
        )
    }
}
```

#### Posts Service - New Comment
```go
// services/posts/handlers/comment.go
func (h *PostHandlers) CreateComment(w http.ResponseWriter, r *http.Request) {
    // ... save comment ...
    
    createNotification(
        postAuthorID,
        "comment",
        fmt.Sprintf("%s commented on your post: '%s'", 
            commenterName, 
            truncate(commentContent, 50)),
        commentID,
    )
}
```

[Back to Top](#table-of-contents)

---

## Security Considerations

### 1. Authentication

**WebSocket Authentication:**
- Token passed as query parameter: `?token=xxx`
- Validated with Auth Service on connection
- Connection rejected if invalid

**REST API Authentication:**
- Bearer token in Authorization header
- Middleware validates with Auth Service
- 401 Unauthorized if invalid

### 2. Authorization

**User can only:**
- Read their own notifications (`WHERE user_id = ?`)
- Mark their own notifications as read
- Delete their own notifications

**Enforced by:**
```go
func (h *NotificationHandlers) GetNotifications(w http.ResponseWriter, r *http.Request) {
    userID, _ := middleware.GetUserIDFromContext(r)  // From token
    notifications := db.GetUserNotifications(database, userID, ...)
    // ↑ SQL query filters by userID automatically
}
```

### 3. Service-to-Service Authentication

**Current Implementation:**
- `/notifications` endpoint is **open** (no auth)
- Any service can create notifications

**Production Considerations:**
1. **API Keys**: Each service has a secret key
```go
apiKey := r.Header.Get("X-API-Key")
if apiKey != os.Getenv("SERVICE_API_KEY") {
    return 401 Unauthorized
}
```

2. **Service Accounts**: Services authenticate with JWT
```go
serviceToken := r.Header.Get("X-Service-Token")
// Validate service JWT
```

3. **Internal Network**: Services only accessible within Docker network
```yaml
# docker-compose.yml
services:
  notification-service:
    ports:
      - "8086:8086"  # External access for frontend
    networks:
      - internal     # Services communicate internally
```

### 4. Input Validation

**Notification Type:**
```go
validTypes := map[string]bool{
    "follow": true,
    "follow_request": true,
    // ... only these 8 types allowed
}
if !validTypes[req.Type] {
    return 400 Bad Request
}
```

**Content Sanitization:**
```go
// Limit content length
if len(req.Content) > 500 {
    return 400 Bad Request "Content too long"
}

// In frontend, always escape HTML
Utils.escapeHtml(notification.content)
```

### 5. Rate Limiting

**Prevent notification spam:**
```go
// Example: Max 10 notifications per minute per user
func rateLimitNotifications(userID int) bool {
    count := getNotificationsInLastMinute(userID)
    return count < 10
}
```

### 6. XSS Prevention

**Always escape in frontend:**
```javascript
// ✅ GOOD
div.textContent = notification.content;
// or
div.innerHTML = Utils.escapeHtml(notification.content);

// ❌ BAD
div.innerHTML = notification.content;  // XSS vulnerability!
```

[Back to Top](#table-of-contents)

---

## Summary

### Key Takeaways

1. **NotificationHub** = Central manager for all WebSocket connections
2. **NotificationClient** = Individual user's WebSocket connection
3. **8 Notification Types** = follow, follow_request, group_invite, group_request, event, message, comment, post
4. **Two Delivery Methods**:
   - REST API for fetching history
   - WebSocket for real-time push
5. **Service Communication**: Other services call `POST /notifications` to create notifications
6. **Database Persistence**: All notifications saved, viewable even if user was offline

### Architecture Benefits

✅ **Centralized**: One place for all notification logic
✅ **Real-time**: WebSocket pushes notifications instantly
✅ **Reliable**: Database persistence ensures no lost notifications
✅ **Scalable**: Each service independent, can scale separately
✅ **Maintainable**: Clear separation of concerns

### What You Learned

- WebSocket vs HTTP for real-time communication
- Hub-and-spoke pattern for managing connections
- Go channels and goroutines for concurrency
- Service-to-service HTTP calls
- Database queries with pagination
- Frontend WebSocket integration

---

**Built with Go, Gorilla WebSocket, and SQLite** 🚀
