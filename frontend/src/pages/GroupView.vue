<template>
  <div class="group-view">
    <div v-if="loading" class="loading-state">Loading group...</div>
    
    <div v-else-if="error" class="error-state">
      <p>{{ error }}</p>
      <button class="cta" @click="$router.push('/')">Back to Feed</button>
    </div>

    <template v-else-if="group">
      <!-- Group Header -->
      <section class="group-header">
        <div class="group-banner">
          <div class="group-icon-large">
            <img v-if="group.image_url" :src="getGroupImageUrl(group.image_url)" alt="Group avatar" />
            <span v-else>{{ getGroupInitials(group.name) }}</span>
            <button v-if="isCreator" class="avatar-upload-btn" @click="triggerAvatarUpload" title="Change group image">
              📷
            </button>
            <input 
              type="file" 
              accept="image/*" 
              ref="avatarInput" 
              style="display: none"
              @change="handleAvatarChange" 
            />
          </div>
          <div class="group-info">
            <h1>{{ group.name }}</h1>
            <p class="group-description">{{ group.description }}</p>
            <div class="group-stats">
              <span><strong>{{ group.member_count || 0 }}</strong> members</span>
              <span><strong>{{ events.length }}</strong> events</span>
              <span>Created {{ formatDate(group.created_at) }}</span>
            </div>
          </div>
        </div>
        
        <!-- Group Actions -->
        <div class="group-actions">
          <button v-if="!isMember && !hasPendingRequest" class="cta" @click="requestToJoin">
            Request to Join
          </button>
          <button v-else-if="hasPendingRequest" class="ghost" disabled>
            Request Pending
          </button>
          <template v-else-if="isMember">
            <button class="cta" @click="showInviteModal = true">
              <span>+</span> Invite
            </button>
            <button v-if="!isCreator" class="ghost" @click="leaveGroup">
              Leave Group
            </button>
          </template>
        </div>
      </section>

      <!-- Group Tabs -->
      <div class="group-tabs">
        <button
          :class="['tab-btn', { active: activeTab === 'posts' }]"
          @click="activeTab = 'posts'"
        >
          <span class="icon-posts">📝</span>
          Posts
        </button>
        <button
          v-if="isMember"
          :class="['tab-btn', { active: activeTab === 'chat' }]"
          @click="activeTab = 'chat'"
        >
          <span class="icon-chat">💬</span>
          Chat
        </button>
        <button
          :class="['tab-btn', { active: activeTab === 'events' }]"
          @click="activeTab = 'events'"
        >
          <span class="icon-events">📅</span>
          Events
        </button>
        <button
          :class="['tab-btn', { active: activeTab === 'members' }]"
          @click="activeTab = 'members'"
        >
          <span class="icon-members">👥</span>
          Members
        </button>
        <button
          v-if="isCreator"
          :class="['tab-btn', { active: activeTab === 'requests' }]"
          @click="activeTab = 'requests'"
        >
          <span class="icon-requests">📬</span>
          Requests
          <span v-if="pendingRequests.length > 0" class="badge">{{ pendingRequests.length }}</span>
        </button>
      </div>

      <!-- Tab Content -->
      <section class="tab-content">
        <!-- Posts Tab (Feed-like with images, comments, edit/delete) -->
        <div v-if="activeTab === 'posts'" class="posts-tab">
          <CreatePost 
            v-if="isMember" 
            :groupId="groupId" 
            @postCreated="loadGroupPosts"
          />
          
          <div class="posts-feed">
            <div v-if="loadingPosts" class="loading">Loading posts...</div>
            <p v-else-if="!isMember" class="empty-state">
              Join the group to see posts
            </p>
            <p v-else-if="groupPosts.length === 0" class="empty-state">
              No posts yet. Be the first to post!
            </p>
            
            <article
              v-else
              v-for="post in groupPosts"
              :key="post.id"
              class="post-card"
              @click="navigateToPost(post.id)"
            >
              <div class="post-header">
                <div class="avatar" v-if="post.author?.avatar_path">
                  <img :src="getUserAvatarUrl(post.author, 48)" alt="Avatar" />
                </div>
                <div class="avatar" v-else>
                  <span class="initials">{{ getAuthorInitials(post.author) }}</span>
                </div>
                <div class="author-info">
                  <h3>{{ getAuthorName(post.author) }}</h3>
                  <p class="meta">
                    {{ formatTime(post.created_at) }}
                  </p>
                </div>
                <button class="menu-btn" @click.stop="showPostMenu(post.id)">⋮</button>
              </div>

              <h3 v-if="post.title" class="post-title">{{ post.title }}</h3>
              <p class="post-content">{{ post.content }}</p>

              <img
                v-if="post.image_path"
                :src="getImageUrl(post.image_path)"
                alt="Post image"
                class="post-image"
              />
            </article>
          </div>
        </div>

        <!-- Chat Tab (Real-time messaging) -->
        <div v-if="activeTab === 'chat'" class="chat-tab">
          <div class="chat-container">
            <div class="chat-messages" ref="chatMessagesEl">
              <div v-if="loadingChat" class="loading">Loading chat...</div>
              <div v-else-if="chatMessages.length === 0" class="empty-state">
                No messages yet. Start the conversation!
              </div>
              <div v-else v-for="message in chatMessages" :key="message.id" class="chat-message">
                <div class="message-avatar">
                  <span class="initials">{{ getMemberInitials(message.sender_id) }}</span>
                </div>
                <div class="message-content">
                  <div class="message-header">
                    <strong>{{ getMemberName(message.sender_id) }}</strong>
                    <small>{{ formatTime(message.created_at) }}</small>
                  </div>
                  <p>{{ message.content }}</p>
                </div>
              </div>
            </div>
            
            <div class="chat-input-area">
              <textarea
                v-model="newChatMessage"
                placeholder="Type a message..."
                rows="2"
                maxlength="1000"
                @keydown.enter.prevent="sendChatMessage"
              />
              <div class="chat-input-buttons">
                <button class="emoji-btn" @click="showEmojiPicker = !showEmojiPicker" type="button" title="Add emoji">
                  😊
                </button>
                <button class="cta" @click="sendChatMessage" :disabled="!newChatMessage.trim() || sendingChat">
                  {{ sendingChat ? 'Sending...' : 'Send' }}
                </button>
              </div>
              <EmojiPicker 
                :isOpen="showEmojiPicker" 
                @select="selectEmoji" 
                @close="showEmojiPicker = false" 
              />
            </div>
          </div>
        </div>

        <!-- Events Tab -->
        <div v-if="activeTab === 'events'" class="events-tab">
          <button v-if="isMember" class="cta mb-2" @click="showCreateEventModal = true">
            Create Event
          </button>
          
          <div v-if="loadingEvents" class="loading">Loading events...</div>
          <p v-else-if="!isMember" class="empty-state">
            Join the group to see events
          </p>
          <p v-else-if="events.length === 0" class="empty-state">
            No events scheduled
          </p>
          <article v-else v-for="event in events" :key="event.id" class="event-card">
            <h3>{{ event.title }}</h3>
            <p class="event-description">{{ event.description }}</p>
            <div class="event-meta">
              <span class="event-date">📅 {{ formatEventDate(event.event_time) }}</span>
              <span class="event-creator">By {{ event.creator_name }}</span>
            </div>
            
            <div class="rsvp-section">
              <div class="rsvp-counts">
                <span class="going-count">✓ {{ event.going_count || 0 }} Going</span>
                <span class="not-going-count">✗ {{ event.not_going_count || 0 }} Not Going</span>
              </div>
              <div class="rsvp-buttons">
                <button
                  :class="['rsvp-btn', 'going', { active: event.user_response === 'going' }]"
                  @click="respondToEvent(event.id, 'going')"
                  :disabled="respondingToEvent === event.id"
                >
                  ✓ Going
                </button>
                <button
                  :class="['rsvp-btn', 'not-going', { active: event.user_response === 'not_going' }]"
                  @click="respondToEvent(event.id, 'not_going')"
                  :disabled="respondingToEvent === event.id"
                >
                  ✗ Not Going
                </button>
              </div>
            </div>
          </article>
        </div>

        <!-- Members Tab -->
        <div v-if="activeTab === 'members'" class="members-tab">
          <div v-if="loadingMembers" class="loading">Loading members...</div>
          <div v-else class="members-grid">
            <article 
              v-for="member in members" 
              :key="member.user_id" 
              class="member-card"
              @click="viewUserProfile(member.user_id)"
            >
              <div class="member-avatar">{{ getMemberInitialsFromData(member) }}</div>
              <div class="member-info">
                <button
                  class="member-name-btn"
                  type="button"
                  @click.stop="viewUserProfile(member.user_id)"
                >
                  {{ getMemberNameFromData(member) }}
                </button>
                <br/>
                <span class="member-role">{{ member.role === 'creator' ? 'Creator' : 'Member' }}</span>
              </div>
            </article>
          </div>
        </div>

        <!-- Requests Tab (Creator Only) -->
        <div v-if="activeTab === 'requests' && isCreator" class="requests-tab">
          <div v-if="loadingRequests" class="loading">Loading requests...</div>
          <p v-else-if="pendingRequests.length === 0" class="empty-state">
            No pending requests
          </p>
          <article v-else v-for="request in pendingRequests" :key="request.id" class="request-card">
            <div class="member-avatar">{{ getMemberInitialsFromData(request) }}</div>
            <div class="request-info">
              <strong>{{ getMemberNameFromData(request) }}</strong>
              <small>Requested {{ formatTime(request.joined_at) }}</small>
            </div>
            <div class="request-actions">
              <button class="cta mini" @click="respondToRequest(request.id, true)">Accept</button>
              <button class="ghost mini" @click="respondToRequest(request.id, false)">Decline</button>
            </div>
          </article>
        </div>
      </section>
    </template>

    <!-- Invite Modal -->
    <Teleport to="body">
      <div v-if="showInviteModal" class="modal-overlay" @click="closeInviteModal">
        <div class="modal-content" @click.stop>
          <header class="modal-header">
            <h3>Invite to {{ group?.name }}</h3>
            <button class="close-btn" @click="closeInviteModal">✕</button>
          </header>
          <div class="invite-search">
            <input
              v-model="inviteSearchQuery"
              type="search"
              placeholder="Search users to invite..."
              @input="searchUsersToInvite"
            />
          </div>
          <div class="invite-results">
            <p v-if="inviteSearchQuery && searchResults.length === 0" class="empty-state">
              No users found
            </p>
            <article
              v-for="user in searchResults"
              :key="user.id"
              class="invite-user-card"
              @click="inviteUser(user.id)"
            >
              <div class="user-avatar">{{ getInitials(user.first_name + ' ' + user.last_name) }}</div>
              <div>
                <strong>{{ user.first_name }} {{ user.last_name }}</strong>
                <small>@{{ user.username }}</small>
              </div>
              <button class="cta mini">Invite</button>
            </article>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Create Event Modal -->
    <Teleport to="body">
      <div v-if="showCreateEventModal" class="modal-overlay" @click="closeCreateEventModal">
        <div class="modal-content" @click.stop>
          <header class="modal-header">
            <h3>Create Event</h3>
            <button class="close-btn" @click="closeCreateEventModal">✕</button>
          </header>
          <form @submit.prevent="createEvent">
            <div class="form-field">
              <label for="event-title">Event Title *</label>
              <input
                id="event-title"
                v-model="newEvent.title"
                type="text"
                placeholder="Event name..."
                required
                maxlength="100"
              />
            </div>
            <div class="form-field">
              <label for="event-desc">Description *</label>
              <textarea
                id="event-desc"
                v-model="newEvent.description"
                placeholder="What's this event about?"
                required
                rows="3"
                maxlength="500"
              />
            </div>
            <div class="form-field">
              <label for="event-date">Date & Time *</label>
              <input
                id="event-date"
                v-model="newEvent.event_time"
                type="datetime-local"
                :min="minEventDateTime"
                required
              />
            </div>
            <p v-if="eventError" class="error-message">{{ eventError }}</p>
            <div class="modal-actions">
              <button type="button" class="ghost" @click="closeCreateEventModal">Cancel</button>
              <button type="submit" class="cta" :disabled="creatingEvent">
                {{ creatingEvent ? 'Creating...' : 'Create Event' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { getToken, getUser } from '@/stores/auth'
import CreatePost from '@/components/CreatePost.vue'
import EmojiPicker from '@/components/EmojiPicker.vue'
import {
  getGroup,
  getGroupMembers,
  getGroupEvents,
  getGroupMessages,
  createGroupMessage,
  requestJoinGroup,
  respondToEvent as respondToEventService,
  createEvent as createEventService,
  inviteToGroup,
  getPendingRequests,
  respondToRequest as respondToRequestService,
  updateGroupImage,
  getGroupImageUrl
} from '@/services/groupsService'
import { getGroupPosts as fetchGroupPosts, getPostImageUrl } from '@/services/postsService'
import { searchUsersForGroup } from '@/services/usersService'
import { useToast } from '@/composables/useToast'
import { useWebSocket } from '@/composables/useWebSocket'
import { leaveGroup as leaveGroupService } from '@/services/groupsService'
import { useAvatar } from '@/composables/useAvatar'
import { throttle, debounce } from '@/utils/timing'

const router = useRouter()
const route = useRoute()
const { getUserAvatarUrl, getAvatarUrl } = useAvatar()

// WebSocket connection for live chat
const { connected: wsConnected, sendGroupMessage, on: wsOn, off: wsOff } = useWebSocket()
const { success: showSuccess, error: showError } = useToast()

const groupId = computed(() => parseInt(route.params.id))
const currentUser = getUser()

const loading = ref(true)
const error = ref('')
const group = ref(null)
const members = ref([])
const events = ref([])
const groupPosts = ref([])
const chatMessages = ref([])
const pendingRequests = ref([])

const loadingMembers = ref(false)
const loadingEvents = ref(false)
const loadingPosts = ref(false)
const loadingChat = ref(false)
const loadingRequests = ref(false)

const activeTab = ref('posts')
const newChatMessage = ref('')
const sendingChat = ref(false)
const chatMessagesEl = ref(null)
const showEmojiPicker = ref(false)

const showInviteModal = ref(false)
const inviteSearchQuery = ref('')
const searchResults = ref([])
const searchTimeout = ref(null)

const showCreateEventModal = ref(false)
const newEvent = ref({
  title: '',
  description: '',
  event_time: ''
})
const creatingEvent = ref(false)
const eventError = ref('')
const respondingToEvent = ref(null)

const avatarInput = ref(null)
const uploadingAvatar = ref(false)

// Minimum datetime for event creation (now)
const minEventDateTime = computed(() => {
  const now = new Date()
  // Format as YYYY-MM-DDTHH:mm for datetime-local input
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  const hours = String(now.getHours()).padStart(2, '0')
  const minutes = String(now.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day}T${hours}:${minutes}`
})

const isMember = computed(() => {
  if (!group.value || !currentUser) return false
  return group.value.is_member === true
})

const isCreator = computed(() => {
  if (!group.value || !currentUser) return false
  return group.value.creator_id === currentUser.id
})

const hasPendingRequest = computed(() => {
  if (!group.value) return false
  return group.value.has_pending_request === true
})

function getGroupInitials(title) {
  if (!title) return '?'
  const words = title.split(' ')
  if (words.length >= 2) {
    return (words[0][0] + words[1][0]).toUpperCase()
  }
  return title.substring(0, 2).toUpperCase()
}

function getInitials(name) {
  if (!name) return '?'
  const parts = name.trim().split(' ')
  if (parts.length >= 2) {
    return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase()
  }
  return name.substring(0, 2).toUpperCase()
}

function getAuthorName(author) {
  if (!author) return 'Unknown'
  if (author.first_name && author.last_name) {
    return `${author.first_name} ${author.last_name}`
  }
  if (author.first_name) return author.first_name
  return author.username || 'Unknown'
}

function getAuthorInitials(author) {
  if (!author) return '?'
  if (author.first_name && author.last_name) {
    return `${author.first_name[0]}${author.last_name[0]}`.toUpperCase()
  }
  if (author.first_name) return author.first_name.substring(0, 2).toUpperCase()
  if (author.username) return author.username.substring(0, 2).toUpperCase()
  return '?'
}

function formatDate(dateString) {
  if (!dateString) return ''
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })
}

function formatTime(timestamp) {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  const now = new Date()
  const diff = Math.floor((now - date) / 1000)
  if (diff < 60) return `${diff}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  return `${Math.floor(diff / 86400)}d ago`
}

function formatEventDate(dateString) {
  if (!dateString) return ''
  const date = new Date(dateString)
  return date.toLocaleString('en-US', {
    weekday: 'short',
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

async function loadGroup() {
  const token = getToken()
  if (!token) {
    error.value = 'You must be logged in'
    loading.value = false
    return
  }

  try {
    const data = await getGroup(groupId.value, token)
    // Backend returns the group object directly after unwrapping
    group.value = data
  } catch (err) {
    console.error('Failed to load group:', err)
    error.value = err.message || 'Failed to load group'
  } finally {
    loading.value = false
  }
}

async function loadMembers() {
  const token = getToken()
  if (!token) return

  loadingMembers.value = true
  try {
    const data = await getGroupMembers(groupId.value, token)
    // Backend returns array directly after unwrapping
    members.value = Array.isArray(data) ? data : []
  } catch (err) {
    console.error('Failed to load members:', err)
    members.value = []
  } finally {
    loadingMembers.value = false
  }
}

async function loadEvents() {
  if (!isMember.value) return
  
  const token = getToken()
  if (!token) return

  loadingEvents.value = true
  try {
    const data = await getGroupEvents(groupId.value, token)
    // Backend returns array directly after unwrapping
    events.value = Array.isArray(data) ? data : []
  } catch (err) {
    console.error('Failed to load events:', err)
  } finally {
    loadingEvents.value = false
  }
}

async function loadGroupPosts() {
  if (!isMember.value) return
  
  const token = getToken()
  if (!token) return

  loadingPosts.value = true
  try {
    const data = await fetchGroupPosts(groupId.value, token)
    groupPosts.value = data.posts || []
  } catch (err) {
    console.error('Failed to load group posts:', err)
  } finally {
    loadingPosts.value = false
  }
}

async function loadChatMessages() {
  if (!isMember.value) return
  
  const token = getToken()
  if (!token) return

  loadingChat.value = true
  try {
    const data = await getGroupMessages(groupId.value, token)
    chatMessages.value = Array.isArray(data) ? data : []
    await nextTick()
    scrollChatToBottom()
  } catch (err) {
    console.error('Failed to load chat messages:', err)
  } finally {
    loadingChat.value = false
  }
}

const sendChatMessage = throttle(() => {
  if (!newChatMessage.value.trim() || sendingChat.value) return
  
  const user = getUser()
  if (!user) return

  sendingChat.value = true
  
  // Send via WebSocket if connected, otherwise fall back to HTTP
  if (wsConnected.value) {
    sendViaWebSocket()
  } else {
    sendViaHTTP()
  }
}, 500)

function sendViaWebSocket() {
  const user = getUser()
  const content = newChatMessage.value.trim()
  
  // Send group message via WebSocket composable
  const success = sendGroupMessage(groupId.value, content)
  
  if (success) {
    newChatMessage.value = ''
    sendingChat.value = false
    
    // Optimistically add message to UI (will be replaced by server confirmation)
    const optimisticMessage = {
      id: Date.now(), // Temporary ID
      group_id: groupId.value,
      sender_id: user.id,
      content: content,
      created_at: new Date().toISOString(),
      _optimistic: true
    }
    chatMessages.value.push(optimisticMessage)
    nextTick(() => scrollChatToBottom())
  } else {
    // Fallback to HTTP if WebSocket send failed
    sendViaHTTP()
  }
}

async function sendViaHTTP() {
  const token = getToken()
  if (!token) return

  try {
    await createGroupMessage(groupId.value, newChatMessage.value, token)
    newChatMessage.value = ''
    await loadChatMessages()
  } catch (err) {
    console.error('Failed to send message:', err)
    showError(err.message || 'Failed to send message')
  } finally {
    sendingChat.value = false
  }
}

function scrollChatToBottom() {
  if (chatMessagesEl.value) {
    chatMessagesEl.value.scrollTop = chatMessagesEl.value.scrollHeight
  }
}

function selectEmoji(emoji) {
  newChatMessage.value += emoji
  showEmojiPicker.value = false
}

async function loadRequests() {
  if (!isCreator.value) return
  
  const token = getToken()
  if (!token) return

  loadingRequests.value = true
  try {
    const data = await getPendingRequests(groupId.value, token)
    // Backend returns array directly after unwrapping
    pendingRequests.value = Array.isArray(data) ? data : []
  } catch (err) {
    console.error('Failed to load requests:', err)
  } finally {
    loadingRequests.value = false
  }
}

async function requestToJoin() {
  const token = getToken()
  if (!token) return

  try {
    await requestJoinGroup(groupId.value, token)
    showSuccess('Join request sent!')
    await loadGroup()
  } catch (err) {
    console.error('Failed to request join:', err)
    showError(err.message || 'Failed to send request')
  }
}

const createEvent = throttle(async () => {
  const token = getToken()
  if (!token) return

  creatingEvent.value = true
  eventError.value = ''

  try {
    // Convert datetime-local format to RFC3339 (ISO 8601)
    // datetime-local gives us "2025-02-12T18:00"
    // We need to add timezone info for RFC3339
    const eventData = {
      ...newEvent.value,
      event_time: newEvent.value.event_time + ':00Z' // Add seconds and UTC timezone
    }
    
    await createEventService(groupId.value, eventData, token)
    showSuccess('Event created!')
    closeCreateEventModal()
    await loadEvents()
  } catch (err) {
    console.error('Failed to create event:', err)
    eventError.value = err.message || 'Failed to create event'
  } finally {
    creatingEvent.value = false
  }
}, 1000)

const respondToEvent = throttle(async (eventId, response) => {
  const token = getToken()
  if (!token) return

  respondingToEvent.value = eventId
  try {
    await respondToEventService(eventId, response, token)
    showSuccess(`Marked as ${response === 'going' ? 'Going' : 'Not Going'}`)
    await loadEvents()
  } catch (err) {
    console.error('Failed to respond to event:', err)
    showError(err.message || 'Failed to update response')
  } finally {
    respondingToEvent.value = null
  }
}, 1000)

function closeInviteModal() {
  showInviteModal.value = false
  inviteSearchQuery.value = ''
  searchResults.value = []
}

function closeCreateEventModal() {
  showCreateEventModal.value = false
  newEvent.value = { title: '', description: '', event_time: '' }
  eventError.value = ''
}

const debouncedUserSearch = debounce(async () => {
  const query = inviteSearchQuery.value.trim()
  if (!query) {
    searchResults.value = []
    return
  }

  const token = getToken()
  if (!token) return

  try {
    const { users = [] } = await searchUsersForGroup(query, groupId.value, token)
    searchResults.value = users
  } catch (err) {
    console.error('Failed to search users:', err)
  }
}, 300)

function searchUsersToInvite() {
  debouncedUserSearch()
}

const inviteUser = throttle(async (userId) => {
  const token = getToken()
  if (!token) return

  try {
    await inviteToGroup(groupId.value, userId, token)
    showSuccess('Invitation sent!')
    closeInviteModal()
  } catch (err) {
    console.error('Failed to invite user:', err)
    showError(err.message || 'Failed to send invitation')
  }
}, 1000)

const respondToRequest = throttle(async (memberId, accept) => {
  const token = getToken()
  if (!token) return

  try {
    await respondToRequestService(groupId.value, memberId, accept, token)
    showSuccess(accept ? 'Request accepted' : 'Request declined')
    await loadRequests()
    await loadMembers()
  } catch (err) {
    console.error('Failed to respond to request:', err)
    showError(err.message || 'Failed to process request')
  }
}, 1000)

async function leaveGroup() {
  const token = getToken()
  if (!token) return
  try {
    await leaveGroupService(groupId.value, token)
    showSuccess('Left group')
    router.push('/')
  } catch (err) {
    console.error('Failed to leave group:', err)
    showError(err.message || 'Failed to leave group')
  }
}

function viewUserProfile(userId) {
  router.push(`/profile/${userId}`)
}

function triggerAvatarUpload() {
  if (!isCreator.value) return
  avatarInput.value?.click()
}

async function handleAvatarChange(event) {
  if (!isCreator.value) return

  const file = event.target.files?.[0]
  if (!file) return

  // Validate file type
  if (!file.type.startsWith('image/')) {
    showError('Please select an image file')
    return
  }

  // Validate file size (max 5MB)
  if (file.size > 5 * 1024 * 1024) {
    showError('File size must be less than 5MB')
    return
  }

  uploadingAvatar.value = true
  try {
    const token = getToken()
    const result = await updateGroupImage(groupId.value, file, token)
    
    // Update group with new image URL
    if (group.value && result.image_url) {
      group.value.image_url = result.image_url
    }
    
    showSuccess('Group image updated successfully!')
  } catch (error) {
    console.error('Failed to upload group image:', error)
    showError(error.message || 'Failed to upload group image')
  } finally {
    uploadingAvatar.value = false
    // Reset input
    if (avatarInput.value) {
      avatarInput.value.value = ''
    }
  }
}

function getImageUrl(path) {
  // For post images, use post service
  return getPostImageUrl(path)
}

function getMemberName(senderId) {
  const user = getUser()
  if (senderId === user?.id) return 'You'
  
  const member = members.value.find(m => m.user_id === senderId)
  if (member) {
    return getMemberNameFromData(member)
  }
  return 'Unknown'
}

function getMemberNameFromData(member) {
  if (member.nickname) return member.nickname
  if (member.first_name && member.last_name) {
    return `${member.first_name} ${member.last_name}`
  }
  if (member.first_name) return member.first_name
  return member.username || 'Unknown'
}

function getMemberInitialsFromData(member) {
  if (member.first_name && member.last_name) {
    return `${member.first_name[0]}${member.last_name[0]}`.toUpperCase()
  }
  if (member.first_name) return member.first_name[0].toUpperCase()
  if (member.username) return member.username.substring(0, 2).toUpperCase()
  return '?'
}

function getMemberInitials(senderId) {
  const member = members.value.find(m => m.user_id === senderId)
  if (member) {
    if (member.first_name && member.last_name) {
      return `${member.first_name[0]}${member.last_name[0]}`.toUpperCase()
    }
    if (member.first_name) return member.first_name[0].toUpperCase()
    if (member.username) return member.username.substring(0, 2).toUpperCase()
  }
  return '?'
}


function navigateToPost(postId) {
  router.push(`/post/${postId}`)
}

function showPostMenu(postId) {
  console.log('Show menu for post:', postId)
  // TODO: Implement edit/delete menu
}

// Handle incoming group messages via WebSocket
function handleGroupMessage(data) {
  // Only handle messages for this group
  if (data.group_id !== groupId.value) return
  
  const user = getUser()
  
  // Remove optimistic message if it exists
  chatMessages.value = chatMessages.value.filter(m => !m._optimistic)
  
  // Add or update message
  const existingIndex = chatMessages.value.findIndex(m => m.id === data.message_id)
  if (existingIndex === -1) {
    chatMessages.value.push({
      id: data.message_id,
      group_id: data.group_id,
      sender_id: data.sender_id,
      content: data.content,
      created_at: data.timestamp
    })
    
    // Scroll to bottom for new messages
    nextTick(() => scrollChatToBottom())
    
    // Show toast for messages from others (only if chat tab is not active)
    if (data.sender_id !== user.id && activeTab.value !== 'chat') {
      showSuccess('New group message received')
    }
  }
}

// Watch for tab changes to load chat when switching to chat tab
watch(activeTab, async (newTab) => {
  if (newTab === 'chat' && isMember.value) {
    await loadChatMessages()
  }
})

onMounted(async () => {
  await loadGroup()
  if (isMember.value) {
    loadMembers()
    loadEvents()
    loadGroupPosts()
    
    // Set up WebSocket listener for group messages
    wsOn('group_message', handleGroupMessage)
  }
  if (isCreator.value) {
    loadRequests()
  }
})

onUnmounted(() => {
  // Clean up WebSocket listener
  wsOff('group_message', handleGroupMessage)
})
</script>

<style scoped>
.group-view {
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.loading-state,
.error-state {
  text-align: center;
  padding: 3rem 1rem;
  color: var(--text-muted);
}

.error-state button {
  margin-top: 1rem;
}

.group-header {
  border-radius: 1.75rem;
  padding: 2rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: linear-gradient(120deg, rgba(0, 247, 255, 0.1), rgba(255, 0, 230, 0.08));
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 2rem;
  flex-wrap: wrap;
}

.group-banner {
  display: flex;
  gap: 1.5rem;
  align-items: flex-start;
  flex: 1;
}

.group-icon-large {
  width: 100px;
  height: 100px;
  border-radius: 1.5rem;
  background: linear-gradient(135deg, var(--neon-cyan), var(--neon-pink));
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 2rem;
  color: #05060d;
  flex-shrink: 0;
  position: relative;
  overflow: hidden;
}

.group-icon-large img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.group-icon-large span {
  display: block;
}

.avatar-upload-btn {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 2rem;
  height: 2rem;
  border-radius: 50%;
  border: 2px solid rgba(5, 6, 13, 0.9);
  background: rgba(0, 247, 255, 0.9);
  color: #05060d;
  font-size: 1rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  z-index: 2;
}

.avatar-upload-btn:hover {
  background: var(--neon-cyan);
  transform: scale(1.1);
  box-shadow: 0 0 15px rgba(0, 247, 255, 0.5);
}

.group-info h1 {
  margin: 0 0 0.5rem 0;
  font-size: 2rem;
  color: var(--neon-cyan);
}

.group-description {
  color: var(--text-muted);
  margin: 0 0 1rem 0;
  line-height: 1.6;
}

.group-stats {
  display: flex;
  gap: 1.5rem;
  flex-wrap: wrap;
}

.group-stats span {
  color: var(--text-muted);
  font-size: 0.9rem;
}

.group-actions {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.group-tabs {
  display: flex;
  gap: 0.5rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  padding-bottom: 0.5rem;
}

.tab-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.25rem;
  border-radius: 0.9rem 0.9rem 0 0;
  border: none;
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
}

.tab-btn.active {
  background: rgba(0, 247, 255, 0.1);
  color: var(--neon-cyan);
  border-bottom: 2px solid var(--neon-cyan);
}

.tab-btn .badge {
  position: absolute;
  top: 0.25rem;
  right: 0.25rem;
  background: var(--neon-pink);
  color: #05060d;
  font-size: 0.65rem;
  font-weight: 700;
  border-radius: 999px;
  padding: 0.15rem 0.45rem;
}

.tab-content {
  border-radius: 1.5rem;
  padding: 2rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(5, 6, 13, 0.85);
}

.create-post-section {
  margin-bottom: 2rem;
  padding: 1.5rem;
  border-radius: 1.25rem;
  background: rgba(8, 10, 24, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.06);
}

.create-post-section h3 {
  margin: 0 0 1rem 0;
  color: var(--neon-cyan);
}

.create-post-section textarea {
  width: 100%;
  padding: 1rem;
  border-radius: 0.9rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: rgba(8, 9, 20, 0.8);
  color: inherit;
  font-family: inherit;
  resize: vertical;
  margin-bottom: 1rem;
}

.post-stack {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.post-card {
  padding: 1.5rem;
  border-radius: 1.25rem;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(8, 10, 24, 0.8);
  box-shadow: 0 12px 25px rgba(0, 0, 0, 0.35);
}

.post-card header {
  display: flex;
  align-items: center;
  margin-bottom: 0.85rem;
}

.author {
  display: flex;
  gap: 0.85rem;
  align-items: center;
}

.avatar {
  width: 2.75rem;
  height: 2.75rem;
  border-radius: 999px;
  border: 2px solid rgba(255, 255, 255, 0.08);
  background: linear-gradient(135deg, var(--neon-cyan), var(--neon-pink));
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 0.9rem;
  color: #05060d;
}

.author small {
  color: var(--text-muted);
}

.post-content {
  line-height: 1.6;
  margin: 0;
  word-wrap: break-word;
}

.event-card {
  padding: 1.5rem;
  border-radius: 1.25rem;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(8, 10, 24, 0.8);
  margin-bottom: 1.25rem;
}

.event-card h3 {
  margin: 0 0 0.5rem 0;
  color: var(--neon-cyan);
}

.event-description {
  color: var(--text-muted);
  margin: 0 0 1rem 0;
  line-height: 1.6;
}

.event-meta {
  display: flex;
  gap: 1.5rem;
  margin-bottom: 1.5rem;
  flex-wrap: wrap;
}

.event-date,
.event-creator {
  color: var(--text-muted);
  font-size: 0.9rem;
}

.rsvp-section {
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  padding-top: 1rem;
}

.rsvp-counts {
  display: flex;
  gap: 1.5rem;
  margin-bottom: 1rem;
  font-size: 0.9rem;
}

.going-count {
  color: var(--neon-cyan);
}

.not-going-count {
  color: var(--neon-pink);
}

.rsvp-buttons {
  display: flex;
  gap: 0.75rem;
}

.rsvp-btn {
  flex: 1;
  padding: 0.75rem 1.25rem;
  border-radius: 0.9rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: rgba(8, 9, 20, 0.8);
  color: inherit;
  cursor: pointer;
  transition: all 0.2s;
}

.rsvp-btn.active.going {
  background: rgba(0, 247, 255, 0.2);
  border-color: var(--neon-cyan);
  color: var(--neon-cyan);
}

.rsvp-btn.active.not-going {
  background: rgba(255, 0, 230, 0.2);
  border-color: var(--neon-pink);
  color: var(--neon-pink);
}

.rsvp-btn:hover:not(:disabled) {
  transform: translateY(-2px);
}

.rsvp-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.members-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 1rem;
}

.member-card {
  padding: 1.25rem;
  border-radius: 1rem;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(8, 10, 24, 0.8);
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}

.member-card:hover {
  background: rgba(8, 10, 24, 0.95);
  border-color: rgba(0, 247, 255, 0.3);
  transform: translateY(-2px);
}

.member-avatar,
.user-avatar {
  width: 3.5rem;
  height: 3.5rem;
  border-radius: 999px;
  background: linear-gradient(135deg, var(--neon-cyan), var(--neon-pink));
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 1.1rem;
  color: #05060d;
}

.member-name-btn {
  background: transparent;
  border: none;
  padding: 0;
  margin: 0;
  color: inherit;
  font: inherit;
  font-weight: 700;
  cursor: pointer;
  text-align: left;
}

.member-name-btn:hover {
  color: var(--neon-cyan);
}

.member-role {
  font-size: 0.8rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.request-card {
  padding: 1.25rem;
  border-radius: 1rem;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(8, 10, 24, 0.8);
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
}

.request-info {
  flex: 1;
}

.request-actions {
  display: flex;
  gap: 0.5rem;
}

.empty-state,
.loading {
  text-align: center;
  padding: 3rem 1rem;
  color: var(--text-muted);
}

.cta {
  border: none;
  border-radius: 999px;
  padding: 0.6rem 1.4rem;
  background: linear-gradient(120deg, var(--neon-cyan), var(--neon-pink));
  color: #05060d;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.cta:hover:not(:disabled) {
  box-shadow: 0 10px 25px rgba(255, 0, 230, 0.3);
  transform: translateY(-2px);
}

.cta:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.cta.mini {
  padding: 0.4rem 1rem;
  font-size: 0.85rem;
}

.ghost {
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: inherit;
  border-radius: 999px;
  padding: 0.6rem 1.4rem;
  cursor: pointer;
  transition: all 0.2s;
}

.ghost:hover:not(:disabled) {
  border-color: rgba(255, 255, 255, 0.4);
}

.ghost:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.ghost.mini {
  padding: 0.4rem 1rem;
  font-size: 0.85rem;
}

.mb-2 {
  margin-bottom: 1.25rem;
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-content {
  background: rgba(5, 6, 13, 0.98);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 1.5rem;
  padding: 2rem;
  max-width: 500px;
  width: 100%;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  max-height: 80vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.5rem;
}

.modal-header h3 {
  margin: 0;
  color: var(--neon-cyan);
}

.close-btn {
  background: transparent;
  border: none;
  color: inherit;
  font-size: 1.5rem;
  cursor: pointer;
  opacity: 0.7;
  transition: opacity 0.2s;
}

.close-btn:hover {
  opacity: 1;
}

.form-field {
  margin-bottom: 1.25rem;
}

.form-field label {
  display: block;
  margin-bottom: 0.5rem;
  font-size: 0.9rem;
  color: var(--text-muted);
}

.form-field input,
.form-field textarea {
  width: 100%;
  padding: 0.75rem 1rem;
  border-radius: 0.9rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: rgba(8, 9, 20, 0.8);
  color: inherit;
  font-family: inherit;
  resize: vertical;
}

.form-field input[type="datetime-local"] {
  color-scheme: dark;
}

.form-field input:focus,
.form-field textarea:focus {
  outline: none;
  border-color: var(--neon-cyan);
  box-shadow: 0 0 0 3px rgba(0, 247, 255, 0.15);
}

.error-message {
  color: var(--neon-pink);
  font-size: 0.9rem;
  margin-bottom: 1rem;
}

.modal-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
  margin-top: 1.5rem;
}

.invite-search {
  margin-bottom: 1rem;
}

.invite-search input {
  width: 100%;
  padding: 0.75rem 1rem;
  border-radius: 0.9rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: rgba(8, 9, 20, 0.8);
  color: inherit;
}

.invite-results {
  max-height: 400px;
  overflow-y: auto;
}

.invite-user-card {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  border-radius: 0.9rem;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(8, 10, 24, 0.6);
  margin-bottom: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}

.invite-user-card:hover {
  background: rgba(8, 10, 24, 0.9);
  border-color: rgba(0, 247, 255, 0.3);
}

@media (max-width: 768px) {
  .group-header {
    flex-direction: column;
  }

  .group-banner {
    flex-direction: column;
  }

  .group-actions {
    width: 100%;
  }

  .group-actions button {
    flex: 1;
  }

  .members-grid {
    grid-template-columns: 1fr;
  }
}

/* Chat Tab Styles */
.chat-tab {
  height: 600px;
  display: flex;
  flex-direction: column;
}

.chat-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  border-radius: 1.25rem;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(8, 10, 24, 0.8);
  overflow: hidden;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.chat-message {
  display: flex;
  gap: 0.75rem;
  align-items: flex-start;
}

.message-avatar {
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 999px;
  background: linear-gradient(135deg, var(--neon-cyan), var(--neon-pink));
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.message-avatar .initials {
  font-weight: 700;
  font-size: 0.9rem;
  color: #05060d;
}

.message-content {
  flex: 1;
  background: rgba(8, 9, 20, 0.6);
  padding: 0.75rem 1rem;
  border-radius: 0.9rem;
  border: 1px solid rgba(255, 255, 255, 0.06);
}

.message-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.message-header strong {
  color: var(--neon-cyan);
  font-size: 0.9rem;
}

.message-header small {
  color: var(--text-muted);
  font-size: 0.75rem;
}

.message-content p {
  margin: 0;
  line-height: 1.5;
  word-wrap: break-word;
}

.chat-input-area {
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  padding: 1rem;
  display: flex;
  gap: 0.75rem;
  align-items: flex-end;
  position: relative;
}

.chat-input-area textarea {
  flex: 1;
  padding: 0.75rem 1rem;
  border-radius: 0.9rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: rgba(8, 9, 20, 0.8);
  color: inherit;
  font-family: inherit;
  resize: none;
}

.chat-input-area textarea:focus {
  outline: none;
  border-color: var(--neon-cyan);
  box-shadow: 0 0 0 3px rgba(0, 247, 255, 0.15);
}

.chat-input-buttons {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.emoji-btn {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  font-size: 18px;
  cursor: pointer;
  width: 36px;
  height: 36px;
  border-radius: 0.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.emoji-btn:hover {
  background: rgba(0, 247, 255, 0.15);
  border-color: var(--border-glow);
  box-shadow: 0 0 12px rgba(0, 247, 255, 0.3);
}

.chat-input-area .emoji-picker {
  bottom: 60px;
  right: 60px;
}

.posts-feed {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.post-card {
  padding: 1.5rem;
  border-radius: 1.25rem;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(8, 10, 24, 0.8);
  box-shadow: 0 12px 25px rgba(0, 0, 0, 0.35);
  backdrop-filter: blur(12px);
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}

.post-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 16px 30px rgba(0, 0, 0, 0.4);
}

.post-header {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  margin-bottom: 0.85rem;
}

.post-header .avatar {
  width: 2.75rem;
  height: 2.75rem;
  border-radius: 999px;
  border: 2px solid rgba(255, 255, 255, 0.08);
  background: linear-gradient(135deg, var(--neon-cyan), var(--neon-pink));
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 0.9rem;
  color: #05060d;
  flex-shrink: 0;
}

.post-header .avatar img {
  width: 100%;
  height: 100%;
  border-radius: 999px;
  object-fit: cover;
}

.post-header .initials {
  font-weight: 700;
  font-size: 0.9rem;
  color: #05060d;
}

.author-info {
  flex: 1;
}

.author-info h3 {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
}

.author-info .meta {
  color: var(--text-muted);
  font-size: 0.85rem;
  margin: 0.25rem 0 0 0;
}

.privacy-badge {
  padding: 0.15rem 0.5rem;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.08);
  font-size: 0.75rem;
  margin-left: 0.5rem;
}

.menu-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 1.5rem;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: 0.5rem;
  transition: background 0.2s;
}

.menu-btn:hover {
  background: rgba(255, 255, 255, 0.08);
}

.post-title {
  font-size: 1.1rem;
  font-weight: 600;
  margin: 0 0 0.5rem 0;
  color: var(--neon-cyan);
}

.post-content {
  line-height: 1.6;
  margin: 0 0 1rem 0;
  word-wrap: break-word;
  word-break: break-word;
  overflow-wrap: break-word;
  white-space: pre-wrap;
  max-width: 100%;
  overflow: hidden;
}

.post-image {
  width: 100%;
  max-height: 500px;
  object-fit: cover;
  border-radius: 0.75rem;
  margin-top: 1rem;
}
</style>
