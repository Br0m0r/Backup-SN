<template>
  <div class="chat">
    <!-- Chat Sidebar (Contact List) -->
    <div class="chat-sidebar" :class="{ 'sidebar-collapsed': !showSidebar }">
      <div class="sidebar-header">
        <h3>Chat</h3>
        <button @click="toggleSidebar" class="toggle-btn">
          {{ showSidebar ? '−' : '+' }}
        </button>
      </div>

      <div v-if="showSidebar" class="sidebar-content">
        <!-- Search Box -->
        <div class="search-box">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search contacts..."
            class="search-input"
          />
        </div>

        <!-- Contacts List -->
        <div class="contacts-list">
          <div v-if="loading" class="loading-state">
            <p>Loading contacts...</p>
          </div>

          <div v-else-if="filteredContacts.length === 0" class="empty-state">
            <p>{{ searchQuery ? 'No contacts found' : 'No contacts available' }}</p>
            <small>Follow someone to start chatting!</small>
          </div>

          <div
            v-for="contact in filteredContacts"
            :key="contact.user_id"
            class="contact-item"
            :class="{ 'active': isContactActive(contact.user_id) }"
            @click="openChat(contact)"
          >
            <div class="contact-avatar">
              <div class="avatar-circle" :style="{ backgroundColor: getAvatarColor(contact.username) }">
                {{ getInitials(contact) }}
              </div>
              <div v-if="contact.is_online" class="online-indicator"></div>
            </div>

            <div class="contact-info">
              <div class="contact-name">
                {{ getDisplayName(contact) }}
              </div>
              <div class="contact-status">
                {{ contact.is_online ? 'Active now' : 'Offline' }}
              </div>
            </div>

            <div v-if="contact.unread_count > 0" class="unread-badge">
              {{ contact.unread_count }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Open Chat Windows -->
    <div class="chat-windows">
      <div
        v-for="chat in openChats"
        :key="chat.user_id"
        class="chat-window"
        :class="{ 'minimized': chat.minimized }"
      >
        <!-- Chat Header -->
        <div class="chat-header" @click="toggleMinimize(chat.user_id)">
          <div class="chat-header-left">
            <div class="chat-avatar">
              <div class="avatar-circle small" :style="{ backgroundColor: getAvatarColor(chat.username) }">
                {{ getInitials(chat) }}
              </div>
              <div v-if="chat.is_online" class="online-indicator small"></div>
            </div>
            <div class="chat-title">
              <span class="chat-name">{{ getDisplayName(chat) }}</span>
              <span class="chat-status">{{ chat.is_online ? 'Active' : 'Offline' }}</span>
            </div>
          </div>
          <div class="chat-header-actions">
            <button @click.stop="toggleMinimize(chat.user_id)" class="header-btn">
              {{ chat.minimized ? '□' : '_' }}
            </button>
            <button @click.stop="closeChat(chat.user_id)" class="header-btn close-btn">
              ×
            </button>
          </div>
        </div>

        <!-- Chat Body -->
        <div v-if="!chat.minimized" class="chat-body" ref="chatBodies">
          <div class="messages-container">
            <div v-if="chat.loadingHistory" class="loading-messages">
              Loading messages...
            </div>

            <template v-else>
              <div
                v-for="msg in chat.messages"
                :key="msg.id || msg.timestamp"
                class="message"
                :class="{ 'message-sent': msg.sender_id === currentUserId, 'message-received': msg.sender_id !== currentUserId }"
              >
                <div class="message-bubble">
                  <p>{{ msg.content }}</p>
                  <span class="message-time">{{ formatMessageTime(msg.created_at || msg.timestamp) }}</span>
                </div>
              </div>

              <div v-if="chat.messages.length === 0" class="no-messages">
                <p>No messages yet. Say hi! 👋</p>
              </div>
            </template>
          </div>
        </div>

        <!-- Chat Footer (Input) -->
        <div v-if="!chat.minimized" class="chat-footer">
          <input
            v-model="chat.messageInput"
            type="text"
            placeholder="Type a message..."
            @keyup.enter="sendMessage(chat)"
            @input="handleTyping(chat)"
            class="message-input"
          />
          <button @click="sendMessage(chat)" class="send-btn" :disabled="!chat.messageInput.trim()">
            ➤
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch, nextTick } from 'vue'
import { useWebSocket } from '../composables/useWebSocket'
import { getUser, getToken } from '../stores/auth'
import axios from 'axios'

const CHAT_API_URL = import.meta.env.VITE_CHAT_API_URL || 'http://localhost:8085'

const { connected, sendMessage: wsSendMessage, on, wsState, connect, disconnect } = useWebSocket()

const showSidebar = ref(true)
const contacts = ref([])
const openChats = ref([])
const searchQuery = ref('')
const loading = ref(false)
const currentUserId = computed(() => getUser()?.id)
const chatBodies = ref([])

// Helper to create authenticated axios instance
const createAuthClient = () => {
  const token = getToken()
  return axios.create({
    baseURL: CHAT_API_URL,
    headers: {
      Authorization: `Bearer ${token}`
    }
  })
}

// Computed
const filteredContacts = computed(() => {
  if (!searchQuery.value) return contacts.value

  const query = searchQuery.value.toLowerCase()
  return contacts.value.filter(contact => {
    const name = getDisplayName(contact).toLowerCase()
    const username = contact.username.toLowerCase()
    return name.includes(query) || username.includes(query)
  })
})

// Methods
async function loadContacts() {
  const token = getToken()
  if (!token) {
    console.warn('No auth token available, skipping contact load')
    return
  }

  loading.value = true
  try {
    const client = createAuthClient()
    const response = await client.get('/chat/contacts')
    if (response.data.success) {
      contacts.value = response.data.data.contacts || []
    }
  } catch (error) {
    console.error('Error loading contacts:', error)
    // Silently fail - chat feature is optional
    contacts.value = []
  } finally {
    loading.value = false
  }
}

function toggleSidebar() {
  showSidebar.value = !showSidebar.value
}

function isContactActive(userId) {
  return openChats.value.some(chat => chat.user_id === userId)
}

async function openChat(contact) {
  // Check if already open
  const existing = openChats.value.find(chat => chat.user_id === contact.user_id)
  if (existing) {
    existing.minimized = false
    await nextTick()
    scrollToBottom(contact.user_id)
    return
  }

  // Limit to 3 open chats
  if (openChats.value.length >= 3) {
    openChats.value.shift()
  }

  const newChat = reactive({
    user_id: contact.user_id,
    username: contact.username,
    first_name: contact.first_name,
    last_name: contact.last_name,
    nickname: contact.nickname,
    avatar: contact.avatar,
    is_online: contact.is_online,
    messages: [],
    messageInput: '',
    minimized: false,
    loadingHistory: true
  })

  openChats.value.push(newChat)

  // Load chat history
  await loadChatHistory(newChat)
  await nextTick()
  scrollToBottom(contact.user_id)
  
  // Mark as read
  markAsRead(contact.user_id)
  // Update contact unread count
  contact.unread_count = 0
}

async function loadChatHistory(chat) {
  chat.loadingHistory = true
  try {
    const client = createAuthClient()
    const response = await client.get(`/chat/history/${chat.user_id}?limit=50`)
    if (response.data.success) {
      chat.messages = response.data.data.messages || []
      console.log(`Loaded ${chat.messages.length} messages for user ${chat.user_id}`)
    }
  } catch (error) {
    console.error('Error loading chat history:', error)
    chat.messages = []
  }
  
  // Ensure reactivity by using nextTick
  await nextTick()
  chat.loadingHistory = false
  console.log(`Loading complete for user ${chat.user_id}, loadingHistory: ${chat.loadingHistory}`)
}

function closeChat(userId) {
  const index = openChats.value.findIndex(chat => chat.user_id === userId)
  if (index !== -1) {
    openChats.value.splice(index, 1)
  }
}

function toggleMinimize(userId) {
  const chat = openChats.value.find(c => c.user_id === userId)
  if (chat) {
    chat.minimized = !chat.minimized
    if (!chat.minimized) {
      nextTick(() => scrollToBottom(userId))
    }
  }
}

async function sendMessage(chat) {
  if (!chat.messageInput.trim()) return

  const content = chat.messageInput.trim()
  chat.messageInput = ''

  // Send via WebSocket
  const success = wsSendMessage(chat.user_id, content)
  
  if (success) {
    // Optimistically add to UI (will be confirmed via WebSocket)
    const optimisticMsg = {
      id: Date.now(),
      sender_id: currentUserId.value,
      receiver_id: chat.user_id,
      content: content,
      timestamp: new Date().toISOString(),
      is_read: false
    }
    chat.messages.push(optimisticMsg)
    await nextTick()
    scrollToBottom(chat.user_id)
  }
}

function handleTyping(chat) {
  // TODO: Implement typing indicator
}

async function markAsRead(userId) {
  try {
    const client = createAuthClient()
    await client.post(`/chat/read/${userId}`)
  } catch (error) {
    console.error('Error marking as read:', error)
  }
}

function scrollToBottom(userId) {
  const chat = openChats.value.find(c => c.user_id === userId)
  if (!chat) return

  const index = openChats.value.indexOf(chat)
  if (chatBodies.value && chatBodies.value[index]) {
    const container = chatBodies.value[index].querySelector('.messages-container')
    if (container) {
      container.scrollTop = container.scrollHeight
    }
  }
}

function getDisplayName(contact) {
  if (contact.nickname) return contact.nickname
  if (contact.first_name && contact.last_name) {
    return `${contact.first_name} ${contact.last_name}`
  }
  if (contact.first_name) return contact.first_name
  return contact.username
}

function getInitials(contact) {
  if (contact.first_name && contact.last_name) {
    return `${contact.first_name[0]}${contact.last_name[0]}`.toUpperCase()
  }
  if (contact.first_name) return contact.first_name[0].toUpperCase()
  return contact.username.substring(0, 2).toUpperCase()
}

function getAvatarColor(username) {
  const colors = ['#3498db', '#e74c3c', '#2ecc71', '#f39c12', '#9b59b6', '#1abc9c', '#e67e22']
  let hash = 0
  for (let i = 0; i < username.length; i++) {
    hash = username.charCodeAt(i) + ((hash << 5) - hash)
  }
  return colors[Math.abs(hash) % colors.length]
}

function formatMessageTime(timestamp) {
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now - date

  if (diff < 60000) return 'Just now'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`
  if (diff < 86400000) return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  return date.toLocaleDateString()
}

// Listen for incoming messages
onMounted(() => {
  // Connect to WebSocket for real-time chat
  connect()
  
  // Load available contacts
  loadContacts()

  // Listen for new messages
  on('message', (data) => {
    const chat = openChats.value.find(c => c.user_id === data.sender_id)
    if (chat) {
      // Check if message already exists (optimistic update)
      const exists = chat.messages.some(m => m.id === data.message_id)
      if (!exists) {
        chat.messages.push({
          id: data.message_id,
          sender_id: data.sender_id,
          receiver_id: data.receiver_id,
          content: data.content,
          created_at: data.timestamp,
          is_read: false
        })
        nextTick(() => scrollToBottom(data.sender_id))
      }
      markAsRead(data.sender_id)
    } else {
      // Update contact unread count
      const contact = contacts.value.find(c => c.user_id === data.sender_id)
      if (contact) {
        contact.unread_count = (contact.unread_count || 0) + 1
      }
    }
  })

  // Update online status
  on('connected', () => {
    loadContacts() // Refresh to get updated online status
  })

  // Listen for follow acceptances to refresh contacts
  window.addEventListener('follow-accepted', () => {
    console.log('Follow accepted, refreshing chat contacts')
    loadContacts()
  })
})

// Watch for online status updates
watch(() => wsState.onlineUsers, () => {
  // Update contacts online status
  contacts.value.forEach(contact => {
    contact.is_online = wsState.onlineUsers.has(contact.user_id)
  })
  // Update open chats online status
  openChats.value.forEach(chat => {
    chat.is_online = wsState.onlineUsers.has(chat.user_id)
  })
}, { deep: true })
</script>

<style scoped>
.chat {
  position: fixed;
  bottom: 0;
  right: 20px;
  z-index: 9999;
  display: flex;
  flex-direction: row-reverse;
  gap: 10px;
  align-items: flex-end;
}

/* Sidebar */
.chat-sidebar {
  width: 280px;
  height: 455px;
  background: rgba(5, 6, 13, 0.95);
  border-radius: 1.25rem 1.25rem 0 0;
  box-shadow: var(--shadow);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-bottom: none;
  display: flex;
  flex-direction: column;
  transition: all 0.35s ease;
  backdrop-filter: blur(20px);
}

.chat-sidebar.sidebar-collapsed {
  height: 48px;
}

.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  background: linear-gradient(135deg, rgba(0, 247, 255, 0.08), rgba(255, 0, 230, 0.08));
  border-radius: 1.25rem 1.25rem 0 0;
}

.sidebar-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: #f8f9ff;
  letter-spacing: 0.05em;
}

.toggle-btn {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  font-size: 20px;
  cursor: pointer;
  color: var(--neon-cyan);
  padding: 0;
  width: 28px;
  height: 28px;
  border-radius: 0.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.toggle-btn:hover {
  background: rgba(0, 247, 255, 0.15);
  border-color: var(--border-glow);
  box-shadow: 0 0 12px rgba(0, 247, 255, 0.3);
}

.sidebar-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.search-box {
  padding: 8px 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.search-input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(0, 0, 0, 0.45);
  border-radius: 1rem;
  font-size: 14px;
  outline: none;
  color: #f8f9ff;
  transition: all 0.2s ease;
}

.search-input::placeholder {
  color: var(--text-muted);
}

.search-input:focus {
  border-color: var(--border-glow);
  box-shadow: 0 0 12px rgba(0, 247, 255, 0.2);
}

.contacts-list {
  flex: 1;
  overflow-y: auto;
}

.loading-state,
.empty-state {
  padding: 40px 20px;
  text-align: center;
  color: var(--text-muted);
}

.empty-state small {
  display: block;
  margin-top: 8px;
  font-size: 12px;
  color: var(--text-muted);
}

.contact-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
  border-left: 2px solid transparent;
}

.contact-item:hover {
  background: rgba(255, 255, 255, 0.05);
  border-left-color: var(--neon-cyan);
}

.contact-item.active {
  background: rgba(0, 247, 255, 0.1);
  border-left-color: var(--neon-cyan);
}

.contact-avatar {
  position: relative;
  margin-right: 12px;
  flex-shrink: 0;
}

.avatar-circle {
  width: 40px;
  height: 40px;
  border-radius: 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #05060d;
  font-weight: 600;
  font-size: 14px;
  border: 1px solid rgba(255, 255, 255, 0.15);
}

.avatar-circle.small {
  width: 32px;
  height: 32px;
  font-size: 12px;
  border-radius: 0.75rem;
}

.online-indicator {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 12px;
  height: 12px;
  background: var(--neon-cyan);
  border: 2px solid rgba(5, 6, 13, 0.95);
  border-radius: 50%;
  box-shadow: 0 0 10px var(--neon-cyan);
}

.online-indicator.small {
  width: 10px;
  height: 10px;
}

.contact-info {
  flex: 1;
  min-width: 0;
}

.contact-name {
  font-weight: 600;
  font-size: 14px;
  color: #f8f9ff;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.contact-status {
  font-size: 12px;
  color: var(--text-muted);
}

.unread-badge {
  background: var(--neon-pink);
  color: #05060d;
  font-size: 11px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 10px;
  min-width: 18px;
  text-align: center;
  box-shadow: 0 0 14px rgba(255, 0, 230, 0.55);
}

/* Chat Windows */
.chat-windows {
  display: flex;
  gap: 10px;
  align-items: flex-end;
}

.chat-window {
  width: 328px;
  height: 455px;
  background: rgba(5, 6, 13, 0.95);
  border-radius: 1.25rem 1.25rem 0 0;
  box-shadow: var(--shadow);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-bottom: none;
  display: flex;
  flex-direction: column;
  transition: height 0.35s ease;
  backdrop-filter: blur(20px);
}

.chat-window.minimized {
  height: 48px;
}

.chat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: linear-gradient(135deg, rgba(0, 247, 255, 0.15), rgba(255, 0, 230, 0.15));
  color: #f8f9ff;
  border-radius: 1.25rem 1.25rem 0 0;
  cursor: pointer;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.chat-header-left {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
}

.chat-avatar {
  position: relative;
  margin-right: 10px;
}

.chat-title {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.chat-name {
  font-weight: 600;
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.chat-status {
  font-size: 11px;
  opacity: 0.8;
  color: var(--text-muted);
}

.chat-header-actions {
  display: flex;
  gap: 4px;
}

.header-btn {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.15);
  color: #f8f9ff;
  width: 28px;
  height: 28px;
  border-radius: 0.75rem;
  cursor: pointer;
  font-size: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.header-btn:hover {
  background: rgba(0, 247, 255, 0.2);
  border-color: var(--border-glow);
  box-shadow: 0 0 12px rgba(0, 247, 255, 0.3);
}

.close-btn {
  font-size: 24px;
}

.close-btn:hover {
  background: rgba(255, 0, 230, 0.2);
  border-color: rgba(255, 0, 230, 0.45);
  box-shadow: 0 0 12px rgba(255, 0, 230, 0.3);
}

.chat-body {
  flex: 1;
  overflow: hidden;
  background: rgba(8, 10, 22, 0.75);
}

.messages-container {
  height: 100%;
  overflow-y: auto;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.loading-messages,
.no-messages {
  text-align: center;
  padding: 40px 20px;
  color: var(--text-muted);
  font-size: 14px;
}

.message {
  display: flex;
  margin-bottom: 4px;
}

.message-sent {
  justify-content: flex-end;
}

.message-received {
  justify-content: flex-start;
}

.message-bubble {
  max-width: 70%;
  padding: 8px 12px;
  border-radius: 1rem;
  word-wrap: break-word;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.message-sent .message-bubble {
  background: linear-gradient(135deg, rgba(0, 247, 255, 0.25), rgba(0, 247, 255, 0.15));
  color: #f8f9ff;
  border-bottom-right-radius: 0.25rem;
  border-color: rgba(0, 247, 255, 0.3);
  box-shadow: 0 4px 12px rgba(0, 247, 255, 0.15);
}

.message-received .message-bubble {
  background: rgba(16, 18, 32, 0.8);
  color: #f8f9ff;
  border-bottom-left-radius: 0.25rem;
}

.message-bubble p {
  margin: 0 0 4px;
  font-size: 14px;
  line-height: 1.4;
}

.message-time {
  font-size: 11px;
  opacity: 0.6;
  display: block;
  color: var(--text-muted);
}

.chat-footer {
  display: flex;
  padding: 8px 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(5, 6, 13, 0.95);
  gap: 8px;
}

.message-input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 1rem;
  font-size: 14px;
  outline: none;
  background: rgba(0, 0, 0, 0.45);
  color: #f8f9ff;
  transition: all 0.2s ease;
}

.message-input::placeholder {
  color: var(--text-muted);
}

.message-input:focus {
  border-color: var(--border-glow);
  box-shadow: 0 0 12px rgba(0, 247, 255, 0.2);
}

.send-btn {
  background: linear-gradient(120deg, var(--neon-cyan), var(--neon-pink));
  border: none;
  color: #05060d;
  width: 36px;
  height: 36px;
  border-radius: 0.75rem;
  cursor: pointer;
  font-size: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  font-weight: 700;
  box-shadow: 0 4px 16px rgba(0, 247, 255, 0.3);
}

.send-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 247, 255, 0.4);
}

.send-btn:disabled {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-muted);
  cursor: not-allowed;
  box-shadow: none;
}

/* Scrollbar styling */
.contacts-list::-webkit-scrollbar,
.messages-container::-webkit-scrollbar {
  width: 6px;
}

.contacts-list::-webkit-scrollbar-track,
.messages-container::-webkit-scrollbar-track {
  background: transparent;
}

.contacts-list::-webkit-scrollbar-thumb,
.messages-container::-webkit-scrollbar-thumb {
  background: rgba(0, 247, 255, 0.3);
  border-radius: 3px;
}

.contacts-list::-webkit-scrollbar-thumb:hover,
.messages-container::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 247, 255, 0.5);
}

@media (max-width: 768px) {
  .chat {
    right: 10px;
    gap: 8px;
  }

  .chat-sidebar,
  .chat-window {
    width: min(280px, calc(100vw - 20px));
  }
}
</style>
