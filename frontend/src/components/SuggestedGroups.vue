<template>
  <aside class="suggested-groups">
    <header class="sidebar-header">
      <h3>Groups</h3>
      <button class="icon-btn" @click="showCreateModal = true" title="Create Group">
        <span>+</span>
      </button>
    </header>

    <!-- Search Groups -->
    <div class="search-box">
      <span class="icon-search">🔍</span>
      <input
        v-model="searchQuery"
        @input="handleSearch"
        type="search"
        placeholder="Search groups..."
        autocomplete="off"
      />
    </div>

    <!-- My Groups -->
    <section v-if="myGroups.length > 0" class="groups-section">
      <h4 class="section-title">My Groups</h4>
      <div class="groups-list">
        <article
          v-for="group in myGroups"
          :key="group.id"
          class="group-card"
          @click="navigateToGroup(group.id)"
        >
          <div class="group-icon">
            <img v-if="group.image_url" :src="getGroupImageUrl(group.image_url)" alt="Group avatar" />
            <span v-else>{{ getGroupInitials(group.name) }}</span>
          </div>
          <div class="group-info">
            <strong>{{ group.name }}</strong>
            <small>{{ group.member_count || 0 }} members</small>
          </div>
        </article>
      </div>
    </section>

    <!-- Suggested Groups -->
    <section class="groups-section">
      <h4 class="section-title">{{ searchQuery ? 'Search Results' : 'Discover' }}</h4>
      <div v-if="loading" class="loading-state">Loading...</div>
      <div v-else-if="displayedGroups.length === 0" class="empty-state">
        <p>{{ searchQuery ? 'No groups found' : 'No groups available' }}</p>
      </div>
      <div v-else class="groups-list">
        <article
          v-for="group in displayedGroups"
          :key="group.id"
          class="group-card"
          @click="navigateToGroup(group.id)"
        >
          <div class="group-icon">
            <img v-if="group.image_url" :src="getGroupImageUrl(group.image_url)" alt="Group avatar" />
            <span v-else>{{ getGroupInitials(group.name) }}</span>
          </div>
          <div class="group-info">
            <strong>{{ group.name }}</strong>
            <small>{{ group.member_count || 0 }} members</small>
          </div>
        </article>
      </div>
    </section>

    <!-- Create Group Modal -->
    <Teleport to="body">
      <div v-if="showCreateModal" class="modal-overlay" @click="closeCreateModal">
        <div class="modal-content" @click.stop>
          <header class="modal-header">
            <h3>Create New Group</h3>
            <button class="close-btn" @click="closeCreateModal">✕</button>
          </header>
          <form @submit.prevent="createGroup">
            <div class="form-field">
              <label for="group-name">Group Name *</label>
              <input
                id="group-name"
                v-model="newGroup.name"
                type="text"
                placeholder="Enter group name..."
                required
                maxlength="100"
              />
            </div>
            <div class="form-field">
              <label for="group-desc">Description *</label>
              <textarea
                id="group-desc"
                v-model="newGroup.description"
                placeholder="Describe your group..."
                required
                rows="4"
                maxlength="500"
              />
            </div>
            <div class="form-field">
              <label for="group-image">Group Image (optional)</label>
              <input
                id="group-image"
                type="file"
                accept="image/*"
                @change="handleImageChange"
                ref="imageInput"
              />
              <div v-if="imagePreview" class="image-preview">
                <img :src="imagePreview" alt="Preview" />
                <button type="button" class="remove-image" @click="removeImage">✕</button>
              </div>
            </div>
            <p v-if="createError" class="error-message">{{ createError }}</p>
            <div class="modal-actions">
              <button type="button" class="ghost" @click="closeCreateModal">Cancel</button>
              <button type="submit" class="cta" :disabled="creating">
                {{ creating ? 'Creating...' : 'Create Group' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>
  </aside>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getToken } from '@/stores/auth'
import { getAllGroups, getMyGroups, createGroup as createGroupService, searchGroups } from '@/services/groupsService'

const router = useRouter()
const searchQuery = ref('')
const myGroups = ref([])
const suggestedGroups = ref([])
const loading = ref(false)
const showCreateModal = ref(false)
const creating = ref(false)
const createError = ref('')

const newGroup = ref({
  name: '',
  description: ''
})

const imagePreview = ref('')
const imageFile = ref(null)
const imageInput = ref(null)

let searchTimeout = null

const displayedGroups = computed(() => {
  if (searchQuery.value.trim()) {
    return suggestedGroups.value
  }
  // Show only groups user is not a member of
  return suggestedGroups.value.filter(g => !myGroups.value.some(mg => mg.id === g.id))
})

function getGroupInitials(name) {
  if (!name) return '?'
  const words = name.split(' ')
  if (words.length >= 2) {
    return (words[0][0] + words[1][0]).toUpperCase()
  }
  return name.substring(0, 2).toUpperCase()
}

function navigateToGroup(groupId) {
  router.push(`/groups/${groupId}`)
}

async function loadMyGroups() {
  const token = getToken()
  if (!token) return

  try {
    const { groups = [] } = await getMyGroups(token)
    myGroups.value = groups
  } catch (error) {
    console.error('Failed to load my groups:', error)
  }
}

async function loadSuggestedGroups() {
  const token = getToken()
  if (!token) return

  loading.value = true
  try {
    const { groups = [] } = await getAllGroups(token)
    suggestedGroups.value = groups.slice(0, 8) // Show top 8 suggested
  } catch (error) {
    console.error('Failed to load suggested groups:', error)
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(async () => {
    const query = searchQuery.value.trim()
    if (!query) {
      loadSuggestedGroups()
      return
    }

    const token = getToken()
    if (!token) return

    loading.value = true
    try {
      const { groups = [] } = await searchGroups(query, token)
      suggestedGroups.value = groups
    } catch (error) {
      console.error('Failed to search groups:', error)
    } finally {
      loading.value = false
    }
  }, 300)
}

function handleImageChange(event) {
  const file = event.target.files?.[0]
  if (!file) return

  if (!file.type.startsWith('image/')) {
    createError.value = 'Please select an image file'
    return
  }

  if (file.size > 5 * 1024 * 1024) {
    createError.value = 'Image must be less than 5MB'
    return
  }

  imageFile.value = file
  const reader = new FileReader()
  reader.onload = (e) => {
    imagePreview.value = e.target.result
  }
  reader.readAsDataURL(file)
  createError.value = ''
}

function removeImage() {
  imageFile.value = null
  imagePreview.value = ''
  if (imageInput.value) {
    imageInput.value.value = ''
  }
}

async function createGroup() {
  const token = getToken()
  if (!token) {
    createError.value = 'You must be logged in to create a group'
    return
  }

  creating.value = true
  createError.value = ''

  try {
    const formData = new FormData()
    formData.append('name', newGroup.value.name)
    formData.append('description', newGroup.value.description)
    if (imageFile.value) {
      formData.append('image', imageFile.value)
    }

    const { group } = await createGroupService(formData, token)
    showCreateModal.value = false
    newGroup.value = { name: '', description: '' }
    removeImage()
    await loadMyGroups()
    await loadSuggestedGroups()
    router.push(`/groups/${group.id}`)
  } catch (error) {
    console.error('Failed to create group:', error)
    createError.value = error.message || 'Failed to create group'
  } finally {
    creating.value = false
  }
}

function closeCreateModal() {
  showCreateModal.value = false
  newGroup.value = { name: '', description: '' }
  removeImage()
  createError.value = ''
}

function getGroupImageUrl(imageUrl) {
  if (!imageUrl) return ''
  if (imageUrl.startsWith('http://') || imageUrl.startsWith('https://')) {
    return imageUrl
  }
  // Assume it's served from the groups service
  const GROUPS_API_URL = import.meta.env.VITE_GROUPS_API_URL || 'http://localhost:8084'
  return `${GROUPS_API_URL}/${imageUrl}`
}

onMounted(() => {
  loadMyGroups()
  loadSuggestedGroups()
})
</script>

<style scoped>
.suggested-groups {
  position: sticky;
  top: 1rem;
  width: min(280px, 100%);
  border-radius: 1.5rem;
  padding: 1.5rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(5, 6, 13, 0.85);
  box-shadow: var(--shadow);
  display: flex;
  flex-direction: column;
  gap: 1rem;
  max-height: calc(100vh - 2rem);
  overflow-y: auto;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}

.sidebar-header h3 {
  margin: 0;
  font-size: 1.3rem;
  color: var(--neon-cyan);
}

.icon-btn {
  width: 2rem;
  height: 2rem;
  border-radius: 50%;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: rgba(8, 10, 22, 0.7);
  color: var(--neon-cyan);
  font-size: 1.3rem;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
}

.icon-btn:hover {
  background: rgba(0, 247, 255, 0.1);
  border-color: var(--neon-cyan);
}

.search-box {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.7rem 1rem;
  border-radius: 0.9rem;
  border: 1px solid rgba(255, 255, 255, 0.15);
  background: rgba(8, 9, 20, 0.8);
}

.icon-search {
  font-size: 1rem;
  opacity: 0.6;
}

.search-box input {
  flex: 1;
  background: transparent;
  border: none;
  color: inherit;
  outline: none;
  font-size: 0.9rem;
}

.groups-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.section-title {
  margin: 0;
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--text-muted);
}

.groups-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.group-card {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  border-radius: 0.9rem;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(8, 10, 24, 0.6);
  cursor: pointer;
  transition: all 0.2s;
}

.group-card:hover {
  background: rgba(8, 10, 24, 0.9);
  border-color: rgba(0, 247, 255, 0.3);
  transform: translateX(3px);
}

.group-icon {
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 0.7rem;
  background: linear-gradient(135deg, var(--neon-cyan), var(--neon-pink));
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 0.8rem;
  color: #05060d;
  flex-shrink: 0;
  overflow: hidden;
}

.group-icon img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.group-icon span {
  display: block;
}

.group-info {
  flex: 1;
  min-width: 0;
}

.group-info strong {
  display: block;
  font-size: 0.9rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.group-info small {
  display: block;
  font-size: 0.75rem;
  color: var(--text-muted);
  margin-top: 0.1rem;
}

.loading-state,
.empty-state {
  text-align: center;
  padding: 1.5rem 0.5rem;
  color: var(--text-muted);
  font-size: 0.85rem;
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

.form-field input:focus,
.form-field textarea:focus {
  outline: none;
  border-color: var(--neon-cyan);
  box-shadow: 0 0 0 3px rgba(0, 247, 255, 0.15);
}

.form-field input[type="file"] {
  padding: 0.5rem;
  cursor: pointer;
}

.image-preview {
  position: relative;
  margin-top: 1rem;
  border-radius: 0.75rem;
  overflow: hidden;
  width: 150px;
  height: 150px;
  border: 2px solid rgba(0, 247, 255, 0.3);
}

.image-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.remove-image {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  width: 1.5rem;
  height: 1.5rem;
  border-radius: 50%;
  border: none;
  background: rgba(255, 0, 230, 0.9);
  color: white;
  font-size: 0.8rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.remove-image:hover {
  background: var(--neon-pink);
  transform: scale(1.1);
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

.ghost:hover {
  border-color: rgba(255, 255, 255, 0.4);
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

@media (max-width: 1200px) {
  .suggested-groups {
    position: relative;
    width: 100%;
    max-height: none;
  }
}
</style>
