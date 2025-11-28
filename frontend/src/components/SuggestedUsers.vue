<template>
  <div class="suggested-users-dropdown" @click.stop>
    <div class="widget-header">
      <span class="icon icon-users"></span>
      <h3>{{ searchQuery ? 'Search Results' : 'People to Follow' }}</h3>
    </div>

    <!-- Search Input -->
    <div class="search-box">
      <input
        v-model="searchQuery"
        type="text"
        placeholder="Search users..."
        class="search-input"
        @input="onSearchInput"
      />
      <span v-if="searchQuery" class="clear-search" @click="clearSearch">×</span>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>{{ searchQuery ? 'Searching...' : 'Finding people...' }}</p>
    </div>

    <div v-else-if="error" class="error-state">
      <p>{{ error }}</p>
      <button class="ghost mini" @click="loadSuggestions">Retry</button>
    </div>

    <div v-else-if="suggestedUsers.length === 0" class="empty-state">
      <p>No suggestions available</p>
    </div>

    <div v-else class="users-list">
      <div v-for="user in suggestedUsers" :key="user.id" class="user-card">
        <div class="user-info">
          <img
            :src="getUserAvatarUrl(user, 48)"
            :alt="user.username"
            class="user-avatar"
          />
          <div class="user-details">
            <strong class="user-name">{{ user.first_name }} {{ user.last_name }}</strong>
            <small class="user-handle">@{{ user.username }}</small>
          </div>
        </div>
        <button
          class="follow-btn"
          :class="{ 
            following: user.isFollowing,
            pending: user.isPending
          }"
          @click="toggleFollow(user)"
          :disabled="user.actionLoading || user.isPending"
        >
          {{ 
            user.actionLoading ? '...' : 
            user.isPending ? 'Pending' :
            user.isFollowing ? 'Following' : 
            'Follow' 
          }}
        </button>
      </div>
    </div>

    <button v-if="suggestedUsers.length > 0" class="ghost mini full-width refresh-btn" @click="loadSuggestions">
      Refresh
    </button>
  </div>
</template>

<script setup>
import { onMounted, ref, watch } from 'vue'
import { getToken } from '../stores/auth'
import { searchUsers, followUser, unfollowUser } from '../services/usersService'
import { useAvatar } from '../composables/useAvatar'

const { getUserAvatarUrl } = useAvatar()
const suggestedUsers = ref([])
const loading = ref(false)
const error = ref('')
const searchQuery = ref('')
let searchTimeout = null

onMounted(() => {
  loadSuggestions()
})

// Debounced search (waits 400ms after user stops typing)
function onSearchInput() {
  // Clear previous timeout
  if (searchTimeout) {
    clearTimeout(searchTimeout)
  }

  // If search is empty, show suggestions
  if (!searchQuery.value.trim()) {
    loadSuggestions()
    return
  }

  // Wait 400ms before searching
  searchTimeout = setTimeout(() => {
    performSearch()
  }, 400)
}

function clearSearch() {
  searchQuery.value = ''
  loadSuggestions()
}

async function performSearch() {
  const query = searchQuery.value.trim()
  if (!query) return

  const token = getToken()
  if (!token) {
    error.value = 'Please log in to search'
    return
  }

  loading.value = true
  error.value = ''

  try {
    const response = await searchUsers(query, token)
    const allUsers = response.users || []
    
    // Backend already filters out accepted and pending follows
    // Just add UI state properties
    suggestedUsers.value = allUsers
      .slice(0, 10)
      .map(user => ({
        ...user,
        isFollowing: false,
        isPending: false,
        actionLoading: false
      }))
  } catch (err) {
    console.error('Search failed:', err)
    error.value = err.message || 'Search failed'
  } finally {
    loading.value = false
  }
}

async function loadSuggestions() {
  const token = getToken()
  if (!token) {
    error.value = 'Please log in to see suggestions'
    return
  }

  loading.value = true
  error.value = ''

  try {
    // Use a single letter to get a broad range of users
    // Alternating between different letters for variety
    const searchLetters = ['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm']
    const randomLetter = searchLetters[Math.floor(Math.random() * searchLetters.length)]
    const response = await searchUsers(randomLetter, token)
    
    // Backend already filters out accepted and pending follows
    // Filter and limit to 5 users, randomize order
    const allUsers = response.users || []
    const shuffled = allUsers.sort(() => 0.5 - Math.random())
    
    suggestedUsers.value = shuffled
      .slice(0, 5)
      .map(user => ({
        ...user,
        isFollowing: false,
        isPending: false,
        actionLoading: false
      }))
  } catch (err) {
    console.error('Failed to load suggestions:', err)
    error.value = err.message || 'Failed to load suggestions'
  } finally {
    loading.value = false
  }
}

async function toggleFollow(user) {
  const token = getToken()
  if (!token || user.actionLoading || user.isPending) return

  user.actionLoading = true

  try {
    if (user.isFollowing) {
      await unfollowUser(user.id, token)
      user.isFollowing = false
      user.isPending = false
    } else {
      await followUser(user.id, token)
      // After following, mark as pending (will be 'accepted' for public profiles)
      // The backend handles this automatically
      user.isPending = true
      // Notify other components (like Chat) that a follow happened
      window.dispatchEvent(new CustomEvent('follow-accepted'))
      
      // Remove user from suggestions after following
      setTimeout(() => {
        const index = suggestedUsers.value.findIndex(u => u.id === user.id)
        if (index > -1) {
          suggestedUsers.value.splice(index, 1)
        }
      }, 1000)
    }
  } catch (err) {
    console.error('Follow action failed:', err)
    error.value = err.message || 'Failed to update follow status'
    // Reset states on error
    user.isFollowing = false
    user.isPending = false
  } finally {
    user.actionLoading = false
  }
}
</script>

<style scoped>
.suggested-users-dropdown {
  position: absolute;
  top: calc(100% + 0.5rem);
  right: 0;
  background: rgba(6, 8, 18, 0.98);
  border: 1px solid rgba(0, 247, 255, 0.3);
  border-radius: 1.25rem;
  padding: 1.25rem;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.6), 0 0 20px rgba(0, 247, 255, 0.2);
  backdrop-filter: blur(20px);
  min-width: 320px;
  max-width: 380px;
  max-height: 80vh;
  overflow-y: auto;
  z-index: 100;
}

.widget-header {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  margin-bottom: 1rem;
  padding-bottom: 0.85rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.widget-header h3 {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  letter-spacing: 0.02em;
}

.search-box {
  position: relative;
  margin-bottom: 1rem;
}

.search-input {
  width: 100%;
  padding: 0.65rem 2rem 0.65rem 0.85rem;
  border-radius: 0.75rem;
  border: 1px solid rgba(255, 255, 255, 0.15);
  background: rgba(8, 10, 24, 0.6);
  color: inherit;
  font-size: 0.9rem;
  transition: border-color 0.2s ease;
}

.search-input:focus {
  outline: none;
  border-color: var(--neon-cyan);
}

.search-input::placeholder {
  color: var(--text-muted);
}

.clear-search {
  position: absolute;
  right: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  font-size: 1.5rem;
  color: var(--text-muted);
  cursor: pointer;
  transition: color 0.2s ease;
  line-height: 1;
}

.clear-search:hover {
  color: var(--neon-pink);
}

.icon {
  display: inline-block;
  width: 1.1rem;
  height: 1.1rem;
}

.icon-users {
  background: radial-gradient(circle, var(--neon-cyan), rgba(0, 0, 0, 0));
  mask: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="white" d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5s-3 1.34-3 3 1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/></svg>')
    center / contain no-repeat;
}

.loading-state,
.error-state,
.empty-state {
  text-align: center;
  padding: 1.5rem 0.5rem;
  color: var(--text-muted);
}

.spinner {
  width: 2rem;
  height: 2rem;
  border: 3px solid rgba(0, 247, 255, 0.1);
  border-top-color: var(--neon-cyan);
  border-radius: 50%;
  margin: 0 auto 0.75rem;
  animation: spin 0.8s linear infinite;
}

.users-list {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

.user-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem;
  border-radius: 0.9rem;
  background: rgba(8, 10, 24, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.05);
  transition: border-color 0.2s ease, transform 0.2s ease;
}

.user-card:hover {
  border-color: rgba(0, 247, 255, 0.2);
  transform: translateY(-1px);
}

.user-info {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  flex: 1;
  min-width: 0;
}

.user-avatar {
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid rgba(255, 255, 255, 0.08);
  flex-shrink: 0;
}

.user-details {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  min-width: 0;
}

.user-name {
  font-size: 0.9rem;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-handle {
  color: var(--text-muted);
  font-size: 0.8rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.follow-btn {
  border: none;
  border-radius: 999px;
  padding: 0.45rem 1rem;
  font-size: 0.8rem;
  font-weight: 600;
  letter-spacing: 0.03em;
  cursor: pointer;
  transition: all 0.2s ease;
  background: linear-gradient(120deg, var(--neon-cyan), var(--neon-pink));
  color: #05060d;
  box-shadow: 0 4px 12px rgba(0, 247, 255, 0.3);
  flex-shrink: 0;
}

.follow-btn:hover:not(:disabled) {
  transform: scale(1.05);
  box-shadow: 0 6px 16px rgba(0, 247, 255, 0.4);
}

.follow-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.follow-btn.following {
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: inherit;
  box-shadow: none;
}

.follow-btn.following:hover:not(:disabled) {
  border-color: rgba(255, 0, 230, 0.5);
  color: var(--neon-pink);
}

.follow-btn.pending {
  background: transparent;
  border: 1px solid rgba(255, 165, 0, 0.4);
  color: rgba(255, 165, 0, 0.9);
  box-shadow: none;
  cursor: not-allowed;
}

.follow-btn.pending:hover {
  border-color: rgba(255, 165, 0, 0.6);
}

.ghost.mini {
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.15);
  color: inherit;
  border-radius: 999px;
  padding: 0.4rem 0.85rem;
  font-size: 0.85rem;
  cursor: pointer;
  transition: border-color 0.2s ease;
}

.ghost.mini:hover {
  border-color: var(--neon-cyan);
}

.ghost.mini.full-width {
  width: 100%;
  margin-top: 0.75rem;
}

.refresh-btn {
  text-align: center;
  display: block;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 768px) {
  .suggested-users-widget {
    max-width: 100%;
    min-width: 240px;
  }
}
</style>
