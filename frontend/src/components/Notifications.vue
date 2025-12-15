<template>
  <div class="notifications-widget">
    <!-- Toggle Button -->
    <button @click="togglePanel" class="notif-toggle-btn" :class="{ 'has-unread': unreadCount > 0 }">
      <span class="bell-icon">🔔</span>
      <span v-if="unreadCount > 0" class="notif-badge">{{ unreadCount }}</span>
    </button>

    <!-- Notifications Panel -->
    <transition name="slide-fade">
      <div v-if="showPanel" class="notif-panel">
        <!-- Header -->
        <div class="notif-header">
          <h3>Notifications</h3>
          <div class="header-actions">
            <button @click="markAllAsRead" class="action-btn" :disabled="unreadCount === 0">
              Mark all read
            </button>
            <button @click="closePanel" class="close-btn">×</button>
          </div>
        </div>

        <!-- Connection Status -->
        <div v-if="!connected && !connecting" class="connection-status warning">
          <span>⚠️ Not connected to notifications</span>
          <button @click="connect" class="reconnect-btn">Reconnect</button>
        </div>
        <div v-else-if="connecting" class="connection-status">
          <span>🔄 Connecting...</span>
        </div>

        <!-- Notifications List -->
        <div class="notif-list">
          <div v-if="loading" class="loading-state">
            <p>Loading notifications...</p>
          </div>

          <div v-else-if="notifications.length === 0" class="empty-state">
            <span class="empty-icon">🔕</span>
            <p>No notifications yet</p>
            <small>You'll see updates here when something happens</small>
          </div>

          <div
            v-for="notif in notifications"
            :key="notif.id"
            class="notif-item"
            :class="{ 'unread': !notif.is_read }"
          >
            <!-- Notification Icon -->
            <div class="notif-icon" :class="`notif-type-${notif.type}`">
              {{ getNotificationIcon(notif.type) }}
            </div>

            <!-- Notification Content -->
            <div class="notif-content">
              <p class="notif-text">{{ notif.content }}</p>
              <span class="notif-time">{{ formatTime(notif.created_at) }}</span>

              <!-- Action Buttons for Follow Requests -->
              <div v-if="notif.type === 'follow_request' && !notif.is_read" class="notif-actions">
                <button
                  @click="respondToFollowRequest(notif, true)"
                  class="accept-btn"
                  :disabled="processingNotif === notif.id"
                >
                  {{ processingNotif === notif.id ? '...' : '✓ Accept' }}
                </button>
                <button
                  @click="respondToFollowRequest(notif, false)"
                  class="reject-btn"
                  :disabled="processingNotif === notif.id"
                >
                  {{ processingNotif === notif.id ? '...' : '✗ Reject' }}
                </button>
              </div>

              <!-- Action Buttons for Group Invitations -->
              <div v-if="notif.type === 'group_invite' && !notif.is_read" class="notif-actions">
                <button
                  @click="respondToGroupInvitation(notif, true)"
                  class="accept-btn"
                  :disabled="processingNotif === notif.id"
                >
                  {{ processingNotif === notif.id ? '...' : '✓ Accept' }}
                </button>
                <button
                  @click="respondToGroupInvitation(notif, false)"
                  class="reject-btn"
                  :disabled="processingNotif === notif.id"
                >
                  {{ processingNotif === notif.id ? '...' : '✗ Decline' }}
                </button>
              </div>
            </div>

            <!-- Mark as Read / Delete -->
            <div class="notif-item-actions">
              <button
                v-if="!notif.is_read"
                @click="markAsRead(notif.id)"
                class="mark-read-btn"
                title="Mark as read"
              >
                ✓
              </button>
              <button
                @click="deleteNotification(notif.id)"
                class="delete-btn"
                title="Delete"
              >
                ×
              </button>
            </div>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, onBeforeUnmount } from 'vue'
import { useNotifications } from '../composables/useNotifications'
import { getToken } from '../stores/auth'
import { useToast } from '@/composables/useToast'
import { respondToFollowRequest as respondToFollowRequestService } from '@/services/usersService'
import { respondToInvitation, getMyInvitations } from '@/services/groupsService'
import { throttle } from '@/utils/timing'

const { error, success } = useToast()

const {
  connected,
  connecting,
  notifications,
  unreadCount,
  connect,
  disconnect,
  markAsRead: markNotifAsRead,
  markAllAsRead: markAllNotifAsRead,
  deleteNotification: deleteNotif,
  on
} = useNotifications()

const showPanel = ref(false)
const loading = ref(false)
const processingNotif = ref(null)

function togglePanel() {
  showPanel.value = !showPanel.value
}

function closePanel() {
  showPanel.value = false
}

function markAllAsRead() {
  markAllNotifAsRead()
}

function markAsRead(notifId) {
  markNotifAsRead(notifId)
}

function deleteNotification(notifId) {
  deleteNotif(notifId)
}

const respondToFollowRequest = throttle(async (notif, accept) => {
  processingNotif.value = notif.id
  
  try {
    const token = getToken()
    await respondToFollowRequestService(notif.related_id, accept, token)

    // Mark notification as read
    markAsRead(notif.id)
    
    // Emit event for other components to refresh (e.g., chat contacts)
    if (accept) {
      window.dispatchEvent(new CustomEvent('follow-accepted'))
    }
    
    // Show feedback
    success(accept ? 'Follow request accepted' : 'Follow request rejected')
  } catch (err) {
    console.error('Error responding to follow request:', err.message)
    error(err.message || 'Failed to respond to follow request. Please try again.')
  } finally {
    processingNotif.value = null
  }
}, 1000)

const respondToGroupInvitation = throttle(async (notif, accept) => {
  processingNotif.value = notif.id
  
  try {
    const token = getToken()
    
    // notif.related_id is the group_id, we need to find the invitation_id
    // Get all invitations and find the one for this group
    const invitations = await getMyInvitations(token)
    const invitation = invitations.find(inv => inv.group_id === notif.related_id)
    
    if (!invitation) {
      throw new Error('Invitation not found or already responded')
    }
    
    // Now respond with the correct invitation ID
    await respondToInvitation(invitation.id, accept, token)

    // Mark notification as read
    markAsRead(notif.id)
    
    // Emit event for groups page to refresh
    if (accept) {
      window.dispatchEvent(new CustomEvent('group-joined'))
    }
    
    // Show feedback
    success(accept ? 'Group invitation accepted' : 'Group invitation declined')
  } catch (err) {
    console.error('Error responding to group invitation:', err.message)
    error(err.message || 'Failed to respond to invitation. Please try again.')
  } finally {
    processingNotif.value = null
  }
}, 1000)

function getNotificationIcon(type) {
  const icons = {
    'follow_request': '👤',
    'follow': '✅',
    'group_invite': '👥',
    'group_request': '📝',
    'group_activity': '🔔',
    'event': '📅',
    'message': '💬',
    'comment': '💭',
    'post': '📄'
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
onMounted(() => {
  // Connect to notifications WebSocket
  connect()

  // Listen for new notification events
  on('notification', (notification) => {
    console.log('New notification received:', notification)
    // Notification is already added to the notifications array by the composable
    
    // Play sound or show browser notification here if desired
  })
})

// Close panel when clicking outside
const handleClickOutside = (event) => {
  if (!showPanel.value) return

  const target = event.target
  if (target instanceof Element && !target.closest('.notifications-widget')) {
    closePanel()
  }
}

watch(showPanel, (isOpen) => {
  if (isOpen) {
    document.addEventListener('click', handleClickOutside)
  } else {
    document.removeEventListener('click', handleClickOutside)
  }
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.notifications-widget {
  position: relative;
}

/* Toggle Button */
.notif-toggle-btn {
  position: relative;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 1rem;
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 24px;
}

.notif-toggle-btn:hover {
  background: rgba(0, 247, 255, 0.15);
  border-color: var(--border-glow);
  box-shadow: 0 0 12px rgba(0, 247, 255, 0.3);
}

.notif-toggle-btn.has-unread {
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(0, 247, 255, 0.7); }
  50% { box-shadow: 0 0 0 8px rgba(0, 247, 255, 0); }
}

.bell-icon {
  filter: drop-shadow(0 0 8px rgba(0, 247, 255, 0.5));
}

.notif-badge {
  position: absolute;
  top: -4px;
  right: -4px;
  background: var(--neon-pink);
  color: #05060d;
  font-size: 11px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 10px;
  min-width: 20px;
  text-align: center;
  box-shadow: 0 0 14px rgba(255, 0, 230, 0.55);
}

/* Notifications Panel */
.notif-panel {
  position: absolute;
  top: calc(100% + 10px);
  right: 0;
  width: 420px;
  max-height: 600px;
  background: rgba(5, 6, 13, 0.98);
  border-radius: 1.25rem;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4), 0 0 0 1px rgba(255, 255, 255, 0.08);
  backdrop-filter: blur(20px);
  display: flex;
  flex-direction: column;
  z-index: 10000;
}

/* Transitions */
.slide-fade-enter-active {
  transition: all 0.3s ease;
}

.slide-fade-leave-active {
  transition: all 0.2s ease;
}

.slide-fade-enter-from {
  transform: translateY(-10px);
  opacity: 0;
}

.slide-fade-leave-to {
  transform: translateY(-5px);
  opacity: 0;
}

/* Header */
.notif-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  background: linear-gradient(135deg, rgba(0, 247, 255, 0.08), rgba(255, 0, 230, 0.08));
  border-radius: 1.25rem 1.25rem 0 0;
}

.notif-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
  color: #f8f9ff;
  letter-spacing: 0.05em;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.action-btn {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #f8f9ff;
  padding: 6px 12px;
  border-radius: 0.75rem;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-btn:hover:not(:disabled) {
  background: rgba(0, 247, 255, 0.15);
  border-color: var(--border-glow);
}

.action-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.close-btn {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #f8f9ff;
  width: 32px;
  height: 32px;
  border-radius: 0.75rem;
  font-size: 24px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background: rgba(255, 0, 230, 0.2);
  border-color: rgba(255, 0, 230, 0.45);
}

/* Connection Status */
.connection-status {
  padding: 12px 20px;
  background: rgba(0, 247, 255, 0.1);
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
  color: var(--neon-cyan);
}

.connection-status.warning {
  background: rgba(255, 193, 7, 0.1);
  color: #ffc107;
}

.reconnect-btn {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: #f8f9ff;
  padding: 4px 12px;
  border-radius: 0.5rem;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.reconnect-btn:hover {
  background: rgba(255, 255, 255, 0.2);
}

/* Notifications List */
.notif-list {
  flex: 1;
  overflow-y: auto;
  max-height: 500px;
}

.loading-state,
.empty-state {
  padding: 60px 20px;
  text-align: center;
  color: var(--text-muted);
}

.empty-icon {
  font-size: 48px;
  display: block;
  margin-bottom: 16px;
  opacity: 0.3;
}

.empty-state p {
  margin: 0 0 8px;
  font-size: 16px;
  color: #f8f9ff;
}

.empty-state small {
  font-size: 13px;
  color: var(--text-muted);
}

/* Notification Item */
.notif-item {
  display: flex;
  gap: 12px;
  padding: 16px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  transition: all 0.2s ease;
  position: relative;
}

.notif-item:hover {
  background: rgba(255, 255, 255, 0.03);
}

.notif-item.unread {
  background: rgba(0, 247, 255, 0.05);
  border-left: 3px solid var(--neon-cyan);
}

.notif-icon {
  width: 40px;
  height: 40px;
  border-radius: 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  flex-shrink: 0;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.notif-type-follow_request {
  background: rgba(0, 123, 255, 0.1);
  border-color: rgba(0, 123, 255, 0.3);
}

.notif-type-follow {
  background: rgba(40, 167, 69, 0.1);
  border-color: rgba(40, 167, 69, 0.3);
}

.notif-type-group_invite {
  background: rgba(255, 193, 7, 0.1);
  border-color: rgba(255, 193, 7, 0.3);
}

.notif-type-message {
  background: rgba(0, 247, 255, 0.1);
  border-color: rgba(0, 247, 255, 0.3);
}

.notif-content {
  flex: 1;
  min-width: 0;
}

.notif-text {
  margin: 0 0 4px;
  font-size: 14px;
  line-height: 1.4;
  color: #f8f9ff;
}

.notif-time {
  font-size: 12px;
  color: var(--text-muted);
}

/* Notification Actions */
.notif-actions {
  display: flex;
  gap: 8px;
  margin-top: 10px;
}

.accept-btn,
.reject-btn {
  padding: 6px 16px;
  border-radius: 0.75rem;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid;
}

.accept-btn {
  background: rgba(40, 167, 69, 0.15);
  border-color: rgba(40, 167, 69, 0.3);
  color: #28a745;
}

.accept-btn:hover:not(:disabled) {
  background: rgba(40, 167, 69, 0.25);
  transform: translateY(-1px);
}

.reject-btn {
  background: rgba(220, 53, 69, 0.15);
  border-color: rgba(220, 53, 69, 0.3);
  color: #dc3545;
}

.reject-btn:hover:not(:disabled) {
  background: rgba(220, 53, 69, 0.25);
  transform: translateY(-1px);
}

.accept-btn:disabled,
.reject-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Item Actions */
.notif-item-actions {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.mark-read-btn,
.delete-btn {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  width: 28px;
  height: 28px;
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  font-size: 16px;
  color: #f8f9ff;
  transition: all 0.2s ease;
}

.mark-read-btn:hover {
  background: rgba(0, 247, 255, 0.15);
  border-color: var(--border-glow);
}

.delete-btn:hover {
  background: rgba(220, 53, 69, 0.15);
  border-color: rgba(220, 53, 69, 0.3);
  color: #dc3545;
}

/* Scrollbar */
.notif-list::-webkit-scrollbar {
  width: 6px;
}

.notif-list::-webkit-scrollbar-track {
  background: transparent;
}

.notif-list::-webkit-scrollbar-thumb {
  background: rgba(0, 247, 255, 0.3);
  border-radius: 3px;
}

.notif-list::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 247, 255, 0.5);
}

@media (max-width: 768px) {
  .notif-panel {
    width: calc(100vw - 40px);
    max-width: 420px;
  }
}
</style>
