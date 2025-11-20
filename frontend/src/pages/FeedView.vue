<template>
  <section class="main-panel">
    <!-- Create Post Component -->
    <CreatePost @posted="loadPosts" />
    
    <div class="filters-row">
      <button
        :class="['filter-btn', { active: activeView === 'feed' }]"
        @click="activeView = 'feed'"
      >
        <span class="icon icon-home" />
        Feed
      </button>
      <button
        :class="['filter-btn', { active: activeView === 'search' }]"
        @click="activeView = 'search'"
      >
        <span class="icon icon-search" />
        Search
      </button>
    </div>
    
    <div v-if="activeView === 'search'" class="search-pane">
      <div class="search-input">
        <span class="icon icon-search" />
        <input 
          v-model="searchQuery" 
          @input="handleSearch"
          placeholder="Search posts by content or title..." 
        />
      </div>
      
      <!-- Search Results -->
      <div v-if="searchQuery.trim()" class="search-results">
        <p class="results-header">Search Results ({{ searchResults.length }})</p>
        <div v-if="searchingPosts" class="loading">Searching...</div>
        <div v-else-if="searchResults.length === 0" class="empty-state">
          <p>No posts found matching "{{ searchQuery }}"</p>
        </div>
        <div v-else class="post-stack">
          <article class="post-card" v-for="post in searchResults" :key="post.id" @click="navigateToPost(post.id)">
            <header>
              <div class="author">
                <div class="avatar">{{ getInitials(post.author) }}</div>
                <div>
                  <strong>{{ post.author.first_name }} {{ post.author.last_name }}</strong>
                  <small>{{ formatTime(post.created_at) }} · {{ formatPrivacy(post.privacy_level) }}</small>
                </div>
              </div>
              <button class="ghost" @click.stop>•••</button>
            </header>
            <h3 v-if="post.title" class="post-title">{{ post.title }}</h3>
            <p class="post-content">{{ post.content }}</p>
            <img v-if="post.image_path" :src="getImageUrl(post.image_path)" class="post-image" alt="Post image" />
          </article>
        </div>
      </div>
    </div>

    <TransitionGroup name="post" tag="div" class="post-stack" v-else>
      <div v-if="loading" key="loading" class="loading">Loading posts...</div>
      <div v-else-if="posts.length === 0" key="empty" class="empty-state">
        <p>No posts yet. Be the first to post!</p>
      </div>
      <article v-else class="post-card" v-for="post in posts" :key="post.id" @click="navigateToPost(post.id)">
        <header>
          <div class="author">
            <div class="avatar">{{ getInitials(post.author) }}</div>
            <div>
              <strong>{{ post.author.first_name }} {{ post.author.last_name }}</strong>
              <small>{{ formatTime(post.created_at) }} · {{ formatPrivacy(post.privacy_level) }}</small>
            </div>
          </div>
          <button class="ghost" @click.stop>•••</button>
        </header>
        <h3 v-if="post.title" class="post-title">{{ post.title }}</h3>
        <p class="post-content">{{ post.content }}</p>
        <img v-if="post.image_path" :src="getImageUrl(post.image_path)" class="post-image" alt="Post image" />
      </article>
    </TransitionGroup>
  </section>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import CreatePost from '@/components/CreatePost.vue'
import { getToken } from '@/stores/auth'

const router = useRouter()
const activeView = ref('feed')
const searchQuery = ref('')
const posts = ref([])
const searchResults = ref([])
const loading = ref(false)
const searchingPosts = ref(false)
let searchTimeout = null

const POSTS_API_URL = import.meta.env.VITE_POSTS_API_URL || 'http://localhost:8083'

async function loadPosts() {
  loading.value = true
  try {
    const token = getToken()
    if (!token) {
      console.log('No token, skipping post load')
      posts.value = []
      loading.value = false
      return
    }

    console.log('Loading posts from:', `${POSTS_API_URL}/posts/feed`)
    const response = await axios.get(`${POSTS_API_URL}/posts/feed`, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    })

    console.log('Posts response:', response.data)
    posts.value = response.data.data.posts || []
    console.log('Loaded posts:', posts.value.length)
  } catch (error) {
    console.error('Failed to load posts:', error.response?.data || error.message)
    posts.value = []
  } finally {
    loading.value = false
  }
}

function getInitials(author) {
  if (!author) return '?'
  if (author.first_name && author.last_name) {
    return `${author.first_name[0]}${author.last_name[0]}`.toUpperCase()
  }
  if (author.first_name) return author.first_name[0].toUpperCase()
  return author.username?.[0]?.toUpperCase() || '?'
}

function formatTime(timestamp) {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  const now = new Date()
  const diff = Math.floor((now - date) / 1000) // seconds

  if (diff < 60) return `${diff}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  return `${Math.floor(diff / 86400)}d ago`
}

function formatPrivacy(privacy) {
  const map = {
    'public': 'Public',
    'almost_private': 'Followers',
    'private': 'Private'
  }
  return map[privacy] || privacy
}

function getImageUrl(path) {
  if (!path) return ''
  if (path.startsWith('http')) return path
  return `${POSTS_API_URL}${path}`
}

async function searchPosts(query) {
  if (!query.trim()) {
    searchResults.value = []
    return
  }

  searchingPosts.value = true
  try {
    const token = getToken()
    if (!token) {
      console.log('No token, skipping search')
      searchResults.value = []
      searchingPosts.value = false
      return
    }

    const response = await axios.get(`${POSTS_API_URL}/posts/search`, {
      params: { q: query },
      headers: {
        Authorization: `Bearer ${token}`
      }
    })

    searchResults.value = response.data.data.posts || []
  } catch (error) {
    console.error('Failed to search posts:', error.response?.data || error.message)
    searchResults.value = []
  } finally {
    searchingPosts.value = false
  }
}

function handleSearch() {
  // Debounce search to avoid too many API calls
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    searchPosts(searchQuery.value)
  }, 300)
}

function navigateToPost(postId) {
  router.push(`/post/${postId}`)
}

const suggestions = [
  { name: 'Neon District', handle: '@neond', avatar: 'https://placehold.co/56x56/15162a/fff?text=ND' },
  { name: 'Circuit Club', handle: '@circuit', avatar: 'https://placehold.co/56x56/181931/fff?text=CC' },
  { name: 'Glitch Bloom', handle: '@glitch', avatar: 'https://placehold.co/56x56/121325/fff?text=GB' },
]

const filteredSuggestions = computed(() => {
  if (!searchQuery.value) return suggestions
  return suggestions.filter((user) => user.name.toLowerCase().includes(searchQuery.value.toLowerCase()))
})

onMounted(() => {
  loadPosts()
})
</script>

<style scoped>
.main-panel {
  width: min(900px, 100%);
  border-radius: 1.5rem;
  padding: clamp(1rem, 3vw, 2rem);
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(5, 6, 13, 0.85);
  box-shadow: var(--shadow);
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.filters-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 0.85rem;
}

.filter-btn {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  padding: 0.85rem 1rem;
  border-radius: 1rem;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(7, 9, 20, 0.6);
  color: inherit;
  cursor: pointer;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  font-weight: 600;
  transition: background 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease;
}

.filter-btn.active {
  border-color: var(--border-glow);
  box-shadow: 0 0 22px rgba(0, 247, 255, 0.2);
  background: rgba(0, 247, 255, 0.1);
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
  backdrop-filter: blur(12px);
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}

.post-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 16px 30px rgba(0, 0, 0, 0.4);
}

.post-card header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.85rem;
}

.author {
  display: flex;
  gap: 0.85rem;
  align-items: center;
}

.author img {
  width: 2.75rem;
  height: 2.75rem;
  border-radius: 999px;
  border: 2px solid rgba(255, 255, 255, 0.08);
}

.author .avatar {
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

.post-title {
  font-size: 1.1rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
  color: var(--neon-cyan);
}

.post-content {
  line-height: 1.6;
  margin-bottom: 1rem;
}

.post-image {
  width: 100%;
  max-height: 500px;
  object-fit: cover;
  border-radius: 0.75rem;
  margin-top: 1rem;
}

.loading,
.empty-state {
  text-align: center;
  padding: 3rem 1rem;
  color: var(--text-muted);
  font-size: 1.1rem;
}

.ghost {
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.15);
  color: inherit;
  border-radius: 999px;
  padding: 0.35rem 0.85rem;
  cursor: pointer;
}

.search-pane {
  display: flex;
  flex-direction: column;
  gap: 1.2rem;
}

.search-input {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem 1.25rem;
  border-radius: 1rem;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(0, 0, 0, 0.45);
}

.search-input input {
  background: transparent;
  border: none;
  flex: 1;
  color: inherit;
  font-size: 1rem;
}

.search-input input:focus {
  outline: none;
}

.suggestions {
  border-radius: 1.25rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
  padding: 1.25rem;
  background: rgba(8, 10, 24, 0.75);
}

.suggestion-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 0.85rem;
  padding: 0.85rem;
  border-radius: 1rem;
  background: rgba(10, 12, 22, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.suggestion-card .user {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.suggestion-card img {
  width: 3rem;
  height: 3rem;
  border-radius: 999px;
}

.icon {
  display: inline-block;
  width: 1.1rem;
  height: 1.1rem;
}

.icon-globe {
  background: radial-gradient(circle, var(--neon-cyan), rgba(0, 0, 0, 0));
  mask: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="white" d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm6.93 8h-3.54a16.06 16.06 0 0 0-1.05-4.29A8.026 8.026 0 0 1 18.93 10zM12 4c.83 0 2.36 2.17 2.82 6H9.18C9.64 6.17 11.17 4 12 4zm-3.34.71A16.06 16.06 0 0 0 7.61 10H4.07a8.026 8.026 0 0 1 4.59-5.29zM4.07 14h3.54c.21 1.49.67 2.98 1.39 4.29A8.026 8.026 0 0 1 4.07 14zm4.75 0h6.36c-.46 3.83-1.99 6-2.82 6s-2.36-2.17-2.82-6zm7.57 4.29c.72-1.31 1.18-2.8 1.39-4.29h3.54a8.026 8.026 0 0 1-4.93 4.29z"/></svg>')
    center / contain no-repeat;
}

.icon-home {
  background: radial-gradient(circle, var(--neon-pink), rgba(0, 0, 0, 0));
  mask: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="white" d="m12 3l9 8h-3v9h-5v-5H11v5H6v-9H3l9-8z"/></svg>')
    center / contain no-repeat;
}

.icon-search {
  background: radial-gradient(circle, var(--neon-cyan), rgba(0, 0, 0, 0));
  mask: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="white" d="m15.5 14h-.79l-.28-.27A6.471 6.471 0 0 0 16 9.5 6.5 6.5 0 1 0 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79L20 21.49 21.49 20 15.5 14zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"/></svg>')
    center / contain no-repeat;
}

@media (max-width: 960px) {
  .main-panel {
    width: 100%;
  }
}

@media (max-width: 640px) {
  .filters-row {
    grid-template-columns: 1fr;
  }
}
</style>
