# Notifications Composable - Usage Guide

## Overview

The `useNotifications()` composable manages real-time notifications via WebSocket connection to your notifications service (port 8086).

## Quick Start

### 1. Basic Setup in Component

```vue
<script setup>
import { useNotifications } from '@/composables/useNotifications'

const { 
  connected, 
  notifications, 
  unreadCount,
  markAsRead,
  on 
} = useNotifications()

// Listen for new notifications
onMounted(() => {
  on('notification', (notification) => {
    console.log('New notification:', notification)
    // Show toast, play sound, etc.
  })
})
</script>

<template>
  <div class="notifications">
    <div class="badge" v-if="unreadCount > 0">
      {{ unreadCount }}
    </div>
    
    <div v-for="notif in notifications" :key="notif.id">
      {{ notif.content }}
      <button @click="markAsRead(notif.id)">Mark Read</button>
    </div>
  </div>
</template>
```

### 2. Connect on Login

```javascript
// After successful login
import { useNotifications } from '@/composables/useNotifications'

async function login(email, password) {
  const response = await api.login(email, password)
  setUser(response.user, response.token)
  
  // Connect to notifications
  const notifications = useNotifications()
  notifications.connect()
}
```

### 3. Disconnect on Logout

```javascript
function logout() {
  const notifications = useNotifications()
  notifications.disconnect()
  clearUser()
}
```

## Notification Types

Your backend sends these notification types (from `models/notification.go`):

### Follow Notifications
```javascript
{
  type: 'follow_request',
  content: 'John Doe wants to follow you',
  actor_id: 123,
  created_at: '2025-11-12T...'
}

{
  type: 'follow_accepted',
  content: 'Jane Smith accepted your follow request',
  actor_id: 456
}

{
  type: 'new_follower',
  content: 'Mike Johnson started following you',
  actor_id: 789
}
```

### Interaction Notifications
```javascript
{
  type: 'like',
  content: 'Sarah liked your post',
  actor_id: 111,
  post_id: 555
}

{
  type: 'comment',
  content: 'Tom commented on your post',
  actor_id: 222,
  post_id: 666,
  comment_id: 777
}
```

### Group/Event Notifications
```javascript
{
  type: 'group_invite',
  content: 'You were invited to Tech Nerds',
  actor_id: 333,
  group_id: 888
}

{
  type: 'event_invite',
  content: 'You were invited to Hackathon 2025',
  actor_id: 444,
  event_id: 999
}
```

### Chat Notifications
```javascript
{
  type: 'new_message',
  content: 'Alex sent you a message',
  actor_id: 555,
  message_id: 1234
}
```

## Available Methods

### connect()
Establishes WebSocket connection to notifications service.

```javascript
const { connect } = useNotifications()
connect()
```

**When to call:**
- After user logs in
- When app initializes with existing session

### disconnect()
Closes WebSocket connection.

```javascript
const { disconnect } = useNotifications()
disconnect()
```

**When to call:**
- When user logs out
- When app is closing

### markAsRead(notificationId)
Marks a single notification as read.

```javascript
const { markAsRead } = useNotifications()

function handleClick(notification) {
  markAsRead(notification.id)
  // Navigate to relevant page
  router.push(`/post/${notification.post_id}`)
}
```

**Returns:** `boolean` - Success status

### markAllAsRead()
Marks all notifications as read at once.

```javascript
const { markAllAsRead } = useNotifications()

function clearAll() {
  markAllAsRead()
}
```

**Use case:** "Mark all as read" button

### deleteNotification(notificationId)
Deletes a notification.

```javascript
const { deleteNotification } = useNotifications()

function remove(notificationId) {
  if (confirm('Delete this notification?')) {
    deleteNotification(notificationId)
  }
}
```

**Returns:** `boolean` - Success status

### clearNotifications()
Clears all notifications from local state (doesn't delete from server).

```javascript
const { clearNotifications } = useNotifications()

function onLogout() {
  clearNotifications()
  disconnect()
}
```

**Use case:** Clean up on logout

## Event System

### Available Events

#### 'notification' - New notification received
```javascript
on('notification', (notification) => {
  console.log('New notification:', notification)
  
  // Show toast notification
  showToast(notification.content)
  
  // Play sound
  playNotificationSound()
  
  // Update favicon badge
  updateFaviconBadge(unreadCount.value)
})
```

#### 'unread_count' - Unread count updated
```javascript
on('unread_count', (count) => {
  console.log('Unread count:', count)
  
  // Update page title
  document.title = count > 0 
    ? `(${count}) Neon Connex` 
    : 'Neon Connex'
})
```

#### 'notification_read' - Notification marked as read
```javascript
on('notification_read', (notificationId) => {
  console.log('Notification read:', notificationId)
  // Update UI to show as read
})
```

#### 'all_read' - All notifications marked as read
```javascript
on('all_read', () => {
  console.log('All notifications marked as read')
  // Clear all notification badges
})
```

#### 'notification_deleted' - Notification deleted
```javascript
on('notification_deleted', (notificationId) => {
  console.log('Notification deleted:', notificationId)
  // Remove from UI
})
```

#### 'connected' - WebSocket connected
```javascript
on('connected', () => {
  console.log('Notifications connected')
  // Show "Connected" indicator
})
```

#### 'disconnected' - WebSocket disconnected
```javascript
on('disconnected', ({ code, reason }) => {
  console.log('Notifications disconnected:', code, reason)
  // Show "Reconnecting..." indicator
})
```

## Complete Component Example

### Notification Dropdown Component

```vue
<template>
  <div class="notification-dropdown">
    <!-- Trigger Button -->
    <button @click="toggleDropdown" class="notification-bell">
      🔔
      <span v-if="unreadCount > 0" class="badge">
        {{ unreadCount }}
      </span>
    </button>

    <!-- Dropdown -->
    <div v-if="isOpen" class="dropdown">
      <div class="dropdown-header">
        <h3>Notifications</h3>
        <button 
          v-if="unreadCount > 0" 
          @click="handleMarkAllRead"
          class="mark-all"
        >
          Mark all read
        </button>
      </div>

      <!-- Loading State -->
      <div v-if="!connected" class="loading">
        Connecting...
      </div>

      <!-- Empty State -->
      <div v-else-if="notifications.length === 0" class="empty">
        <p>No notifications yet</p>
      </div>

      <!-- Notifications List -->
      <div v-else class="notifications-list">
        <div
          v-for="notif in notifications"
          :key="notif.id"
          :class="['notification-item', { unread: !notif.is_read }]"
          @click="handleNotificationClick(notif)"
        >
          <div class="notification-icon">
            {{ getIcon(notif.type) }}
          </div>
          
          <div class="notification-content">
            <p>{{ notif.content }}</p>
            <small>{{ formatTime(notif.created_at) }}</small>
          </div>

          <button 
            @click.stop="handleDelete(notif.id)"
            class="delete-btn"
          >
            ×
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useNotifications } from '@/composables/useNotifications'
import { useRouter } from 'vue-router'

const router = useRouter()
const isOpen = ref(false)

const {
  connected,
  notifications,
  unreadCount,
  markAsRead,
  markAllAsRead,
  deleteNotification,
  on,
  off
} = useNotifications()

function toggleDropdown() {
  isOpen.value = !isOpen.value
}

function handleNotificationClick(notification) {
  // Mark as read
  markAsRead(notification.id)
  
  // Navigate based on type
  switch (notification.type) {
    case 'follow_request':
      router.push('/follow-requests')
      break
    case 'like':
    case 'comment':
      router.push(`/post/${notification.post_id}`)
      break
    case 'group_invite':
      router.push(`/group/${notification.group_id}`)
      break
    case 'new_message':
      router.push(`/chat/${notification.actor_id}`)
      break
    default:
      router.push('/notifications')
  }
  
  isOpen.value = false
}

function handleMarkAllRead() {
  markAllAsRead()
}

function handleDelete(notificationId) {
  if (confirm('Delete this notification?')) {
    deleteNotification(notificationId)
  }
}

function getIcon(type) {
  const icons = {
    follow_request: '👤',
    follow_accepted: '✅',
    new_follower: '👥',
    like: '❤️',
    comment: '💬',
    group_invite: '👥',
    event_invite: '📅',
    new_message: '✉️'
  }
  return icons[type] || '🔔'
}

function formatTime(timestamp) {
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now - date

  if (diff < 60000) return 'Just now'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`
  if (diff < 604800000) return `${Math.floor(diff / 86400000)}d ago`
  return date.toLocaleDateString()
}

// Listen for new notifications
let handleNewNotification

onMounted(() => {
  handleNewNotification = (notification) => {
    console.log('New notification received:', notification)
    
    // Play sound
    playSound()
    
    // Show desktop notification if allowed
    // (handled automatically by composable)
  }
  
  on('notification', handleNewNotification)
})

onUnmounted(() => {
  off('notification', handleNewNotification)
})

function playSound() {
  const audio = new Audio('/notification.mp3')
  audio.volume = 0.5
  audio.play().catch(err => console.log('Could not play sound:', err))
}
</script>

<style scoped>
.notification-dropdown {
  position: relative;
}

.notification-bell {
  position: relative;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 1rem;
  width: 3.25rem;
  height: 3.25rem;
  cursor: pointer;
  font-size: 1.5rem;
  transition: all 0.2s ease;
}

.notification-bell:hover {
  transform: translateY(-2px);
  border-color: var(--border-glow);
}

.badge {
  position: absolute;
  top: -0.35rem;
  right: -0.35rem;
  background: var(--neon-pink);
  color: #05060d;
  font-size: 0.65rem;
  font-weight: 700;
  border-radius: 999px;
  padding: 0.15rem 0.45rem;
  box-shadow: 0 0 14px rgba(255, 0, 230, 0.55);
}

.dropdown {
  position: absolute;
  right: 0;
  top: 110%;
  width: 380px;
  max-height: 500px;
  background: rgba(5, 6, 13, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 1rem;
  box-shadow: var(--shadow);
  backdrop-filter: blur(20px);
  z-index: 1000;
}

.dropdown-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.dropdown-header h3 {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
}

.mark-all {
  background: transparent;
  border: 1px solid var(--neon-cyan);
  color: var(--neon-cyan);
  padding: 0.35rem 0.75rem;
  border-radius: 0.5rem;
  cursor: pointer;
  font-size: 0.85rem;
}

.notifications-list {
  max-height: 400px;
  overflow-y: auto;
}

.notification-item {
  display: flex;
  gap: 0.75rem;
  padding: 0.85rem 1.25rem;
  cursor: pointer;
  transition: background 0.2s;
  border-left: 3px solid transparent;
}

.notification-item:hover {
  background: rgba(255, 255, 255, 0.05);
}

.notification-item.unread {
  background: rgba(0, 247, 255, 0.05);
  border-left-color: var(--neon-cyan);
}

.notification-icon {
  font-size: 1.5rem;
}

.notification-content {
  flex: 1;
  min-width: 0;
}

.notification-content p {
  margin: 0 0 0.25rem;
  font-size: 0.9rem;
}

.notification-content small {
  color: var(--text-muted);
  font-size: 0.75rem;
}

.delete-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  font-size: 1.5rem;
  cursor: pointer;
  padding: 0;
  width: 1.5rem;
  height: 1.5rem;
}

.delete-btn:hover {
  color: var(--neon-pink);
}

.loading,
.empty {
  padding: 2rem;
  text-align: center;
  color: var(--text-muted);
}
</style>
```

## Browser Notifications

### Request Permission

```javascript
async function requestNotificationPermission() {
  if (!('Notification' in window)) {
    console.log('Browser does not support notifications')
    return false
  }

  if (Notification.permission === 'granted') {
    return true
  }

  if (Notification.permission !== 'denied') {
    const permission = await Notification.requestPermission()
    return permission === 'granted'
  }

  return false
}

// Call on user action (e.g., settings page)
onMounted(() => {
  requestNotificationPermission()
})
```

### Desktop notifications are automatically shown
The composable handles this internally when notifications arrive.

## Environment Variables

Create `.env` file:
```bash
# Notifications WebSocket URL
VITE_NOTIFICATIONS_WS_URL=ws://localhost:8086/ws

# For production:
# VITE_NOTIFICATIONS_WS_URL=wss://api.yoursite.com/notifications/ws
```

## Integration with App.vue

```vue
<script setup>
import { onMounted } from 'vue'
import { useWebSocket } from './composables/useWebSocket'
import { useNotifications } from './composables/useNotifications'
import { isAuthenticated } from './stores/auth'

onMounted(() => {
  // If user is already logged in (from localStorage)
  if (isAuthenticated()) {
    // Connect both WebSockets
    const chat = useWebSocket()
    const notifications = useNotifications()
    
    chat.connect()
    notifications.connect()
  }
})
</script>
```

## Best Practices

### 1. Connect After Login
```javascript
async function handleLogin() {
  // Login first
  await login(email, password)
  
  // Then connect WebSockets
  const notifications = useNotifications()
  notifications.connect()
}
```

### 2. Disconnect on Logout
```javascript
function handleLogout() {
  const notifications = useNotifications()
  
  // Disconnect and clear
  notifications.disconnect()
  notifications.clearNotifications()
  
  // Clear auth
  clearUser()
}
```

### 3. Handle Reconnection
The composable automatically reconnects on connection loss. No action needed!

### 4. Clean Up Listeners
```javascript
onUnmounted(() => {
  off('notification', handleNotification)
})
```

### 5. Show Connection Status
```vue
<div v-if="!connected" class="connection-status">
  <span v-if="connecting">Connecting...</span>
  <span v-else>Disconnected</span>
</div>
```

## Troubleshooting

### No notifications appearing
1. Check if WebSocket is connected: `connected.value`
2. Check browser console for errors
3. Verify backend is running on port 8086
4. Check token is valid

### Desktop notifications not showing
1. Check permission: `Notification.permission`
2. Request permission on user action
3. Check browser supports notifications

### Unread count not updating
1. Verify `unread_count` event is emitted by backend
2. Check `requestUnreadCount()` is called on connect
3. Look for errors in console

### Connection keeps dropping
1. Check heartbeat is working (ping/pong)
2. Verify backend timeout settings
3. Check network connection
