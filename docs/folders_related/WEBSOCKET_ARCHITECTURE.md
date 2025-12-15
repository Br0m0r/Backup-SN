# WebSocket Architecture Documentation

## Overview

Your backend has **TWO separate WebSocket servers** running on different ports:

1. **Chat Service** (Port 8085) - Real-time chat messages
2. **Notification Service** (Port 8086) - Real-time notifications

## Why Two Connections?

### Separation of Concerns
- **Chat** handles high-frequency messages (typing, messages, read receipts)
- **Notifications** handles low-frequency events (follows, likes, comments)
- Prevents chat message flooding from blocking notifications

### Scalability
- Each service can scale independently
- Chat service can have different rate limits
- Notification service can be more restricted

### Backend Architecture Alignment
Your Go services are already separated, so frontend mirrors this:
```
services/chat/         → Chat WebSocket
services/notifications/ → Notifications WebSocket
```

## Authentication Flow

### How Backend Authenticates WebSocket

From your `handlers/websocket.go`:
```go
userID, ok := middleware.GetUserIDFromContext(r)
if !ok {
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
    return
}
```

The middleware extracts user ID from the request. Here's how:

1. **User logs in** → Gets JWT token
2. **Frontend stores** token in localStorage
3. **WebSocket connects** with token in URL: `ws://localhost:8085/ws?token=xxx`
4. **Backend middleware** validates token and extracts user ID
5. **Connection established** with authenticated user ID

### Security Considerations

**Current (Development):**
```javascript
const wsUrl = `${config.chatUrl}?token=${token}&username=${username}`
```

**Production Improvements:**
- Move token to WebSocket headers (if supported)
- Use secure WebSocket (wss://)
- Implement token refresh mechanism
- Add connection fingerprinting

## Message Flow

### Sending a Message (1-on-1 Chat)

**Frontend:**
```javascript
sendMessage(receiverId, content)
  ↓
{
  type: 'message',
  receiver_id: 123,
  content: 'Hello!',
  timestamp: '2025-11-12T...'
}
  ↓
WebSocket.send()
```

**Backend (your chat/handlers/websocket.go):**
```go
1. Receives message in readPump()
2. Parses JSON → WebSocketMessage
3. handleChatMessage() is called:
   - Checks if users can chat (following/privacy)
   - Saves message to database
   - Gets message ID
   - Broadcasts to receiver if online
   - Sends notification if receiver offline
4. Sends confirmation back to sender
```

**Receiver Frontend:**
```javascript
WebSocket.onmessage
  ↓
handleIncomingMessage(data)
  ↓
emit('message', data)
  ↓
Chat component listener receives message
  ↓
Updates UI
```

### Group Message Flow

**Frontend:**
```javascript
{
  type: 'group_message',
  group_id: 456,
  content: 'Hello group!',
  timestamp: '2025-11-12T...'
}
```

**Backend:**
```go
1. handleGroupChatMessage() is called
2. Checks group membership
3. Saves to database
4. Broadcasts to ALL group members
5. Notifies offline members
```

## Connection Lifecycle

### Initial Connection
```
App loads
  ↓
User logs in → Token stored
  ↓
WebSocket.connect() called
  ↓
Authentication successful
  ↓
connected = true
  ↓
Start heartbeat timer
```

### Heartbeat (Keep-Alive)
```
Every 30 seconds:
  Frontend sends: { type: 'ping' }
  Backend responds: pong message
  
If no response → Connection stale
  → Auto reconnect
```

### Reconnection
```
Connection lost (network issue, server restart)
  ↓
onclose triggered
  ↓
Attempt 1: Wait 3 seconds → Reconnect
Attempt 2: Wait 6 seconds → Reconnect
Attempt 3: Wait 9 seconds → Reconnect
...
Max 5 attempts
```

### Graceful Shutdown
```
User logs out
  ↓
disconnect() called
  ↓
Clear timers
  ↓
ws.close(1000, 'Client disconnect')
  ↓
No reconnection attempted
```

## Event System

### Why Event-Driven?

**Problem:** Multiple components need the same WebSocket data
- Chat component needs messages
- Notification badge needs message count
- Profile needs online status

**Solution:** Publish-Subscribe Pattern
```javascript
// Component A
on('message', (data) => {
  updateChatUI(data)
})

// Component B
on('message', (data) => {
  updateUnreadCount()
})

// When message arrives
emit('message', data) // Both listeners triggered
```

### Event Types

#### Chat Events
- `message` - New 1-on-1 message
- `group_message` - New group message
- `typing` - Someone is typing
- `read` - Message read receipt
- `online_status` - User online/offline
- `connected` - WebSocket connected
- `disconnected` - WebSocket disconnected
- `error` - Server error

#### Usage Pattern
```javascript
const { on, off } = useWebSocket()

onMounted(() => {
  // Register listener
  on('message', handleNewMessage)
})

onUnmounted(() => {
  // Cleanup listener
  off('message', handleNewMessage)
})
```

## State Management

### Singleton Pattern

**Why:** Only ONE WebSocket connection per service
```javascript
// Global state (outside function)
const ws = ref(null)
const connected = ref(false)

// Every component calling useWebSocket() gets the SAME instance
export function useWebSocket() {
  return { connected, connect, ... }
}
```

**Benefits:**
- Prevents duplicate connections
- Shared state across components
- Efficient resource usage

### Reactive State

```javascript
const connected = ref(false)  // Reactive

// In component
const { connected } = useWebSocket()

watchEffect(() => {
  if (connected.value) {
    console.log('Connected!')
  }
})
```

Changes propagate automatically to all components using the composable.

## Online Status Tracking

### Why Use Set?
```javascript
const onlineUsers = ref(new Set())

// O(1) lookup
if (onlineUsers.value.has(userId)) {
  // User is online
}

// O(1) add/remove
onlineUsers.value.add(userId)
onlineUsers.value.delete(userId)
```

**Alternative (Array):** O(n) lookup - slower for many users

### Backend Integration

Your backend tracks connected clients:
```go
type Hub struct {
    clients map[int]*Client  // userID -> Client
}
```

Frontend mirrors this with Set of online user IDs.

## Error Handling

### Connection Errors
```javascript
ws.onerror = (error) => {
  console.error('WebSocket error:', error)
  // User sees "Connecting..." in UI
  // Auto-reconnect attempts
}
```

### Message Send Failures
```javascript
function sendMessage(receiverId, content) {
  if (!connected) {
    return false  // Caller can show error UI
  }
  
  try {
    ws.send(...)
    return true
  } catch (error) {
    return false
  }
}
```

### Server Errors
```javascript
// Backend sends
{ type: 'error', error: 'You cannot send messages to this user' }

// Frontend emits
emit('error', data)

// Components can listen
on('error', (data) => {
  showErrorToast(data.error)
})
```

## Integration with Chat Component

### Current Implementation
```javascript
// Chat.vue uses this line:
import { useWebSocket } from '../composables/useWebSocket'

const { connected, sendMessage, on, wsState } = useWebSocket()

// Listen for messages
on('message', (data) => {
  // Add to chat window
})

// Send message
sendMessage(chat.user_id, content)

// Show online status
chat.is_online = wsState.onlineUsers.has(chat.user_id)
```

## Next Steps

### 1. Create Notifications Composable
Similar to chat, but for notifications:
- Connect to `:8086/ws`
- Handle notification events
- Show notification badge

### 2. Initialize on App Load
```javascript
// In App.vue or main.js
onMounted(() => {
  if (isAuthenticated()) {
    const chat = useWebSocket()
    const notif = useNotifications()
    
    chat.connect()
    notif.connect()
  }
})
```

### 3. Cleanup on Logout
```javascript
function logout() {
  const chat = useWebSocket()
  const notif = useNotifications()
  
  chat.disconnect()
  notif.disconnect()
  clearUser()
}
```

## Testing Checklist

- [ ] Connect to WebSocket on login
- [ ] Send message successfully
- [ ] Receive message in real-time
- [ ] Online status updates
- [ ] Reconnect after network loss
- [ ] Graceful disconnect on logout
- [ ] Multiple tabs (same user) handling
- [ ] Error messages display correctly
- [ ] Typing indicators work
- [ ] Group messages work
- [ ] Read receipts work

## Common Issues

### Issue: "WebSocket is already closed"
**Cause:** Trying to send when disconnected
**Fix:** Check `connected.value` before sending

### Issue: Messages not received
**Cause:** Not listening to events or wrong user ID
**Fix:** Verify `on('message', ...)` is registered

### Issue: Connection keeps dropping
**Cause:** Backend timeout or heartbeat not working
**Fix:** Check backend logs, verify ping/pong

### Issue: Duplicate connections
**Cause:** Calling `connect()` multiple times
**Fix:** Check if already connected before connecting

## Performance Considerations

### Memory Leaks
- Clear event listeners on unmount
- Don't accumulate messages indefinitely
- Limit online users Set size

### Network Efficiency
- Batch messages when possible
- Compress large payloads
- Implement message queuing for offline sends

### UI Performance
- Debounce typing indicators
- Virtualize message lists for long conversations
- Lazy load chat history
