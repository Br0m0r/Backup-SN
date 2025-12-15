# WebSocket Composable - Step by Step Summary

## What We Created

### 1. **Authentication Store** (`stores/auth.js`)
**Purpose:** Manages user session and provides auth credentials to WebSocket

**Key Functions:**
- `getUser()` - Get current user (needed for WebSocket)
- `getToken()` - Get JWT token (for WebSocket auth)
- `setUser()` - Store session after login
- `clearUser()` - Clear session on logout

**Why Needed:**
Your backend requires authentication:
```go
userID, ok := middleware.GetUserIDFromContext(r)
```

### 2. **Chat WebSocket Composable** (`composables/useWebSocket.js`)
**Purpose:** Manages real-time chat WebSocket connection

**Key Features:**
- ✅ Single connection (singleton pattern)
- ✅ Auto-reconnect on disconnect
- ✅ Heartbeat to keep alive
- ✅ Event system (pub/sub)
- ✅ Online status tracking
- ✅ Type-safe message handling

## How It Works - Visual Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    USER LOGS IN                              │
│  Email/Password → Backend → JWT Token + User Data           │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│          STORE SESSION (stores/auth.js)                      │
│  setUser(userData, token)                                    │
│  localStorage.setItem('user', ...)                           │
│  localStorage.setItem('token', ...)                          │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│     CONNECT WEBSOCKET (useWebSocket composable)              │
│  const { connect } = useWebSocket()                          │
│  connect()                                                   │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│  BUILD WEBSOCKET URL WITH TOKEN                              │
│  ws://localhost:8085/ws?token=xxx&username=john              │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│            BACKEND VALIDATES TOKEN                           │
│  middleware.GetUserIDFromContext(r) → userID: 123            │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│          CONNECTION ESTABLISHED ✅                            │
│  ws.onopen() triggered                                       │
│  connected = true                                            │
│  Start heartbeat timer                                       │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│     CHAT COMPONENT REGISTERS EVENT LISTENERS                 │
│  on('message', handleNewMessage)                             │
│  on('online_status', updateOnlineIndicator)                  │
│  on('typing', showTypingIndicator)                           │
└─────────────────────────────────────────────────────────────┘
```

## Message Flow - Sending

```
USER TYPES MESSAGE → "Hello!"
         │
         ▼
┌────────────────────────────────────┐
│  Chat Component                    │
│  sendMessage(receiverId, content)  │
└────────┬───────────────────────────┘
         │
         ▼
┌────────────────────────────────────┐
│  useWebSocket Composable           │
│  Formats message:                  │
│  {                                 │
│    type: 'message',                │
│    receiver_id: 456,               │
│    content: 'Hello!',              │
│    timestamp: '2025-11-12...'      │
│  }                                 │
└────────┬───────────────────────────┘
         │
         ▼
┌────────────────────────────────────┐
│  WebSocket Connection              │
│  ws.send(JSON.stringify(message))  │
└────────┬───────────────────────────┘
         │
         ▼
┌────────────────────────────────────┐
│  Backend (Go)                      │
│  1. readPump() receives message    │
│  2. Unmarshal JSON                 │
│  3. handleChatMessage():           │
│     - Check permissions            │
│     - Save to database             │
│     - Get message ID               │
│     - Broadcast to receiver        │
│  4. Send confirmation to sender    │
└────────┬───────────────────────────┘
         │
         ▼
┌────────────────────────────────────┐
│  Receiver's WebSocket              │
│  Receives broadcasted message      │
└────────┬───────────────────────────┘
         │
         ▼
┌────────────────────────────────────┐
│  Receiver's Frontend               │
│  ws.onmessage → emit('message')    │
│  Chat component displays message   │
└────────────────────────────────────┘
```

## Key Concepts Explained

### 1. **Singleton Pattern**
```javascript
// Outside function = shared across all components
const ws = ref(null)
const connected = ref(false)

export function useWebSocket() {
  // Every component gets the SAME ws instance
  return { connected, connect, ... }
}
```

**Why:** Prevent duplicate connections, share state

### 2. **Event System (Pub/Sub)**
```javascript
// Component A subscribes
on('message', handleMessageA)

// Component B subscribes  
on('message', handleMessageB)

// When message arrives
emit('message', data) // Both A and B notified
```

**Why:** Decouple WebSocket from UI, multiple listeners

### 3. **Auto-Reconnect**
```javascript
ws.onclose = () => {
  // Connection lost!
  if (reconnectAttempts < 5) {
    setTimeout(() => connect(), 3000)
  }
}
```

**Why:** Handle network issues gracefully

### 4. **Heartbeat**
```javascript
setInterval(() => {
  ws.send(JSON.stringify({ type: 'ping' }))
}, 30000)
```

**Why:** Keep connection alive, detect stale connections

### 5. **Reactive State**
```javascript
const connected = ref(false)  // Vue reactive

// In component - auto-updates UI
<div v-if="connected">✅ Connected</div>
```

**Why:** UI automatically reflects connection status

## Usage in Components

### Example: Chat Component
```javascript
import { useWebSocket } from '@/composables/useWebSocket'

export default {
  setup() {
    const { connected, sendMessage, on } = useWebSocket()
    
    // Listen for messages
    onMounted(() => {
      on('message', (data) => {
        // Add message to UI
        messages.value.push(data)
      })
    })
    
    // Send message
    function send() {
      const success = sendMessage(receiverId, messageText)
      if (!success) {
        alert('Failed to send - not connected')
      }
    }
    
    return { connected, send }
  }
}
```

### Example: Online Status Badge
```javascript
import { useWebSocket } from '@/composables/useWebSocket'

export default {
  setup() {
    const { wsState } = useWebSocket()
    
    function isOnline(userId) {
      return wsState.value.onlineUsers.has(userId)
    }
    
    return { isOnline }
  }
}
```

## Environment Variables

Create `.env` file:
```bash
# Chat WebSocket URL
VITE_CHAT_WS_URL=ws://localhost:8085/ws

# For production:
# VITE_CHAT_WS_URL=wss://api.yoursite.com/chat/ws
```

## Next Steps

### Immediate:
1. ✅ Auth store created
2. ✅ Chat WebSocket composable created
3. ⏳ Create Notifications composable (similar pattern)
4. ⏳ Initialize WebSocket in App.vue on login
5. ⏳ Test with backend

### Testing:
```javascript
// In browser console after login:
const ws = new WebSocket('ws://localhost:8085/ws?token=YOUR_TOKEN')
ws.onopen = () => console.log('Connected!')
ws.onmessage = (e) => console.log('Message:', e.data)
ws.send(JSON.stringify({
  type: 'message',
  receiver_id: 2,
  content: 'Test message'
}))
```

### Production Checklist:
- [ ] Use WSS (secure WebSocket)
- [ ] Implement token refresh
- [ ] Add connection status UI
- [ ] Handle multiple tabs
- [ ] Add message queuing for offline sends
- [ ] Implement proper error handling
- [ ] Add analytics/monitoring
- [ ] Rate limiting on frontend

## Common Patterns

### Pattern 1: Connect on Login
```javascript
// After successful login
setUser(userData, token)
const { connect } = useWebSocket()
connect()
```

### Pattern 2: Disconnect on Logout
```javascript
function logout() {
  const { disconnect } = useWebSocket()
  disconnect()
  clearUser()
}
```

### Pattern 3: Show Connection Status
```vue
<template>
  <div class="status-indicator" :class="{ connected, connecting }">
    <span v-if="connecting">Connecting...</span>
    <span v-else-if="connected">✅ Connected</span>
    <span v-else>❌ Disconnected</span>
  </div>
</template>

<script setup>
const { connected, connecting } = useWebSocket()
</script>
```

### Pattern 4: Optimistic Updates
```javascript
function sendMessage(chat) {
  // Show message immediately (optimistic)
  chat.messages.push({
    id: Date.now(),
    content: messageText,
    sending: true
  })
  
  // Send via WebSocket
  const success = sendMessage(chat.user_id, messageText)
  
  if (!success) {
    // Remove optimistic message
    chat.messages.pop()
    showError('Failed to send')
  }
}
```

## Architecture Benefits

### 1. Separation of Concerns
- WebSocket logic isolated from UI
- Easy to test independently
- Can swap implementation without changing UI

### 2. Reusability
- Use in any component
- No prop drilling
- Consistent API

### 3. Maintainability
- Single source of truth
- Centralized error handling
- Easy to debug

### 4. Performance
- Single connection
- Efficient event system
- Automatic cleanup

### 5. Developer Experience
- Simple API
- TypeScript-ready
- Great console logs for debugging

## Debugging Tips

### Check Connection
```javascript
const { connected, wsState } = useWebSocket()
console.log('Connected:', connected.value)
console.log('Online users:', wsState.value.onlineUsers)
```

### Monitor Messages
```javascript
// In useWebSocket.js
ws.onmessage = (event) => {
  console.log('📨 Received:', event.data)  // See all messages
  // ...
}
```

### Track Events
```javascript
function emit(event, data) {
  console.log(`🔔 Event: ${event}`, data)  // See all events
  // ...
}
```

### Network Tab
- Open DevTools → Network → WS
- See WebSocket connection
- Monitor frames sent/received

## Questions & Answers

**Q: Why not use Socket.IO?**
A: Your backend uses native WebSockets (Gorilla), not Socket.IO protocol. Native is simpler and faster.

**Q: Can I use this in multiple components?**
A: Yes! It's a singleton - all components share the same connection.

**Q: What if user opens multiple tabs?**
A: Each tab creates its own connection. Backend handles multiple connections per user.

**Q: How to handle token expiration?**
A: Add token refresh logic before connecting, or reconnect with new token on 401 error.

**Q: Is this production-ready?**
A: Almost! Add WSS, better error handling, and monitoring for production.
