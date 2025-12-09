<template>
  <div class="create-post-card" :class="{ collapsed: !isOpen }">
    <button class="post-header" type="button" @click="toggleOpen" :aria-expanded="isOpen">
      <h3>Create Post</h3>
      <span class="chevron" :class="{ open: isOpen }">▾</span>
    </button>

    <Transition name="slide-fade">
      <form v-show="isOpen" @submit.prevent="handleSubmit" class="post-form">
      <!-- Title input -->
      <input
        v-model="form.title"
        type="text"
        placeholder="Title (optional)"
        maxlength="200"
        class="title-input"
      />
      
      <!-- Content textarea -->
      <textarea
        v-model="form.content"
        placeholder="What's on your mind?"
        rows="4"
        required
        maxlength="5000"
      ></textarea>

      <!-- Privacy selector (hidden for group posts) -->
      <div v-if="!props.groupId" class="privacy-selector">
        <label>
          <span class="icon icon-lock"></span>
          <select v-model="form.privacy" @change="handlePrivacyChange">
            <option value="public">Public - Everyone can see</option>
            <option value="almost_private">Almost Private - Only followers</option>
            <option value="private">Private - Choose specific followers</option>
          </select>
        </label>
      </div>

      <!-- Follower selector for private posts (hidden for group posts) -->
      <div v-if="!props.groupId && form.privacy === 'private'" class="follower-selector">
        <p class="selector-label">Choose who can see this post:</p>
        <div v-if="loadingFollowers" class="loading">Loading followers...</div>
        <div v-else-if="followers.length === 0" class="no-followers">
          You don't have any followers yet
        </div>
        <div v-else class="followers-list">
          <label
            v-for="follower in followers"
            :key="follower.id"
            class="follower-item"
          >
            <input
              type="checkbox"
              :value="follower.id"
              v-model="form.selectedViewers"
            />
            <div class="follower-info">
              <div class="follower-avatar" :style="{ backgroundColor: getAvatarColor(follower.username) }">
                {{ getInitials(follower) }}
              </div>
              <div class="follower-details">
                <strong>{{ getDisplayName(follower) }}</strong>
                <small>@{{ follower.username }}</small>
              </div>
            </div>
          </label>
        </div>
      </div>

      <!-- Image preview -->
      <div v-if="imagePreview" class="image-preview">
        <img :src="imagePreview" alt="Preview" />
        <button type="button" @click="removeImage" class="remove-image">×</button>
      </div>

      <!-- Actions -->
      <div class="post-actions">
        <label class="upload-btn">
          <input
            type="file"
            accept="image/jpeg,image/jpg,image/png,image/gif"
            @change="handleImageSelect"
            ref="fileInput"
          />
          <span class="icon icon-image"></span>
          <span>{{ form.image ? 'Change Image' : 'Add Image/GIF' }}</span>
        </label>

        <button type="submit" class="post-btn" :disabled="submitting || !canSubmit">
          {{ submitting ? 'Posting...' : 'Post' }}
        </button>
      </div>

      <p v-if="error" class="error-message">{{ error }}</p>
      </form>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getToken, getUser } from '../stores/auth'
import { useToast } from '@/composables/useToast'
import { uploadImage, createPost } from '@/services/postsService'
import { getFollowers } from '@/services/usersService'

const { success, error: showError } = useToast()

const props = defineProps({
  groupId: {
    type: Number,
    default: null
  }
})

const emit = defineEmits(['posted', 'postCreated'])

const form = ref({
  title: '',
  content: '',
  privacy: 'public',
  image: null,
  selectedViewers: []
})

const imagePreview = ref(null)
const fileInput = ref(null)
const followers = ref([])
const loadingFollowers = ref(false)
const submitting = ref(false)
const error = ref('')
const isOpen = ref(false)

const canSubmit = computed(() => {
  if (!form.value.content.trim()) return false
  // Only check privacy/viewers if not in a group
  if (!props.groupId && form.value.privacy === 'private' && form.value.selectedViewers.length === 0) {
    return false
  }
  return true
})

onMounted(() => {
  // Only load followers if not posting to a group
  if (!props.groupId) {
    loadFollowers()
  }
})

function toggleOpen() {
  isOpen.value = !isOpen.value
}

async function loadFollowers() {
  const token = getToken()
  if (!token) return

  const user = getUser()
  if (!user?.id) return

  loadingFollowers.value = true
  try {
    const data = await getFollowers(token, user.id)
    // API returns { followers: [...], count: N }
    followers.value = data.followers || []
    console.log('Loaded followers:', followers.value)
  } catch (err) {
    console.warn('Failed to load followers:', err.message)
    followers.value = [] // Set empty array so private posts still work
  } finally {
    loadingFollowers.value = false
  }
}

function handlePrivacyChange() {
  if (form.value.privacy !== 'private') {
    form.value.selectedViewers = []
  }
}

function handleImageSelect(event) {
  const file = event.target.files[0]
  if (!file) return

  // Validate file type
  const validTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif']
  if (!validTypes.includes(file.type)) {
    showError('Please select a valid image (JPG, PNG, or GIF)')
    error.value = 'Please select a valid image (JPG, PNG, or GIF)'
    return
  }

  // Validate file size (5MB max)
  if (file.size > 5 * 1024 * 1024) {
    showError('Image must be less than 5MB')
    error.value = 'Image must be less than 5MB'
    return
  }

  form.value.image = file
  error.value = ''

  // Create preview
  const reader = new FileReader()
  reader.onload = (e) => {
    imagePreview.value = e.target.result
  }
  reader.readAsDataURL(file)
}

function removeImage() {
  form.value.image = null
  imagePreview.value = null
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

async function handleSubmit() {
  if (!canSubmit.value || submitting.value) return

  const token = getToken()
  if (!token) {
    error.value = 'You must be logged in to post'
    showError('You must be logged in to post')
    return
  }

  submitting.value = true
  error.value = ''

  try {
    let imagePath = null

    // Upload image first if present
    if (form.value.image) {
      const uploadData = await uploadImage(form.value.image, token)
      imagePath = uploadData.image_path
    }

    // Create post with image path
    const postData = {
      title: form.value.title.trim() || null,
      content: form.value.content.trim(),
      image_path: imagePath
    }

    // Add group_id if posting to a group
    if (props.groupId) {
      postData.group_id = props.groupId
      // Group posts are always public (visible to group members only)
      postData.privacy_level = 'public'
    } else {
      // Only include privacy settings for non-group posts
      postData.privacy_level = form.value.privacy
      
      if (form.value.privacy === 'private' && form.value.selectedViewers.length > 0) {
        postData.viewers = form.value.selectedViewers
      }
    }

    await createPost(postData, token)

    // Reset form
    form.value = {
      title: '',
      content: '',
      privacy: 'public',
      image: null,
      selectedViewers: []
    }
    imagePreview.value = null
    if (fileInput.value) {
      fileInput.value.value = ''
    }

    success('Post created successfully!')
    emit('posted')
    emit('postCreated')
  } catch (err) {
    console.error('Failed to create post:', err)
    const errorMessage = err.message || 'Failed to create post. Please try again.'
    showError(errorMessage)
    error.value = errorMessage
  } finally {
    submitting.value = false
  }
}

function getDisplayName(user) {
  if (user.nickname) return user.nickname
  if (user.first_name && user.last_name) {
    return `${user.first_name} ${user.last_name}`
  }
  if (user.first_name) return user.first_name
  return user.username
}

function getInitials(user) {
  if (user.first_name && user.last_name) {
    return `${user.first_name[0]}${user.last_name[0]}`.toUpperCase()
  }
  if (user.first_name) return user.first_name[0].toUpperCase()
  return user.username.substring(0, 2).toUpperCase()
}

function getAvatarColor(username) {
  const colors = ['#3498db', '#e74c3c', '#2ecc71', '#f39c12', '#9b59b6', '#1abc9c', '#e67e22']
  let hash = 0
  for (let i = 0; i < username.length; i++) {
    hash = username.charCodeAt(i) + ((hash << 5) - hash)
  }
  return colors[Math.abs(hash) % colors.length]
}
</script>

<style scoped>
.create-post-card {
  background: rgba(8, 10, 24, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 1.25rem;
  padding: 1.5rem;
  margin-bottom: 1.5rem;
  box-shadow: 0 12px 25px rgba(0, 0, 0, 0.35);
}

.create-post-card.collapsed {
  padding: 1rem 1.25rem;
}

.post-header {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  margin: 0 0 0.5rem;
  background: transparent;
  border: none;
  color: inherit;
  padding: 0;
  cursor: pointer;
  text-align: left;
}

.post-header h3 {
  margin: 0;
  font-size: 1.1rem;
  color: var(--neon-cyan);
}

.chevron {
  transition: transform 0.2s ease;
  font-size: 1rem;
}

.chevron.open {
  transform: rotate(180deg);
}

.post-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.title-input {
  width: 100%;
  padding: 0.75rem 1rem;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 0.75rem;
  background: rgba(0, 0, 0, 0.45);
  color: #f8f9ff;
  font-size: 1rem;
  font-weight: 600;
  resize: none;
  transition: border-color 0.2s, box-shadow 0.2s;
  word-wrap: break-word;
  word-break: break-word;
  overflow-wrap: break-word;
}

.title-input:focus {
  outline: none;
  border-color: var(--border-glow);
  box-shadow: 0 0 12px rgba(0, 247, 255, 0.2);
}

.title-input::placeholder {
  color: var(--text-muted);
  font-weight: 400;
}

textarea {
  width: 100%;
  padding: 1rem;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 0.75rem;
  background: rgba(0, 0, 0, 0.45);
  color: #f8f9ff;
  font-size: 1rem;
  font-family: inherit;
  resize: vertical;
  min-height: 100px;
  word-wrap: break-word;
  word-break: break-word;
  overflow-wrap: break-word;
}

textarea:focus {
  outline: none;
  border-color: var(--border-glow);
  box-shadow: 0 0 12px rgba(0, 247, 255, 0.2);
}

textarea::placeholder {
  color: var(--text-muted);
}

.privacy-selector label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.privacy-selector select {
  flex: 1;
  padding: 0.75rem;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 0.75rem;
  background: rgba(0, 0, 0, 0.45);
  color: #f8f9ff;
  font-size: 0.95rem;
  cursor: pointer;
}

.privacy-selector select:focus {
  outline: none;
  border-color: var(--border-glow);
}

.follower-selector {
  padding: 1rem;
  border: 1px solid rgba(0, 247, 255, 0.3);
  border-radius: 0.75rem;
  background: rgba(0, 247, 255, 0.05);
}

.selector-label {
  margin: 0 0 0.75rem;
  font-size: 0.9rem;
  color: var(--neon-cyan);
  font-weight: 600;
}

.followers-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-height: 200px;
  overflow-y: auto;
}

.follower-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: background 0.2s;
}

.follower-item:hover {
  background: rgba(255, 255, 255, 0.05);
}

.follower-item input[type="checkbox"] {
  cursor: pointer;
}

.follower-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex: 1;
}

.follower-avatar {
  width: 32px;
  height: 32px;
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 0.8rem;
  color: #05060d;
}

.follower-details {
  display: flex;
  flex-direction: column;
}

.follower-details strong {
  font-size: 0.9rem;
}

.follower-details small {
  color: var(--text-muted);
  font-size: 0.8rem;
}

.image-preview {
  position: relative;
  border-radius: 0.75rem;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.12);
}

.image-preview img {
  width: 100%;
  max-height: 300px;
  object-fit: contain;
  background: rgba(0, 0, 0, 0.3);
}

.remove-image {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  width: 2rem;
  height: 2rem;
  border-radius: 50%;
  background: rgba(255, 0, 0, 0.8);
  border: none;
  color: white;
  font-size: 1.5rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
}

.remove-image:hover {
  background: rgba(255, 0, 0, 1);
}

.post-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.upload-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 0.75rem;
  background: transparent;
  color: var(--neon-cyan);
  cursor: pointer;
  transition: all 0.2s;
}

.upload-btn:hover {
  background: rgba(0, 247, 255, 0.1);
  border-color: var(--border-glow);
}

.upload-btn input[type="file"] {
  display: none;
}

.post-btn {
  padding: 0.75rem 2rem;
  border: none;
  border-radius: 0.75rem;
  background: linear-gradient(120deg, var(--neon-cyan), var(--neon-pink));
  color: #05060d;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
}

.post-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 247, 255, 0.4);
}

.post-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}

.error-message {
  color: #ff6b6b;
  font-size: 0.9rem;
  margin: 0;
}

.loading,
.no-followers {
  text-align: center;
  color: var(--text-muted);
  padding: 1rem;
  font-size: 0.9rem;
}

.icon {
  width: 1.2rem;
  height: 1.2rem;
}

.icon-lock {
  background: var(--neon-cyan);
  mask: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="white" d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zM9 6c0-1.66 1.34-3 3-3s3 1.34 3 3v2H9V6zm9 14H6V10h12v10zm-6-3c1.1 0 2-.9 2-2s-.9-2-2-2-2 .9-2 2 .9 2 2 2z"/></svg>') center/contain no-repeat;
}

.icon-image {
  background: var(--neon-cyan);
  mask: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="white" d="M21 19V5c0-1.1-.9-2-2-2H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2zM8.5 13.5l2.5 3.01L14.5 12l4.5 6H5l3.5-4.5z"/></svg>') center/contain no-repeat;
}

.followers-list::-webkit-scrollbar {
  width: 6px;
}

.followers-list::-webkit-scrollbar-track {
  background: transparent;
}

.followers-list::-webkit-scrollbar-thumb {
  background: rgba(0, 247, 255, 0.3);
  border-radius: 3px;
}

.slide-fade-enter-active,
.slide-fade-leave-active {
  transition: all 0.25s ease;
}

.slide-fade-enter-from,
.slide-fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
</style>
