<template>
  <section class="post-view">
    <div v-if="loading" class="loading">Loading post...</div>
    <div v-else-if="error" class="error-state">
      <p>{{ error }}</p>
      <button @click="$router.push('/feed')">Back to Feed</button>
    </div>
    <div v-else-if="post" class="post-container">
      <!-- Post Header -->
      <header class="post-header">
        <button class="back-button" @click="$router.push('/feed')">
          <span>←</span> Back to Feed
        </button>
      </header>

      <!-- Main Post -->
      <article class="main-post">
        <div class="post-author">
          <div class="avatar">{{ getInitials(post.author) }}</div>
          <div class="author-info">
            <strong>{{ post.author.first_name }} {{ post.author.last_name }}</strong>
            <small>{{ formatTime(post.created_at) }} · {{ formatPrivacy(post.privacy_level) }}</small>
          </div>
        </div>
        <h2 v-if="post.title" class="post-title">{{ post.title }}</h2>
        <p class="post-content">{{ post.content }}</p>
        <img v-if="post.image_path" :src="getImageUrl(post.image_path)" class="post-image" alt="Post image" />
      </article>

      <!-- Comments Section -->
      <div class="comments-section">
        <h3>Comments ({{ comments.length }})</h3>
        
        <!-- Comments List -->
        <div v-if="loadingComments" class="loading">Loading comments...</div>
        <div v-else-if="comments.length === 0" class="empty-comments">
          <p>No comments yet. Be the first to comment!</p>
        </div>
        <div v-else class="comments-list">
          <article v-for="comment in comments" :key="comment.id" class="comment">
            <div class="avatar">{{ getInitials(comment.author) }}</div>
            <div class="comment-content">
              <div class="comment-header">
                <strong>{{ comment.author.first_name }} {{ comment.author.last_name }}</strong>
                <small>{{ formatTime(comment.created_at) }}</small>
              </div>
              <p>{{ comment.content }}</p>
              <img v-if="comment.image_path" :src="getImageUrl(comment.image_path)" class="comment-image" alt="Comment image" />
            </div>
          </article>
        </div>

        <!-- Comment Form -->
        <div class="comment-form">
          <div class="avatar">{{ currentUserInitials }}</div>
          <div class="form-content">
            <textarea 
              v-model="commentForm.content" 
              placeholder="Write a comment..."
              rows="3"
              @keydown.meta.enter="submitComment"
              @keydown.ctrl.enter="submitComment"
            ></textarea>
            <div class="form-actions">
              <label v-if="!commentForm.image" class="image-upload-btn">
                <input type="file" @change="handleImageSelect" accept="image/*" hidden />
                <span>📷 Add Image</span>
              </label>
              <div v-else class="image-preview">
                <img :src="commentForm.imagePreview" alt="Preview" />
                <button type="button" @click="removeImage" class="remove-image">✕</button>
              </div>
              <button 
                @click="submitComment" 
                :disabled="!commentForm.content.trim() || submittingComment"
                class="submit-btn"
              >
                {{ submittingComment ? 'Posting...' : 'Post Comment' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'
import { getToken, getUser } from '@/stores/auth'

const route = useRoute()
const router = useRouter()

const post = ref(null)
const comments = ref([])
const loading = ref(true)
const loadingComments = ref(false)
const submittingComment = ref(false)
const error = ref(null)

const commentForm = ref({
  content: '',
  image: null,
  imagePreview: null
})

const POSTS_API_URL = import.meta.env.VITE_POSTS_API_URL || 'http://localhost:8083'

const currentUserInitials = computed(() => {
  const user = getUser()
  if (!user) return '?'
  if (user.first_name && user.last_name) {
    return `${user.first_name[0]}${user.last_name[0]}`.toUpperCase()
  }
  return user.username?.[0]?.toUpperCase() || '?'
})

async function loadPost() {
  loading.value = true
  error.value = null
  try {
    const token = getToken()
    if (!token) {
      error.value = 'Please log in to view this post'
      loading.value = false
      return
    }

    const postId = route.params.id
    const response = await axios.get(`${POSTS_API_URL}/posts/${postId}`, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    })

    post.value = response.data.data.post
    await loadComments()
  } catch (err) {
    console.error('Failed to load post:', err.response?.data || err.message)
    error.value = err.response?.data?.error || 'Failed to load post'
  } finally {
    loading.value = false
  }
}

async function loadComments() {
  loadingComments.value = true
  try {
    const token = getToken()
    const postId = route.params.id
    
    const response = await axios.get(`${POSTS_API_URL}/comments`, {
      params: { post_id: postId },
      headers: {
        Authorization: `Bearer ${token}`
      }
    })

    comments.value = response.data.data.comments || []
  } catch (err) {
    console.error('Failed to load comments:', err.response?.data || err.message)
  } finally {
    loadingComments.value = false
  }
}

function handleImageSelect(event) {
  const file = event.target.files[0]
  if (!file) return

  // Validate file type
  if (!file.type.startsWith('image/')) {
    alert('Please select an image file')
    return
  }

  // Validate file size (5MB max)
  if (file.size > 5 * 1024 * 1024) {
    alert('Image must be less than 5MB')
    return
  }

  commentForm.value.image = file
  commentForm.value.imagePreview = URL.createObjectURL(file)
}

function removeImage() {
  if (commentForm.value.imagePreview) {
    URL.revokeObjectURL(commentForm.value.imagePreview)
  }
  commentForm.value.image = null
  commentForm.value.imagePreview = null
}

async function submitComment() {
  if (!commentForm.value.content.trim() || submittingComment.value) return

  submittingComment.value = true
  try {
    const token = getToken()
    let imagePath = null

    // Upload image if selected
    if (commentForm.value.image) {
      const formData = new FormData()
      formData.append('image', commentForm.value.image)

      const uploadResponse = await axios.post(`${POSTS_API_URL}/upload/image`, formData, {
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'multipart/form-data'
        }
      })

      imagePath = uploadResponse.data.data.image_path
    }

    // Create comment
    const commentData = {
      post_id: parseInt(route.params.id),
      content: commentForm.value.content,
      image_path: imagePath
    }

    await axios.post(`${POSTS_API_URL}/comments`, commentData, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    })

    // Reset form
    commentForm.value.content = ''
    removeImage()

    // Reload comments
    await loadComments()
  } catch (err) {
    console.error('Failed to submit comment:', err.response?.data || err.message)
    alert(err.response?.data?.error || 'Failed to submit comment')
  } finally {
    submittingComment.value = false
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
  const diff = Math.floor((now - date) / 1000)

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

onMounted(() => {
  loadPost()
})
</script>

<style scoped>
.post-view {
  max-width: 900px;
  margin: 0 auto;
  padding: 24px;
  min-height: 100vh;
}

.loading, .error-state {
  text-align: center;
  padding: 60px 20px;
  color: rgba(255, 255, 255, 0.9);
  font-size: 16px;
}

.error-state button {
  margin-top: 16px;
  padding: 10px 28px;
  background: linear-gradient(135deg, var(--neon-cyan), var(--neon-pink));
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  transition: transform 0.2s, box-shadow 0.2s;
}

.error-state button:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(0, 240, 255, 0.3);
}

.post-header {
  margin-bottom: 20px;
}

.back-button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  background: rgba(8, 10, 24, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.9);
  transition: all 0.2s;
  backdrop-filter: blur(12px);
}

.back-button:hover {
  background: rgba(8, 10, 24, 0.95);
  border-color: var(--neon-cyan);
  transform: translateX(-4px);
  box-shadow: 0 4px 12px rgba(0, 240, 255, 0.2);
}

.back-button span {
  font-size: 18px;
  transition: transform 0.2s;
}

.back-button:hover span {
  transform: translateX(-2px);
}

.main-post {
  background: rgba(8, 10, 24, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 16px;
  padding: 32px;
  margin-bottom: 24px;
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(12px);
  transition: border-color 0.3s;
}

.main-post:hover {
  border-color: rgba(0, 240, 255, 0.3);
}

.post-author {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 20px;
}

.avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--neon-cyan), var(--neon-pink));
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: bold;
  font-size: 16px;
  flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(0, 240, 255, 0.3);
  border: 2px solid rgba(255, 255, 255, 0.1);
}

.author-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.author-info strong {
  font-size: 16px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.95);
  letter-spacing: 0.2px;
}

.author-info small {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.6);
}

.post-title {
  font-size: 24px;
  font-weight: 700;
  margin: 0 0 16px 0;
  color: rgba(255, 255, 255, 0.95);
  line-height: 1.3;
  background: linear-gradient(135deg, var(--neon-cyan), var(--neon-pink));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.post-content {
  font-size: 16px;
  line-height: 1.7;
  color: rgba(255, 255, 255, 0.85);
  margin-bottom: 20px;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.post-image {
  width: 100%;
  border-radius: 12px;
  margin-top: 16px;
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.3);
  transition: transform 0.3s;
}

.post-image:hover {
  transform: scale(1.02);
}

.comments-section {
  background: rgba(8, 10, 24, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 16px;
  padding: 28px;
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(12px);
}

.comments-section h3 {
  font-size: 20px;
  font-weight: 700;
  margin: 0 0 24px 0;
  color: rgba(255, 255, 255, 0.95);
  display: flex;
  align-items: center;
  gap: 8px;
}

.comments-section h3::before {
  content: '💬';
  font-size: 22px;
}

.comment-form {
  display: flex;
  gap: 14px;
  margin-top: 28px;
  padding-top: 28px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.form-content {
  flex: 1;
}

.comment-form textarea {
  width: 100%;
  padding: 14px 16px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  font-size: 15px;
  font-family: inherit;
  resize: vertical;
  margin-bottom: 10px;
  background: rgba(255, 255, 255, 0.03);
  color: rgba(255, 255, 255, 0.9);
  transition: all 0.2s;
}

.comment-form textarea::placeholder {
  color: rgba(255, 255, 255, 0.4);
}

.comment-form textarea:focus {
  outline: none;
  border-color: var(--neon-cyan);
  background: rgba(255, 255, 255, 0.05);
  box-shadow: 0 0 0 3px rgba(0, 240, 255, 0.1);
}

.form-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.image-upload-btn {
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.8);
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 4px;
}

.image-upload-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  border-color: var(--neon-cyan);
  transform: translateY(-1px);
}

.image-preview {
  position: relative;
  display: inline-block;
}

.image-preview img {
  max-width: 180px;
  max-height: 120px;
  border-radius: 10px;
  object-fit: cover;
  border: 2px solid rgba(255, 255, 255, 0.1);
}

.remove-image {
  position: absolute;
  top: -10px;
  right: -10px;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: linear-gradient(135deg, #ff4444, #cc0000);
  color: white;
  border: 2px solid rgba(8, 10, 24, 0.9);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: bold;
  transition: all 0.2s;
  box-shadow: 0 4px 12px rgba(255, 68, 68, 0.4);
}

.remove-image:hover {
  transform: scale(1.1) rotate(90deg);
  box-shadow: 0 6px 16px rgba(255, 68, 68, 0.6);
}

.submit-btn {
  padding: 10px 28px;
  background: linear-gradient(135deg, var(--neon-cyan), var(--neon-pink));
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 15px;
  font-weight: 700;
  margin-left: auto;
  transition: all 0.2s;
  box-shadow: 0 4px 12px rgba(0, 240, 255, 0.3);
  letter-spacing: 0.3px;
}

.submit-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(0, 240, 255, 0.4);
}

.submit-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  transform: none;
}

.empty-comments {
  text-align: center;
  padding: 50px 20px;
  color: rgba(255, 255, 255, 0.5);
  font-size: 15px;
}

.comments-list {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.comment {
  display: flex;
  gap: 12px;
  animation: slideIn 0.3s ease-out;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.comment .avatar {
  width: 36px;
  height: 36px;
  font-size: 13px;
}

.comment-content {
  flex: 1;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 14px;
  padding: 14px 16px;
  transition: all 0.2s;
}

.comment-content:hover {
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(0, 240, 255, 0.2);
}

.comment-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
}

.comment-header strong {
  font-size: 14px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
}

.comment-header small {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.5);
}

.comment-content p {
  font-size: 14px;
  line-height: 1.5;
  color: rgba(255, 255, 255, 0.8);
  margin: 0;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.comment-image {
  max-width: 100%;
  border-radius: 10px;
  margin-top: 10px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}
</style>
