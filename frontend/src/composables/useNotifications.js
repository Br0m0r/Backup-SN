/**
 * WebSocket Composable for Notifications
 * 
 * WHY SEPARATE FROM CHAT:
 * Your backend has TWO WebSocket servers:
 * 1. Chat Service (port 8085) - handled by useWebSocket.js
 * 2. Notification Service (port 8086) - THIS FILE
 * 
 * Notifications include:
 * - Follow requests
 * - Follow accepts
 * - New followers
 * - Likes on posts
 * - Comments on posts
 * - Group invites
 * - Event notifications
 * - New messages (when chat is closed)
 */

import { ref, computed, onUnmounted } from 'vue'
import { getUser, getToken } from '../stores/auth'

// WebSocket connection instance (null when disconnected)
const ws = ref(null)

// Connection state
const connected = ref(false)
const connecting = ref(false)
const reconnectAttempts = ref(0)

// Notifications state
const notifications = ref([])
const unreadCount = ref(0)

// Event listeners registry
const eventListeners = ref(new Map())

/**
 * Configuration for notifications WebSocket
 * 
 * WHY DIFFERENT PORT:
 * - Notifications service runs on :8086
 * - Chat service runs on :8085
 * - Keeps concerns separated
 */
const config = {
  notificationsUrl: import.meta.env.VITE_NOTIFICATIONS_WS_URL || 'ws://localhost:8086/ws',
  notificationsApiUrl: import.meta.env.VITE_NOTIFICATIONS_API_URL || 'http://localhost:8086',
  reconnectDelay: 3000,
  maxReconnectAttempts: 5,
  heartbeatInterval: 30000 // 30 seconds
}

/**
 * WHY THIS COMPOSABLE:
 * - Real-time notification delivery
 * - No polling required
 * - Updates UI instantly when events occur
 * - Maintains persistent connection
 * - Separate from chat to prevent blocking
 */
export function useNotifications() {
  // Heartbeat timer to detect stale connections
  let heartbeatTimer = null
  let reconnectTimer = null

  /**
   * Connect to the Notifications WebSocket server
   * 
   * WHY AUTHENTICATION:
   * Same as chat - backend requires authenticated user ID
   * to know where to send notifications
   */
  function connect() {
    // Prevent multiple simultaneous connections
    if (ws.value || connecting.value) {
      console.log('Notifications WebSocket already connected or connecting')
      return
    }

    const user = getUser()
    const token = getToken()

    // Must be authenticated to connect
    if (!user || !token) {
      console.error('Cannot connect notifications: User not authenticated')
      return
    }

    connecting.value = true
    console.log('🔔 Connecting to Notifications WebSocket...')

    try {
      // WHY QUERY PARAMS:
      // Your backend: middleware.GetUserIDFromContext(r)
      // Token validates and provides user ID
      const wsUrl = `${config.notificationsUrl}?token=${encodeURIComponent(token)}`
      
      ws.value = new WebSocket(wsUrl)

      /**
       * WHY ONOPEN:
       * Connection established successfully
       * Reset reconnection counter
       * Start keep-alive heartbeat
       */
      ws.value.onopen = () => {
        console.log('✅ Notifications WebSocket connected')
        connected.value = true
        connecting.value = false
        reconnectAttempts.value = 0
        
        // Start heartbeat to keep connection alive
        startHeartbeat()
        
        // Emit connected event to listeners
        emit('connected', { userId: user.id })
        
        // Load initial notifications from REST API
        loadNotifications()
        
        // Request initial unread count from REST API
        requestUnreadCount()
      }

      /**
       * WHY ONMESSAGE:
       * Receives ALL notifications from server
       * Routes to appropriate handlers based on type
       */
      ws.value.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          console.log('🔔 Notification received:', data)
          
          handleIncomingNotification(data)
        } catch (error) {
          console.error('Failed to parse notification message:', error)
        }
      }

      /**
       * WHY ONERROR:
       * Logs connection errors
       * Helps debug connection issues
       */
      ws.value.onerror = (error) => {
        console.error('❌ Notifications WebSocket error:', error)
        connecting.value = false
      }

      /**
       * WHY ONCLOSE:
       * Detects connection loss
       * Triggers automatic reconnection
       * Cleans up resources
       */
      ws.value.onclose = (event) => {
        console.log('🔔 Notifications WebSocket disconnected:', event.code, event.reason)
        connected.value = false
        connecting.value = false
        ws.value = null
        
        stopHeartbeat()
        
        // Attempt to reconnect if not a normal closure
        if (event.code !== 1000 && reconnectAttempts.value < config.maxReconnectAttempts) {
          scheduleReconnect()
        }
        
        emit('disconnected', { code: event.code, reason: event.reason })
      }

    } catch (error) {
      console.error('Failed to create notifications WebSocket:', error)
      connecting.value = false
    }
  }

  /**
   * Disconnect from WebSocket
   * 
   * WHY EXPLICIT DISCONNECT:
   * - Clean shutdown on logout
   * - Prevents reconnection attempts
   * - Frees resources
   */
  function disconnect() {
    console.log('🔔 Disconnecting Notifications WebSocket...')
    
    // Clear timers
    stopHeartbeat()
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    
    // Close connection with normal closure code
    if (ws.value) {
      ws.value.close(1000, 'Client disconnect')
      ws.value = null
    }
    
    connected.value = false
    connecting.value = false
    reconnectAttempts.value = 0
  }

  /**
   * Schedule reconnection attempt
   * 
   * WHY EXPONENTIAL BACKOFF:
   * - Prevents hammering server
   * - Gradually increases delay
   * - Gives server time to recover
   */
  function scheduleReconnect() {
    reconnectAttempts.value++
    const delay = config.reconnectDelay * reconnectAttempts.value
    
    console.log(`🔄 Reconnecting notifications in ${delay}ms (attempt ${reconnectAttempts.value}/${config.maxReconnectAttempts})`)
    
    reconnectTimer = setTimeout(() => {
      connect()
    }, delay)
  }

  /**
   * Start heartbeat to keep connection alive
   * 
   * WHY HEARTBEAT:
   * - Detects stale connections
   * - Prevents timeout on idle connections
   * - Backend expects ping/pong at protocol level
   */
  function startHeartbeat() {
    stopHeartbeat() // Clear any existing timer
    
    // Backend handles ping/pong automatically at WebSocket protocol level
    // No need to send manual ping messages
  }

  /**
   * Stop heartbeat timer
   */
  function stopHeartbeat() {
    if (heartbeatTimer) {
      clearInterval(heartbeatTimer)
      heartbeatTimer = null
    }
  }

  /**
   * Handle incoming notifications from server
   * 
   * WHY MESSAGE ROUTER:
   * Your backend only sends notification events with this format:
   * - {"type": "notification", "notification": {...}}
   */
  function handleIncomingNotification(data) {
    const { type } = data

    if (type === 'notification') {
      // New notification arrived
      handleNewNotification(data.notification)
    } else {
      console.warn('Unknown notification type:', type)
    }
  }

  /**
   * Handle new notification
   * 
   * WHY ADD TO LOCAL STATE:
   * - Show in notification dropdown immediately
   * - Update badge count
   * - Play sound/show desktop notification
   * - No need to refetch from API
   */
  function handleNewNotification(notification) {
    // Add to notifications array
    notifications.value.unshift(notification)
    
    // Increment unread count if not already read
    if (!notification.is_read) {
      unreadCount.value++
    }
    
    // Emit to listeners (for UI updates, sounds, etc.)
    emit('notification', notification)
    
    // Show browser notification if permitted
    showBrowserNotification(notification)
  }

  /**
   * Show browser desktop notification
   * 
   * WHY BROWSER NOTIFICATIONS:
   * - User sees notifications even when tab is in background
   * - Better UX for important events
   * - Standard web feature
   */
  function showBrowserNotification(notification) {
    // Check if browser supports notifications
    if (!('Notification' in window)) {
      return
    }

    // Check permission
    if (Notification.permission === 'granted') {
      const title = getNotificationTitle(notification)
      const body = notification.content
      
      new Notification(title, {
        body: body,
        icon: '/logo.png', // Your app logo
        badge: '/badge.png',
        tag: `notification-${notification.id}`, // Prevent duplicates
        requireInteraction: false,
        silent: false
      })
    } else if (Notification.permission !== 'denied') {
      // Request permission if not denied
      Notification.requestPermission().then(permission => {
        if (permission === 'granted') {
          showBrowserNotification(notification)
        }
      })
    }
  }

  /**
   * Get human-readable title for notification
   * 
   * WHY DIFFERENT TITLES:
   * - Each notification type has different meaning
   * - Makes it clear what happened
   * - Better UX
   */
  function getNotificationTitle(notification) {
    const typeMap = {
      'follow_request': '👤 New Follow Request',
      'follow_accepted': '✅ Follow Request Accepted',
      'new_follower': '👥 New Follower',
      'like': '❤️ New Like',
      'comment': '💬 New Comment',
      'group_invite': '👥 Group Invitation',
      'event_invite': '📅 Event Invitation',
      'new_message': '✉️ New Message'
    }
    
    return typeMap[notification.type] || '🔔 New Notification'
  }

  /**
   * Handle notification marked as read
   * 
   * WHY UPDATE LOCAL STATE:
   * - Keep UI in sync with server
   * - Decrement unread count
   * - Update notification appearance
   */
  function handleMarkAsRead(notificationId) {
    const notification = notifications.value.find(n => n.id === notificationId)
    
    if (notification && !notification.is_read) {
      notification.is_read = true
      unreadCount.value = Math.max(0, unreadCount.value - 1)
      
      emit('notification_read', notificationId)
    }
  }

  /**
   * Request unread count from server
   * 
   * WHY REQUEST COUNT:
   * - Get accurate count on connect
   * - Sync with server state
   * - Update badge immediately
   */
  function requestUnreadCount() {
    if (!ws.value || !connected.value) return

    try {
      ws.value.send(JSON.stringify({
        type: 'get_unread_count'
      }))
    } catch (error) {
      console.error('Failed to request unread count:', error)
    }
  }

  /**
   * Mark notification as read (via REST API)
   * 
   * @param {number} notificationId - Notification ID to mark as read
   */
  function markAsRead(notificationId) {
    // Optimistically update local state
    const notification = notifications.value.find(n => n.id === notificationId)
    if (notification && !notification.is_read) {
      notification.is_read = true
      unreadCount.value = Math.max(0, unreadCount.value - 1)
    }

    // Make REST API call (don't send via WebSocket)
    const token = getToken()
    if (!token) return false

    fetch(`${config.notificationsApiUrl}/notifications/read/${notificationId}`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    }).catch(error => {
      console.error('Failed to mark notification as read:', error)
      // Revert on error
      if (notification) {
        notification.is_read = false
        unreadCount.value++
      }
    })

    return true
  }

  /**
   * Mark all notifications as read (via REST API)
   */
  function markAllAsRead() {
    // Optimistically update local state
    notifications.value.forEach(n => {
      n.is_read = true
    })
    unreadCount.value = 0

    // Make REST API call
    const token = getToken()
    if (!token) return false

    fetch(`${config.notificationsApiUrl}/notifications/read-all`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    }).catch(error => {
      console.error('Failed to mark all as read:', error)
    })

    emit('all_read')
    return true
  }

  /**
   * Delete notification (via REST API)
   * 
   * @param {number} notificationId - Notification ID to delete
   */
  function deleteNotification(notificationId) {
    // Optimistically remove from local state
    const index = notifications.value.findIndex(n => n.id === notificationId)
    let removedNotification = null
    if (index !== -1) {
      removedNotification = notifications.value[index]
      notifications.value.splice(index, 1)
      
      // Decrement unread count if it was unread
      if (!removedNotification.is_read) {
        unreadCount.value = Math.max(0, unreadCount.value - 1)
      }
    }

    // Make REST API call
    const token = getToken()
    if (!token) return false

    fetch(`${config.notificationsApiUrl}/notifications/delete/${notificationId}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    }).catch(error => {
      console.error('Failed to delete notification:', error)
      // Revert on error
      if (removedNotification && index !== -1) {
        notifications.value.splice(index, 0, removedNotification)
        if (!removedNotification.is_read) {
          unreadCount.value++
        }
      }
    })

    emit('notification_deleted', notificationId)
    return true
  }

  /**
   * Load notifications from REST API
   */
  async function loadNotifications(limit = 20, offset = 0) {
    const token = getToken()
    if (!token) return

    try {
      const response = await fetch(
        `${config.notificationsApiUrl}/notifications/list?limit=${limit}&offset=${offset}`,
        {
          headers: {
            'Authorization': `Bearer ${token}`
          }
        }
      )

      if (response.ok) {
        const data = await response.json()
        if (data.success && data.data.notifications) {
          notifications.value = data.data.notifications
          
          // Calculate unread count
          unreadCount.value = notifications.value.filter(n => !n.is_read).length
        }
      }
    } catch (error) {
      console.error('Failed to load notifications:', error)
    }
  }

  /**
   * Request unread count from REST API
   */
  async function requestUnreadCount() {
    const token = getToken()
    if (!token) return

    try {
      const response = await fetch(
        `${config.notificationsApiUrl}/notifications/unread-count`,
        {
          headers: {
            'Authorization': `Bearer ${token}`
          }
        }
      )

      if (response.ok) {
        const data = await response.json()
        if (data.success && typeof data.data.unread_count === 'number') {
          unreadCount.value = data.data.unread_count
        }
      }
    } catch (error) {
      console.error('Failed to get unread count:', error)
    }
  }

  /**
   * Register event listener
   * 
   * WHY EVENT SYSTEM:
   * - Decouples notification handling from WebSocket logic
   * - Multiple components can listen to same events
   * - Easy to add/remove listeners
   * 
   * Available events:
   * - 'notification' - New notification received
   * - 'unread_count' - Unread count updated
   * - 'notification_read' - Notification marked as read
   * - 'all_read' - All notifications marked as read
   * - 'notification_deleted' - Notification deleted
   * - 'connected' - WebSocket connected
   * - 'disconnected' - WebSocket disconnected
   * 
   * @param {string} event - Event name
   * @param {Function} callback - Handler function
   */
  function on(event, callback) {
    if (!eventListeners.value.has(event)) {
      eventListeners.value.set(event, [])
    }
    eventListeners.value.get(event).push(callback)
  }

  /**
   * Unregister event listener
   * 
   * WHY CLEANUP:
   * - Prevent memory leaks
   * - Remove listeners when component unmounts
   * 
   * @param {string} event - Event name
   * @param {Function} callback - Handler function to remove
   */
  function off(event, callback) {
    if (!eventListeners.value.has(event)) return
    
    const listeners = eventListeners.value.get(event)
    const index = listeners.indexOf(callback)
    if (index > -1) {
      listeners.splice(index, 1)
    }
  }

  /**
   * Emit event to all registered listeners
   * 
   * @param {string} event - Event name
   * @param {any} data - Event data
   */
  function emit(event, data) {
    if (!eventListeners.value.has(event)) return
    
    const listeners = eventListeners.value.get(event)
    listeners.forEach(callback => {
      try {
        callback(data)
      } catch (error) {
        console.error(`Error in ${event} listener:`, error)
      }
    })
  }

  /**
   * Clear all notifications from local state
   * 
   * WHY CLEAR:
   * - Free memory
   * - Clean slate on logout
   * - Can refetch fresh notifications
   */
  function clearNotifications() {
    notifications.value = []
    unreadCount.value = 0
  }

  /**
   * Cleanup on component unmount
   * 
   * WHY ONUNMOUNTED:
   * - Vue lifecycle hook
   * - Note: We DON'T disconnect here
   * - WebSocket persists across component mounts
   */
  onUnmounted(() => {
    // WebSocket connection should persist
    // Only disconnect on logout
  })

  // Expose public API
  return {
    // State
    connected: computed(() => connected.value),
    connecting: computed(() => connecting.value),
    notifications: computed(() => notifications.value),
    unreadCount: computed(() => unreadCount.value),
    
    // Notification state for components
    notificationState: computed(() => ({
      connected: connected.value,
      connecting: connecting.value,
      notifications: notifications.value,
      unreadCount: unreadCount.value
    })),
    
    // Methods
    connect,
    disconnect,
    markAsRead,
    markAllAsRead,
    deleteNotification,
    loadNotifications,
    clearNotifications,
    on,
    off
  }
}
