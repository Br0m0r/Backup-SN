# Chat Service - Complete Guide

**A comprehensive walkthrough of the WebSocket-based real-time chat system**

---

## Table of Contents

1. [Overview - The Big Picture](#overview---the-big-picture)
2. [What is WebSocket?](#what-is-websocket)
3. [The Hub - The Heart of Chat](#the-hub---the-heart-of-chat)
4. [The Client - Individual Connections](#the-client---individual-connections)
5. [Go Channels - Communication Between Goroutines](#go-channels---communication-between-goroutines)
6. [Go Routines - Concurrent Execution](#go-routines---concurrent-execution)
7. [The Complete Flow - User Sends a Message](#the-complete-flow---user-sends-a-message)
8. [Group Chat - Real-time Group Messaging](#group-chat---real-time-group-messaging)
9. [Access Control - Who Can Chat?](#access-control---who-can-chat)
10. [Message Types and Payloads](#message-types-and-payloads)
11. [Database Operations](#database-operations)
12. [HTTP REST Endpoints](#http-rest-endpoints)
13. [Frontend Integration](#frontend-integration)
14. [Error Handling and Recovery](#error-handling-and-recovery)

---

## Overview - The Big Picture

### What Does This Service Do?

The chat service enables **real-time messaging** between users. When User A sends a message, User B receives it **instantly** without refreshing the page.

### Architecture at a Glance

```
Browser (User A)  <--WebSocket-->  Hub (Central Manager)  <--WebSocket-->  Browser (User B)
     |                                    |                                      |
     |                              [Connection Pool]                           |
     |                              [Message Router]                            |
     |                                    |                                      |
     +--------------------[Database: SQLite]----------------------------------+
```

### Key Components

1. **Hub**: Central manager that tracks all connected users
2. **Client**: Represents each user's WebSocket connection
3. **Database**: Stores message history permanently
4. **Channels**: Go's way of communicating between concurrent processes
5. **Goroutines**: Lightweight threads for handling multiple connections

[Back to Top](#table-of-contents)

---

## What is WebSocket?

### The Problem with Regular HTTP

**Normal HTTP** (what you use for regular web pages):
```
Browser: "Hey server, give me the messages" → Server: "Here are the messages"
[1 second passes]
Browser: "Hey server, any new messages?" → Server: "Nope"
[1 second passes]
Browser: "Hey server, any new messages NOW?" → Server: "Nope"
[1 second passes]
Browser: "Hey server, any new messages NOW??" → Server: "Yes! Here's one!"
```

This is called **polling** and it's:
- Wasteful (constant requests)
- Slow (delay between checks)
- Not real-time

### The WebSocket Solution

**WebSocket** keeps a **permanent connection** open:
```
Browser: "Hey server, let's keep this connection open"
Server: "Sure! I'll push messages to you as they arrive"
[User B sends message]
Server: "Hey User A! New message just arrived!" → Browser: [Shows message instantly]
```

Benefits:
- **Real-time**: Messages arrive instantly
- **Efficient**: One connection, no constant polling
- **Bidirectional**: Both sides can send anytime

### How WebSocket Starts

1. **Browser makes HTTP request**: `GET /ws`
2. **Server "upgrades" connection**: Changes HTTP → WebSocket
3. **Connection stays open**: Can now send messages both ways
4. **Connection closes**: When user logs out or closes browser

[Back to Top](#table-of-contents)

---

## The Hub - The Heart of Chat

### What is the Hub?

Think of the Hub as a **telephone switchboard operator**. It:
- Knows who's online (keeps a list)
- Routes messages to the right person
- Handles connections/disconnections

### Hub Structure

```go
type Hub struct {
    clients    map[int]*Client              // userID → Client (who's online?)
    broadcast  chan *models.WebSocketMessage // incoming messages to send
    register   chan *Client                  // new connections
    unregister chan *Client                  // disconnections
    mu         sync.RWMutex                  // lock for thread safety
    database   *sql.DB                       // database connection
}
```

**Fields Explained:**

1. **`clients map[int]*Client`**
   - A dictionary: User ID → Their connection
   - Example: `{5: *Client, 12: *Client, 23: *Client}` means users 5, 12, and 23 are online

2. **`broadcast chan *models.WebSocketMessage`**
   - A **channel** (mailbox) for messages to send out
   - When you send a message, it goes here first

3. **`register chan *Client`**
   - Channel for new connections
   - When someone connects, they send themselves here

4. **`unregister chan *Client`**
   - Channel for disconnections
   - When someone leaves, they send themselves here

5. **`mu sync.RWMutex`**
   - A lock to prevent race conditions
   - Only one goroutine can modify `clients` at a time

6. **`database *sql.DB`**
   - Connection to SQLite database
   - For storing message history

### Hub.Run() - The Main Loop

This function runs **forever** in a goroutine:

```go
func (h *Hub) Run() {
    for {  // Infinite loop
        select {  // Wait for something to happen
        case client := <-h.register:
            // Someone connected!
            h.clients[client.userID] = client
            
        case client := <-h.unregister:
            // Someone disconnected!
            delete(h.clients, client.userID)
            
        case message := <-h.broadcast:
            // New message to send!
            if receiver, online := h.clients[message.ReceiverID]; online {
                receiver.send <- messageData  // Send to receiver
            }
        }
    }
}
```

**How `select` works:**

`select` is like a waiter waiting for orders:
- Waits at multiple channels
- When ANY channel receives something, executes that case
- Goes back to waiting

**Timeline Example:**
```
[Time 0] Hub starts, waiting...
[Time 2] User 5 connects → register channel receives → Add to clients map
[Time 3] Hub waits again...
[Time 8] User 12 connects → register channel receives → Add to clients map
[Time 9] Hub waits again...
[Time 15] User 5 sends message to 12 → broadcast channel receives → Route to User 12
[Time 16] Hub waits again...
```

[Back to Top](#table-of-contents)

---

## The Client - Individual Connections

### What is a Client?

A **Client** represents **one user's WebSocket connection**. Each online user has one Client.

```go
type Client struct {
    hub      *Hub              // Reference to the Hub
    conn     *websocket.Conn   // The actual WebSocket connection
    send     chan []byte       // Outgoing messages (to browser)
    userID   int               // Which user is this?
    username string            // User's name
}
```

**Fields Explained:**

1. **`hub *Hub`**: Pointer back to the Hub (to unregister when done)
2. **`conn *websocket.Conn`**: The WebSocket connection to the browser
3. **`send chan []byte`**: Outgoing message queue (messages to send to this user)
4. **`userID int`**: Database user ID (e.g., 5, 12, 23)
5. **`username string`**: Display name

### Two Goroutines Per Client

Each Client runs **two goroutines simultaneously**:

```
Client
  ├─ readPump()  → Reads messages FROM browser
  └─ writePump() → Sends messages TO browser
```

Why two? Because reading and writing happen **at the same time**:
- User might be typing while receiving a message
- One doesn't block the other

[Back to Top](#table-of-contents)

---

## Go Channels - Communication Between Goroutines

### What is a Channel?

A **channel** is like a **pipe** or **mailbox**:
- One goroutine puts something in
- Another goroutine takes it out
- If empty, the receiver **waits** until something arrives

### Channel Syntax

```go
// Create a channel
ch := make(chan string)

// Send to channel (goroutine 1)
ch <- "Hello"

// Receive from channel (goroutine 2)
message := <-ch  // Blocks until message arrives
```

### Real Example from Chat Service

```go
// Hub has a broadcast channel
broadcast chan *models.WebSocketMessage

// Client sends message to Hub
hub.broadcast <- message  // Put message in mailbox

// Hub receives from channel
message := <-h.broadcast  // Take message from mailbox
```

### Why Channels?

**Problem without channels:**
```go
// WRONG: Direct variable sharing (race condition!)
var sharedMessage string
// Goroutine 1 writes, Goroutine 2 reads → CHAOS!
```

**Solution with channels:**
```go
// RIGHT: Safe communication
ch := make(chan string)
go func() { ch <- "Safe!" }()  // Goroutine 1
message := <-ch                 // Goroutine 2 (safe!)
```

[Back to Top](#table-of-contents)

---

## Go Routines - Concurrent Execution

### What is a Goroutine?

A **goroutine** is a **lightweight thread**. It lets you run functions **concurrently** (at the same time).

### Regular vs Goroutine

**Regular function call:**
```go
doSomething()  // Blocks until done
doSomethingElse()  // Runs AFTER doSomething finishes
```

**With goroutine:**
```go
go doSomething()  // Starts, doesn't wait
doSomethingElse()  // Runs IMMEDIATELY (parallel)
```

### Chat Service Goroutines

When a user connects:
```go
// Create client
client := &Client{...}

// Start TWO goroutines
go client.writePump()  // Goroutine 1: Send messages to browser
go client.readPump()   // Goroutine 2: Read messages from browser

// Main function continues without waiting
```

Both run **simultaneously**:
```
Time: 0s    1s    2s    3s    4s    5s
readPump:  [reading...waiting...reading...waiting...]
writePump: [sending...idle...sending...idle...]
```

### Hub.Run() Goroutine

The Hub runs in its **own goroutine** from the start:
```go
hub := NewHub(database)
go hub.Run()  // Starts infinite loop in background
```

Now the Hub runs **forever** in the background while the main program continues.

[Back to Top](#table-of-contents)

---

## The Complete Flow - User Sends a Message

### 📋 Step-by-Step Timeline

Let's follow what happens when **User A (ID=5)** sends "Hello!" to **User B (ID=12)**.

---

### Step 1: User Opens Browser

```
Browser: Opens http://localhost:3000
Frontend: Loads JavaScript
JavaScript: Creates WebSocket connection
```

**Code in browser:**
```javascript
const ws = new WebSocket('ws://localhost:8085/ws');
```

---

### Step 2: WebSocket Upgrade

**Browser sends HTTP request:**
```http
GET /ws HTTP/1.1
Host: localhost:8085
Upgrade: websocket
Connection: Upgrade
Authorization: Bearer <token>
```

**Server receives in `HandleWebSocket()`:**
```go
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    // 1. Check auth token
    userID := middleware.GetUserIDFromContext(r)  // userID = 5
    
    // 2. Upgrade HTTP → WebSocket
    conn, err := upgrader.Upgrade(w, r, nil)
    
    // 3. Create Client for this user
    client := &Client{
        hub:    h,
        conn:   conn,
        send:   make(chan []byte, 256),
        userID: 5,
    }
    
    // 4. Register with Hub
    h.register <- client  // Put client in register channel
    
    // 5. Start goroutines
    go client.writePump()
    go client.readPump()
}
```

---

### Step 3: Hub Registers User

**Hub's `Run()` receives:**
```go
case client := <-h.register:
    h.mu.Lock()
    h.clients[5] = client  // Add User 5 to online list
    h.mu.Unlock()
    log.Printf("User 5 connected")
```

**Hub's internal state:**
```go
clients = {
    5: *Client{userID: 5, conn: <WebSocket>},
    12: *Client{userID: 12, conn: <WebSocket>},  // User 12 already online
}
```

---

### Step 4: User Types Message

**User A types in browser:**
```javascript
// Frontend JavaScript
sendMessage(12, "Hello!");  // Send to User 12

// Creates WebSocket message
const wsMessage = {
    type: "message",
    receiver_id: 12,
    content: "Hello!",
    timestamp: new Date()
};

// Send via WebSocket
ws.send(JSON.stringify(wsMessage));
```

**Message travels through network to server**

---

### Step 5: readPump() Receives Message

**Client's `readPump()` goroutine:**
```go
func (c *Client) readPump() {
    for {
        // 1. Read from WebSocket
        _, messageData, err := c.conn.ReadMessage()
        // messageData = {"type":"message","receiver_id":12,"content":"Hello!",...}
        
        // 2. Parse JSON
        var wsMsg models.WebSocketMessage
        json.Unmarshal(messageData, &wsMsg)
        
        // 3. Set sender (from authenticated user)
        wsMsg.SenderID = c.userID  // SenderID = 5
        wsMsg.Timestamp = time.Now()
        
        // 4. Route based on type
        switch wsMsg.Type {
        case "message":
            c.handleChatMessage(&wsMsg)  // Handle the message!
        }
    }
}
```

---

### Step 6: handleChatMessage() Processes

```go
func (c *Client) handleChatMessage(wsMsg *WebSocketMessage) {
    // 1. Check permission: Can User 5 chat with User 12?
    canChat, err := db.CanChat(c.hub.database, 5, 12)
    // Checks: (5 follows 12) OR (12 follows 5) OR (12 has public profile)
    
    if !canChat {
        c.sendError("You cannot send messages to this user")
        return
    }
    
    // 2. Save to database
    msg := &models.Message{
        SenderID:   5,
        ReceiverID: 12,
        Content:    "Hello!",
        IsRead:     false,
        CreatedAt:  time.Now(),
    }
    db.SaveMessage(c.hub.database, msg)  // INSERT INTO messages...
    // msg.ID = 142 (auto-generated)
    
    // 3. Update message with database ID
    wsMsg.MessageID = 142
    
    // 4. Broadcast to receiver
    c.hub.broadcast <- wsMsg  // Put in Hub's broadcast channel
    
    // 5. Send confirmation back to sender
    data, _ := json.Marshal(wsMsg)
    c.send <- data  // Put in sender's send channel
}
```

---

### Step 7: Hub Routes Message

**Hub's `Run()` receives broadcast:**
```go
case message := <-h.broadcast:  // Receives message for User 12
    h.mu.RLock()
    
    // Check if User 12 is online
    if client, online := h.clients[12]; online {
        // User 12 is online! Send it!
        data, _ := json.Marshal(message)
        client.send <- data  // Put in User 12's send channel
    } else {
        // User 12 is offline, message already in database
        log.Printf("User 12 offline, message saved for later")
    }
    
    h.mu.RUnlock()
```

---

### Step 8: writePump() Sends to Browser

**User 12's Client `writePump()` goroutine:**
```go
func (c *Client) writePump() {
    for {
        select {
        case message := <-c.send:  // Receives message from Hub
            // 1. Write to WebSocket
            c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            w, err := c.conn.NextWriter(websocket.TextMessage)
            w.Write(message)  // Send to browser!
            w.Close()
        }
    }
}
```

**Message arrives in User 12's browser instantly!**

---

### Step 9: Browser Displays Message

**Frontend JavaScript:**
```javascript
ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    // message = {type:"message", sender_id:5, content:"Hello!", ...}
    
    displayMessage(message);  // Show in chat UI
};
```

**User 12 sees:** "User A: Hello!"

---

### Complete Timeline Summary

```
[0.0s] User A opens browser
[0.1s] WebSocket connects, creates Client for User 5
[0.2s] Hub registers User 5 in clients map
[0.3s] readPump + writePump goroutines start
[1.0s] User A types "Hello!" and hits send
[1.1s] Browser sends JSON via WebSocket
[1.2s] readPump receives message
[1.3s] handleChatMessage checks permission → OK
[1.4s] Message saved to database (ID=142)
[1.5s] Message sent to Hub's broadcast channel
[1.6s] Hub receives, checks if User 12 online → YES
[1.7s] Hub sends to User 12's send channel
[1.8s] User 12's writePump receives from send channel
[1.9s] writePump writes to User 12's WebSocket
[2.0s] User 12's browser receives and displays "Hello!"
```

**Total time: 1 second (real-time!)**

[Back to Top](#table-of-contents)

---

## Group Chat - Real-time Group Messaging

### What is Group Chat?

**Group chat** allows multiple users to communicate in a shared conversation space. Unlike 1-on-1 chat where messages go to one person, group messages are **broadcast to all group members**.

### How It Works

The chat service handles both types of chat:
- **1-on-1 chat**: Message goes from User A → User B
- **Group chat**: Message goes from User A → All members of Group X

### Group Chat Architecture

```
User A (Member)  ──┐
                   │
User B (Member)  ──┤──> WebSocket Hub ──> Broadcast to all ──┐
                   │                                          │
User C (Member)  ──┘                                          │
                                                              │
                                                              ▼
                                                    ┌─────────────────┐
                                                    │  All members    │
                                                    │  receive msg    │
                                                    │  instantly      │
                                                    └─────────────────┘
```

### Database Tables

**groups table:**
```sql
CREATE TABLE groups (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    creator_id INTEGER NOT NULL
);
```

**group_members table:**
```sql
CREATE TABLE group_members (
    id INTEGER PRIMARY KEY,
    group_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    role TEXT CHECK (role IN ('admin', 'member')),
    status TEXT CHECK (status IN ('pending', 'accepted')),
    UNIQUE(group_id, user_id)
);
```

**group_messages table:**
```sql
CREATE TABLE group_messages (
    id INTEGER PRIMARY KEY,
    group_id INTEGER NOT NULL,
    sender_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME NOT NULL
);
```

### Access Control

Only **accepted group members** can:
- Send messages to the group
- Receive group messages
- View group chat history

The `IsGroupMember()` function checks:
```sql
SELECT COUNT(*) FROM group_members 
WHERE group_id = ? AND user_id = ? AND status = 'accepted'
```

### Group Chat Flow

Let's follow what happens when **User A** sends "Meeting at 3pm!" to **Group "Project Team"** (with members: User A, B, C, D):

---

**Step 1: User A sends message via WebSocket**

```javascript
// Browser JavaScript
const groupMessage = {
    type: "group_message",
    group_id: 5,
    content: "Meeting at 3pm!",
    timestamp: new Date()
};

ws.send(JSON.stringify(groupMessage));
```

---

**Step 2: Client.readPump() receives**

```go
// Parse incoming message
var wsMsg models.WebSocketMessage
json.Unmarshal(messageData, &wsMsg)
// wsMsg.Type = "group_message"
// wsMsg.GroupID = 5
// wsMsg.Content = "Meeting at 3pm!"

// Set sender from authenticated user
wsMsg.SenderID = c.userID  // User A's ID

// Route based on type
switch wsMsg.Type {
case "group_message":
    c.handleGroupChatMessage(&wsMsg)  // Handle group message
}
```

---

**Step 3: handleGroupChatMessage() processes**

```go
func (c *Client) handleGroupChatMessage(wsMsg *WebSocketMessage) {
    // 1. Check if sender is a member
    isMember, _ := db.IsGroupMember(c.hub.database, 5, c.userID)
    if !isMember {
        c.sendError("You are not a member of this group")
        return
    }
    
    // 2. Save to database
    msg := &models.GroupMessage{
        GroupID:   5,
        SenderID:  c.userID,  // User A
        Content:   "Meeting at 3pm!",
        CreatedAt: time.Now(),
    }
    db.SaveGroupMessage(c.hub.database, msg)
    // msg.ID = 89 (auto-generated)
    
    // 3. Update WebSocket message
    wsMsg.MessageID = 89
    wsMsg.Timestamp = msg.CreatedAt
    
    // 4. Broadcast to ALL group members
    c.hub.broadcast <- wsMsg
}
```

---

**Step 4: Hub broadcasts to all group members**

```go
case message := <-h.broadcast:
    if message.Type == "group_message" {
        // Get all group members
        members, _ := db.GetGroupMembers(database, 5)
        // members = [User A, User B, User C, User D]
        
        data, _ := json.Marshal(message)
        
        // Send to ALL online members
        for _, memberID := range members {
            if client, ok := h.clients[memberID]; ok {
                client.send <- data  // Send to their send channel
            }
        }
    }
```

---

**Step 5: All online members receive instantly**

- **User A** (sender): Receives confirmation
- **User B**: Sees "Meeting at 3pm!" appear instantly
- **User C**: Sees "Meeting at 3pm!" appear instantly  
- **User D**: Offline, will see it when they load history

---

### Timeline Example

```
[0.0s] User A types "Meeting at 3pm!" in Group "Project Team"
[0.1s] Browser sends WebSocket message (type: "group_message")
[0.2s] readPump receives and routes to handleGroupChatMessage
[0.3s] Check: Is User A a member? → YES
[0.4s] Save to group_messages table (ID=89)
[0.5s] Message sent to Hub's broadcast channel
[0.6s] Hub queries: Who are the members? → [A, B, C, D]
[0.7s] Hub checks: Who's online? → [A, B, C] (D is offline)
[0.8s] Send to User A's send channel → writePump sends to browser
[0.8s] Send to User B's send channel → writePump sends to browser
[0.8s] Send to User C's send channel → writePump sends to browser
[0.9s] All 3 browsers display "Meeting at 3pm!"
```

**Total time: 0.9 seconds (real-time!)**

---

### Differences from 1-on-1 Chat

| Feature | 1-on-1 Chat | Group Chat |
|---------|-------------|------------|
| **Recipients** | One person (ReceiverID) | All group members (GroupID) |
| **Access** | Follow/public profile check | Group membership check |
| **Database** | `messages` table | `group_messages` table |
| **Broadcast** | Send to 1 recipient | Send to N members |
| **Message Type** | `"message"` | `"group_message"` |

---

### REST API for Group Chat

**GET /chat/groups/:groupId/history**

Load group chat history:
```http
GET /chat/groups/5/history?limit=50 HTTP/1.1
Authorization: Bearer <token>
```

Response:
```json
{
    "success": true,
    "data": {
        "messages": [
            {
                "id": 87,
                "group_id": 5,
                "sender_id": 12,
                "content": "Hello team!",
                "created_at": "2025-10-16T14:00:00Z"
            },
            {
                "id": 89,
                "group_id": 5,
                "sender_id": 5,
                "content": "Meeting at 3pm!",
                "created_at": "2025-10-16T14:05:00Z"
            }
        ],
        "count": 2
    }
}
```

**POST /chat/groups/:groupId/messages**

Send group message via HTTP (fallback):
```http
POST /chat/groups/5/messages HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
    "content": "Meeting at 3pm!"
}
```

Response:
```json
{
    "success": true,
    "data": {
        "message": {
            "id": 89,
            "group_id": 5,
            "sender_id": 5,
            "content": "Meeting at 3pm!",
            "created_at": "2025-10-16T14:05:00Z"
        }
    }
}
```

Note: Even when sent via HTTP, the message still broadcasts via WebSocket to online members!

---

### Frontend Integration

**Connect to WebSocket (same connection for both 1-on-1 and group chat):**
```javascript
const ws = new WebSocket('ws://localhost:8085/ws');

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    
    if (message.type === 'message') {
        // 1-on-1 message
        displayPrivateMessage(message);
    } else if (message.type === 'group_message') {
        // Group message
        displayGroupMessage(message);
    }
};
```

**Send group message:**
```javascript
function sendGroupMessage(groupId, content) {
    const message = {
        type: 'group_message',
        group_id: groupId,
        content: content,
        timestamp: new Date().toISOString()
    };
    
    ws.send(JSON.stringify(message));
}
```

**Load group chat history:**
```javascript
async function loadGroupHistory(groupId) {
    const response = await fetch(`http://localhost:8085/chat/groups/${groupId}/history?limit=50`, {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    
    const data = await response.json();
    displayGroupMessages(data.data.messages);
}
```

[Back to Top](#table-of-contents)

---

## Access Control - Who Can Chat?

### The Rule

You can chat with someone if **ANY** of these is true:

1. You follow them (with accepted follow)
2. They follow you (with accepted follow)
3. They have a public profile

### CanChat() Function

```go
func CanChat(db *sql.DB, senderID, receiverID int) (bool, error) {
    query := `
        SELECT 
            CASE 
                -- Check if receiver has public profile
                WHEN (SELECT is_public_profile FROM users WHERE id = ?) = 1 
                    THEN 1
                
                -- Check if sender follows receiver
                WHEN EXISTS (
                    SELECT 1 FROM follows 
                    WHERE follower_id = ? AND following_id = ? AND status = 'accepted'
                ) THEN 1
                
                -- Check if receiver follows sender
                WHEN EXISTS (
                    SELECT 1 FROM follows 
                    WHERE follower_id = ? AND following_id = ? AND status = 'accepted'
                ) THEN 1
                
                ELSE 0
            END as can_chat
    `
    
    var canChat int
    err := db.QueryRow(query, receiverID, senderID, receiverID, receiverID, senderID).Scan(&canChat)
    
    return canChat == 1, err
}
```

### Examples

**Example 1: Following relationship**
```
User A (ID=5) follows User B (ID=12) → Status: accepted
CanChat(5, 12) → TRUE
```

**Example 2: Public profile**
```
User C (ID=20) has is_public_profile = true
User D (ID=30) doesn't follow C, and C doesn't follow D
CanChat(30, 20) → TRUE (because C is public)
```

**Example 3: No relationship**
```
User E (ID=40) has is_public_profile = false
User F (ID=50) doesn't follow E, and E doesn't follow F
CanChat(50, 40) → FALSE
```

### When is CanChat() Called?

Every time a message is sent:
```go
func (c *Client) handleChatMessage(wsMsg *WebSocketMessage) {
    // FIRST THING: Check permission
    canChat, err := db.CanChat(c.hub.database, c.userID, wsMsg.ReceiverID)
    
    if !canChat {
        c.sendError("You cannot send messages to this user")
        return  // Stop here, don't send message
    }
    
    // Permission granted, continue...
}
```

[Back to Top](#table-of-contents)

---

## Message Types and Payloads

### WebSocketMessage Structure

All WebSocket messages use this format:

```go
type WebSocketMessage struct {
    Type       string    `json:"type"`         // "message", "group_message", "typing", "read", "error"
    MessageID  int       `json:"message_id"`   // Database ID (after saved)
    SenderID   int       `json:"sender_id"`    // Who sent it
    ReceiverID int       `json:"receiver_id"`  // Who receives it (1-on-1 chat)
    GroupID    int       `json:"group_id"`     // Which group (group chat)
    Content    string    `json:"content"`      // Message text
    Timestamp  time.Time `json:"timestamp"`    // When sent
    Error      string    `json:"error"`        // Error message (if type="error")
}
```

### Message Type: "message"

**Purpose:** Send a 1-on-1 chat message

**Browser → Server:**
```json
{
    "type": "message",
    "receiver_id": 12,
    "content": "Hello! How are you?",
    "timestamp": "2025-10-16T10:30:00Z"
}
```

**Server → Browser (after saving):**
```json
{
    "type": "message",
    "message_id": 142,
    "sender_id": 5,
    "receiver_id": 12,
    "content": "Hello! How are you?",
    "timestamp": "2025-10-16T10:30:00Z"
}
```

---

### Message Type: "group_message"

**Purpose:** Send a message to a group

**Browser → Server:**
```json
{
    "type": "group_message",
    "group_id": 5,
    "content": "Meeting at 3pm!",
    "timestamp": "2025-10-16T14:05:00Z"
}
```

**Server → All Group Members (after saving):**
```json
{
    "type": "group_message",
    "message_id": 89,
    "sender_id": 12,
    "group_id": 5,
    "content": "Meeting at 3pm!",
    "timestamp": "2025-10-16T14:05:00Z"
}
```

**What happens:**
- Message saved to `group_messages` table
- Broadcast to ALL group members who are online
- Offline members see it when they load history

---

### Message Type: "read"

**Purpose:** Mark messages as read

**Browser → Server:**
```json
{
    "type": "read",
    "sender_id": 12,
    "receiver_id": 5
}
```

**What happens:**
```sql
UPDATE messages 
SET is_read = 1 
WHERE sender_id = 12 AND recipient_id = 5 AND is_read = 0
```

---

### Message Type: "typing"

**Purpose:** Show "User is typing..." indicator

**Browser → Server:**
```json
{
    "type": "typing",
    "sender_id": 5,
    "receiver_id": 12
}
```

**Server → Receiver's Browser:**
```json
{
    "type": "typing",
    "sender_id": 5,
    "receiver_id": 12,
    "timestamp": "2025-10-16T10:30:05Z"
}
```

**Frontend displays:** "User A is typing..."

---

### Message Type: "error"

**Purpose:** Notify user of error

**Server → Browser:**
```json
{
    "type": "error",
    "error": "You cannot send messages to this user",
    "timestamp": "2025-10-16T10:30:10Z"
}
```

**Frontend displays:** Error notification

[Back to Top](#table-of-contents)

---

## Database Operations

### Messages Table Schema

```sql
CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sender_id INTEGER,
    recipient_id INTEGER,
    content TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (recipient_id) REFERENCES users(id) ON DELETE SET NULL
);
```

### SaveMessage()

**Saves message to database permanently**

```go
func SaveMessage(db *sql.DB, msg *models.Message) error {
    query := `
        INSERT INTO messages (sender_id, recipient_id, content, is_read, created_at)
        VALUES (?, ?, ?, ?, ?)
    `
    result, err := db.Exec(query, 
        msg.SenderID,    // 5
        msg.ReceiverID,  // 12
        msg.Content,     // "Hello!"
        msg.IsRead,      // false
        msg.CreatedAt)   // 2025-10-16 10:30:00
    
    // Get auto-generated ID
    id, _ := result.LastInsertId()
    msg.ID = int(id)  // msg.ID = 142
    
    return nil
}
```

---

### GetChatHistory()

**Retrieves messages between two users**

```go
func GetChatHistory(db *sql.DB, user1ID, user2ID int, limit int) ([]models.Message, error) {
    query := `
        SELECT id, sender_id, recipient_id, content, is_read, created_at
        FROM messages
        WHERE (sender_id = ? AND recipient_id = ?) 
           OR (sender_id = ? AND recipient_id = ?)
        ORDER BY created_at DESC
        LIMIT ?
    `
    
    rows, _ := db.Query(query, user1ID, user2ID, user2ID, user1ID, limit)
    
    // Parse all messages
    var messages []models.Message
    for rows.Next() {
        var msg models.Message
        rows.Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID, 
                  &msg.Content, &msg.IsRead, &msg.CreatedAt)
        messages = append(messages, msg)
    }
    
    // Reverse for chronological order (oldest first)
    // [newest] → [oldest] becomes [oldest] → [newest]
    
    return messages, nil
}
```

**Example result:**
```go
messages = [
    {ID: 140, SenderID: 5, ReceiverID: 12, Content: "Hi!", ...},
    {ID: 141, SenderID: 12, ReceiverID: 5, Content: "Hey!", ...},
    {ID: 142, SenderID: 5, ReceiverID: 12, Content: "Hello!", ...},
]
```

---

### GetConversations()

**Lists all conversations for a user**

```go
func GetConversations(db *sql.DB, userID int) ([]models.Conversation, error) {
    query := `
        SELECT 
            u.id,
            u.username,
            u.first_name,
            u.last_name,
            u.nickname,
            m.content as last_message,
            m.created_at as last_message_at,
            (SELECT COUNT(*) FROM messages 
             WHERE sender_id = u.id AND recipient_id = ? AND is_read = 0
            ) as unread_count
        FROM (
            -- Get all users I've chatted with
            SELECT DISTINCT 
                CASE WHEN sender_id = ? THEN recipient_id ELSE sender_id END as other_user_id
            FROM messages
            WHERE sender_id = ? OR recipient_id = ?
        ) conv
        JOIN users u ON u.id = conv.other_user_id
        LEFT JOIN messages m ON m.id = (
            -- Get last message with this user
            SELECT id FROM messages
            WHERE (sender_id = ? AND recipient_id = u.id)
               OR (sender_id = u.id AND recipient_id = ?)
            ORDER BY created_at DESC
            LIMIT 1
        )
        ORDER BY m.created_at DESC
    `
    
    // Returns list of conversations with last message and unread count
}
```

**Example result:**
```go
conversations = [
    {
        UserID: 12,
        Username: "bob",
        FirstName: "Bob",
        LastName: "Smith",
        LastMessage: "See you tomorrow!",
        LastMessageAt: "2025-10-16 10:45:00",
        UnreadCount: 3,
        IsOnline: true
    },
    {
        UserID: 8,
        Username: "alice",
        FirstName: "Alice",
        LastName: "Johnson",
        LastMessage: "Thanks!",
        LastMessageAt: "2025-10-16 09:20:00",
        UnreadCount: 0,
        IsOnline: false
    }
]
```

[Back to Top](#table-of-contents)

---

## HTTP REST Endpoints

### Why REST Endpoints with WebSocket?

**WebSocket** is for real-time messaging, but we also need:
- Load old messages (history)
- List conversations
- Send messages when WebSocket fails (fallback)

### Available Endpoints

| Method | Endpoint | Purpose | Auth Required |
|--------|----------|---------|---------------|
| GET | `/health` | Health check | No |
| GET | `/ws` | WebSocket upgrade | Yes |
| GET | `/chat/conversations` | List all 1-on-1 conversations | Yes |
| GET | `/chat/history/:userId` | Get 1-on-1 chat history | Yes |
| POST | `/chat/read/:userId` | Mark 1-on-1 messages as read | Yes |
| GET | `/chat/unread` | Get unread count | Yes |
| POST | `/chat/send` | Send 1-on-1 message (HTTP fallback) | Yes |
| GET | `/chat/groups/:groupId/history` | Get group chat history | Yes |
| POST | `/chat/groups/:groupId/messages` | Send group message (HTTP fallback) | Yes |

---

### GET /chat/conversations

**Purpose:** Get list of all conversations

**Request:**
```http
GET /chat/conversations HTTP/1.1
Host: localhost:8085
Authorization: Bearer <token>
```

**Response:**
```json
{
    "success": true,
    "data": {
        "conversations": [
            {
                "user_id": 12,
                "username": "bob",
                "first_name": "Bob",
                "last_name": "Smith",
                "nickname": "Bobby",
                "last_message": "See you!",
                "last_message_at": "2025-10-16T10:45:00Z",
                "unread_count": 3,
                "is_online": true
            }
        ],
        "count": 1
    }
}
```

---

### GET /chat/history/:userId

**Purpose:** Get message history with specific user

**Request:**
```http
GET /chat/history/12?limit=50 HTTP/1.1
Host: localhost:8085
Authorization: Bearer <token>
```

**Response:**
```json
{
    "success": true,
    "data": {
        "messages": [
            {
                "id": 140,
                "sender_id": 5,
                "receiver_id": 12,
                "content": "Hi there!",
                "is_read": true,
                "created_at": "2025-10-16T10:30:00Z"
            },
            {
                "id": 141,
                "sender_id": 12,
                "receiver_id": 5,
                "content": "Hey! How are you?",
                "is_read": true,
                "created_at": "2025-10-16T10:31:00Z"
            }
        ],
        "count": 2
    }
}
```

---

### POST /chat/send

**Purpose:** Send message via HTTP (fallback if WebSocket fails)

**Request:**
```http
POST /chat/send HTTP/1.1
Host: localhost:8085
Authorization: Bearer <token>
Content-Type: application/json

{
    "receiver_id": 12,
    "content": "Hello via HTTP!"
}
```

**Response:**
```json
{
    "success": true,
    "data": {
        "message": {
            "id": 143,
            "sender_id": 5,
            "receiver_id": 12,
            "content": "Hello via HTTP!",
            "is_read": false,
            "created_at": "2025-10-16T10:50:00Z"
        }
    }
}
```

**Note:** If receiver is online, they still get it via WebSocket instantly!

[Back to Top](#table-of-contents)

---

## Frontend Integration

### Connecting to WebSocket

```javascript
// 1. Create WebSocket connection
const token = localStorage.getItem('session_token');
const ws = new WebSocket(`ws://localhost:8085/ws`);

// 2. Send token after connection
ws.onopen = () => {
    console.log('Connected to chat!');
    // Send auth in first message or via query param
};

// 3. Handle incoming messages
ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    
    switch(message.type) {
        case 'message':
            displayNewMessage(message);
            break;
        case 'typing':
            showTypingIndicator(message.sender_id);
            break;
        case 'error':
            showError(message.error);
            break;
    }
};

// 4. Handle disconnection
ws.onclose = () => {
    console.log('Disconnected from chat');
    // Auto-reconnect after 3 seconds
    setTimeout(() => connectWebSocket(), 3000);
};

// 5. Handle errors
ws.onerror = (error) => {
    console.error('WebSocket error:', error);
};
```

---

### Sending a Message

```javascript
function sendMessage(receiverId, content) {
    const message = {
        type: 'message',
        receiver_id: receiverId,
        content: content,
        timestamp: new Date().toISOString()
    };
    
    // Send via WebSocket
    ws.send(JSON.stringify(message));
    
    // Display immediately in UI (optimistic update)
    displayMyMessage(content);
}
```

---

### Loading Chat History

```javascript
async function loadChatHistory(userId) {
    const response = await fetch(`http://localhost:8085/chat/history/${userId}?limit=50`, {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    
    const data = await response.json();
    
    if (data.success) {
        displayMessages(data.data.messages);
    }
}
```

---

### Loading Conversations List

```javascript
async function loadConversations() {
    const response = await fetch('http://localhost:8085/chat/conversations', {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    
    const data = await response.json();
    
    if (data.success) {
        displayConversationsList(data.data.conversations);
    }
}
```

---

### Marking Messages as Read

```javascript
async function markAsRead(userId) {
    await fetch(`http://localhost:8085/chat/read/${userId}`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
}
```

---

### Complete Frontend Flow

```
[User opens chat]
  ↓
[Connect WebSocket] → ws://localhost:8085/ws
  ↓
[Load conversations] → GET /chat/conversations
  ↓
[Display conversation list with unread badges]
  ↓
[User clicks on conversation with Bob]
  ↓
[Load history] → GET /chat/history/12
  ↓
[Display messages]
  ↓
[Mark as read] → POST /chat/read/12
  ↓
[User types and sends message]
  ↓
[Send via WebSocket] → ws.send(JSON.stringify({...}))
  ↓
[Receive confirmation from server]
  ↓
[Bob's browser receives via WebSocket instantly]
```

[Back to Top](#table-of-contents)

---

## Error Handling and Recovery

### Connection Loss

**What happens when user loses internet?**

1. **Browser detects:** `ws.onclose` event fires
2. **Auto-reconnect:**
```javascript
ws.onclose = () => {
    console.log('Connection lost, reconnecting...');
    setTimeout(() => {
        connectWebSocket();  // Try again in 3 seconds
    }, 3000);
};
```

3. **Server cleans up:**
```go
// readPump() detects connection closed
defer func() {
    c.hub.unregister <- c  // Remove from Hub
    c.conn.Close()          // Clean up connection
}()
```

---

### Message Delivery Guarantee

**What if receiver is offline?**

```go
case message := <-h.broadcast:
    if client, online := h.clients[receiverID]; online {
        // Online: Send immediately via WebSocket
        client.send <- data
    } else {
        // Offline: Already saved in database!
        // They'll see it when they load history
        log.Printf("User %d offline, message stored for later", receiverID)
    }
```

**When offline user comes back:**
1. Connects to WebSocket
2. Loads conversations → GET /chat/conversations
3. Sees unread count: `"unread_count": 5`
4. Opens conversation → GET /chat/history/12
5. All messages loaded from database!

---

### Permission Denied

**What if user tries to message someone they can't?**

```go
canChat, _ := db.CanChat(senderID, receiverID)
if !canChat {
    // Send error back to sender
    c.sendError("You cannot send messages to this user")
    return  // Don't save message, don't broadcast
}
```

**Browser receives:**
```json
{
    "type": "error",
    "error": "You cannot send messages to this user"
}
```

**Frontend displays:** "No Cannot send message. You must follow this user or they must have a public profile."

---

### Goroutine Panics

**What if a goroutine crashes?**

Each goroutine has `defer` for cleanup:

```go
func (c *Client) readPump() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("readPump panic: %v", r)
        }
        c.hub.unregister <- c  // Always unregister
        c.conn.Close()          // Always close connection
    }()
    
    // Main loop...
}
```

Even if something crashes, connection is cleaned up properly.

---

### Database Errors

**What if database is locked?**

SQLite has retries:
```go
db.SetMaxOpenConns(25)   // Max 25 connections
db.SetMaxIdleConns(5)    // Keep 5 ready
```

If write fails:
```go
err := db.SaveMessage(database, msg)
if err != nil {
    log.Printf("Failed to save message: %v", err)
    c.sendError("Failed to send message, please try again")
    return  // Don't broadcast if not saved
}
```

[Back to Top](#table-of-contents)

---

## Final Summary

### Key Concepts Recap

1. **WebSocket** = Permanent connection for real-time communication
2. **Hub** = Central manager tracking all connections
3. **Client** = Individual user's WebSocket connection
4. **Goroutines** = Concurrent functions (readPump, writePump, Hub.Run)
5. **Channels** = Safe communication between goroutines
6. **Access Control** = Follow relationships + public profiles

### The Complete Picture

```
                         ┌─────────────┐
                         │     Hub     │
                         │  (Central)  │
                         │             │
                         │  clients{}  │
                         │  broadcast  │
                         │  register   │
                         │  unregister │
                         └──────┬──────┘
                                │
                ┌───────────────┼───────────────┐
                │               │               │
         ┌──────▼──────┐ ┌─────▼──────┐ ┌─────▼──────┐
         │  Client #1  │ │ Client #2  │ │ Client #3  │
         │  (User A)   │ │  (User B)  │ │  (User C)  │
         │             │ │            │ │            │
         │ readPump()  │ │ readPump() │ │ readPump() │
         │ writePump() │ │writePump() │ │writePump() │
         └──────┬──────┘ └─────┬──────┘ └─────┬──────┘
                │               │               │
                └───────────────┼───────────────┘
                                │
                         ┌──────▼──────┐
                         │  Database   │
                         │  (SQLite)   │
                         │             │
                         │  messages   │
                         │  users      │
                         │  follows    │
                         └─────────────┘
```

### Message Flow Summary

```
User types message
     ↓
Browser sends via WebSocket
     ↓
Client.readPump() receives
     ↓
handleChatMessage() checks permission
     ↓
Save to database
     ↓
Send to Hub.broadcast channel
     ↓
Hub routes to receiver
     ↓
Receiver's Client.writePump() sends
     ↓
Receiver's browser displays
     ↓
DONE! (All in < 1 second)
```

---

**Congratulations! You now understand the complete WebSocket chat system!(OR at least you think you do...In that case,read again!)**

[Back to Top](#table-of-contents)
