/**
 * WebSocket Composable for Chat
 * 
 * WHY SEPARATE CONNECTIONS:
 * Your backend has TWO separate WebSocket servers:
 * 1. Chat Service (port 8085) - handles chat messages
 * 2. Notification Service (port 8086) - handles notifications
 * 
 * This composable manages the CHAT WebSocket connection.
 * We'll create useNotifications separately for notifications.
 */

import { ref, computed, onUnmounted } from 'vue'
import { getUser, getToken } from '../stores/auth'

// WebSocket connection instance (null when disconnected)
const ws = ref(null)

// Connection state
const connected = ref(false)
const connecting = ref(false)
const reconnectAttempts = ref(0)

// Online users tracking (Set for O(1) lookups)
const onlineUsers = ref(new Set())

// Event listeners registry
// WHY: Allow components to subscribe to specific events (message, typing, etc.)
const eventListeners = ref(new Map())

/**
 * WHY CONFIGURATION OBJECT:
 * - Makes it easy to switch environments (dev/prod)
 * - Centralized URL management
 * - Can be overridden via env variables
 */
const config = {
  chatUrl: import.meta.env.VITE_CHAT_WS_URL || 'ws://localhost:8085/ws',
  reconnectDelay: 3000,
  maxReconnectAttempts: 5,
  heartbeatInterval: 30000 // 30 seconds
}

/**
 * WHY THIS COMPOSABLE PATTERN:
 * - Reusable across multiple components
 * - Maintains single WebSocket connection (singleton)
 * - Automatic reconnection logic
 * - Event-driven architecture (pub/sub pattern)
 */
export function useWebSocket() {
  // Heartbeat timer to detect stale connections
  let heartbeatTimer = null
  let reconnectTimer = null

  /**
   * Connect to the Chat WebSocket server
   * 
   * WHY AUTHENTICATION IN URL:
   * Your backend expects: /ws?token=xxx
   * This authenticates the WebSocket connection
   */
  function connect() {
    // Prevent multiple simultaneous connections
    if (ws.value || connecting.value) {
      console.log('WebSocket already connected or connecting')
      return
    }

    const user = getUser()
    const token = getToken()

    // Must be authenticated to connect
    if (!user || !token) {
      console.error('Cannot connect: User not authenticated')
      return
    }

    connecting.value = true
    console.log('🔌 Connecting to Chat WebSocket...')

    try {
      // WHY QUERY PARAMS:
      // Your backend middleware checks: GetUserIDFromContext(r)
      // The token is sent as query param for WebSocket auth
      const wsUrl = `${config.chatUrl}?token=${encodeURIComponent(token)}&username=${encodeURIComponent(user.username)}`
      
      ws.value = new WebSocket(wsUrl)

      // WHY ONOPEN:
      // Confirms connection is established
      // Reset reconnection attempts
      ws.value.onopen = () => {
        console.log('✅ Chat WebSocket connected')
        connected.value = true
        connecting.value = false
        reconnectAttempts.value = 0
        
        // Start heartbeat to keep connection alive
        startHeartbeat()
        
        // Emit connected event to listeners
        emit('connected', { userId: user.id })
      }

      // WHY ONMESSAGE:
      // Receives ALL messages from server
      // Routes them to appropriate event handlers
      ws.value.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          console.log('📨 WebSocket message received:', data)
          
          handleIncomingMessage(data)
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
        }
      }

      // WHY ONERROR:
      // Logs connection errors
      // Useful for debugging connection issues
      ws.value.onerror = (error) => {
        console.error('❌ Chat WebSocket error:', error)
        connecting.value = false
      }

      // WHY ONCLOSE:
      // Detects when connection is lost
      // Triggers automatic reconnection
      ws.value.onclose = (event) => {
        console.log('🔌 Chat WebSocket disconnected:', event.code, event.reason)
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
      console.error('Failed to create WebSocket connection:', error)
      connecting.value = false
    }
  }

  /**
   * Disconnect from WebSocket
   * 
   * WHY EXPLICIT DISCONNECT:
   * - Clean shutdown on logout
   * - Prevents reconnection attempts
   */
  function disconnect() {
    console.log('🔌 Disconnecting Chat WebSocket...')
    
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
   * - Prevents hammering server with reconnect attempts
   * - Gradually increases delay between attempts
   */
  function scheduleReconnect() {
    reconnectAttempts.value++
    const delay = config.reconnectDelay * reconnectAttempts.value
    
    console.log(`🔄 Reconnecting in ${delay}ms (attempt ${reconnectAttempts.value}/${config.maxReconnectAttempts})`)
    
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
   * - Your backend expects ping/pong messages
   */
  function startHeartbeat() {
    stopHeartbeat() // Clear any existing timer
    
    heartbeatTimer = setInterval(() => {
      if (ws.value && connected.value) {
        try {
          // Send ping message
          ws.value.send(JSON.stringify({ type: 'ping' }))
        } catch (error) {
          console.error('Heartbeat failed:', error)
          disconnect()
          connect()
        }
      }
    }, config.heartbeatInterval)
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
   * Handle incoming messages from server
   * 
   * WHY MESSAGE ROUTER:
   * Your backend sends different message types:
   * - "message" - new chat message
   * - "group_message" - group chat message
   * - "typing" - typing indicator
   * - "error" - error message
   * - "online_status" - user online/offline
   */
  function handleIncomingMessage(data) {
    const { type } = data

    switch (type) {
      case 'message':
        // New 1-on-1 chat message
        emit('message', data)
        break

      case 'group_message':
        // New group chat message
        emit('group_message', data)
        break

      case 'typing':
        // Someone is typing
        emit('typing', data)
        break

      case 'read':
        // Message read receipt
        emit('read', data)
        break

      case 'online_status':
        // User online/offline status update
        handleOnlineStatus(data)
        break

      case 'error':
        // Server error message
        console.error('Server error:', data.error)
        emit('error', data)
        break

      case 'pong':
        // Heartbeat response (ignore)
        break

      default:
        console.warn('Unknown message type:', type)
    }
  }

  /**
   * Handle online status updates
   * 
   * WHY TRACK ONLINE USERS:
   * - Show green dot next to online contacts
   * - Used in chat UI to show "Active now"
   */
  function handleOnlineStatus(data) {
    const { user_id, is_online } = data
    
    if (is_online) {
      onlineUsers.value.add(user_id)
    } else {
      onlineUsers.value.delete(user_id)
    }
    
    emit('online_status', data)
  }

  /**
   * Send a chat message
   * 
   * WHY RETURN BOOLEAN:
   * - Caller knows if send succeeded
   * - Can show error UI if failed
   * 
   * @param {number} receiverId - User ID to send message to
   * @param {string} content - Message content
   * @returns {boolean} - Success status
   */
  function sendMessage(receiverId, content) {
    if (!ws.value || !connected.value) {
      console.error('Cannot send message: not connected')
      return false
    }

    try {
      const message = {
        type: 'message',
        receiver_id: receiverId,
        content: content,
        timestamp: new Date().toISOString()
      }

      ws.value.send(JSON.stringify(message))
      console.log('✉️ Message sent:', message)
      return true
    } catch (error) {
      console.error('Failed to send message:', error)
      return false
    }
  }

  /**
   * Send typing indicator
   * 
   * @param {number} receiverId - User ID being typed to
   */
  function sendTyping(receiverId) {
    if (!ws.value || !connected.value) return

    try {
      ws.value.send(JSON.stringify({
        type: 'typing',
        receiver_id: receiverId
      }))
    } catch (error) {
      console.error('Failed to send typing indicator:', error)
    }
  }

  /**
   * Register event listener
   * 
   * WHY EVENT SYSTEM:
   * - Decouples message handling from WebSocket logic
   * - Multiple components can listen to same events
   * - Easy to add/remove listeners
   * 
   * @param {string} event - Event name ('message', 'typing', etc.)
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
   * Cleanup on component unmount
   * 
   * WHY ONUNMOUNTED:
   * - Vue lifecycle hook
   * - Prevents memory leaks
   * - Cleans up timers
   */
  onUnmounted(() => {
    // Note: We DON'T disconnect here
    // WebSocket should persist across component mounts
    // Only disconnect on logout
  })

  // Expose public API
  return {
    // State
    connected: computed(() => connected.value),
    connecting: computed(() => connecting.value),
    
    // WebSocket state for components
    wsState: computed(() => ({
      connected: connected.value,
      connecting: connecting.value,
      onlineUsers: onlineUsers.value
    })),
    
    // Methods
    connect,
    disconnect,
    sendMessage,
    sendTyping,
    on,
    off
  }
}
