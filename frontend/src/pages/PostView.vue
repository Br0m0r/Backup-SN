<template>
  <section class="post-view">
    <!-- Custom Delete Confirmation Modal -->
    <div v-if="showDeleteModal" class="modal-overlay" @click.self="closeDeleteModal">
      <div class="modal-content">
        <div class="modal-header">
          <h3>⚠️ Confirm Deletion</h3>
        </div>
        <div class="modal-body">
          <p>{{ deleteModalMessage }}</p>
        </div>
        <div class="modal-actions">
          <button @click="confirmDelete" class="confirm-delete-btn">Delete</button>
          <button @click="closeDeleteModal" class="cancel-delete-btn">Cancel</button>
        </div>
      </div>
    </div>

    <div v-if="loading" class="loading">Loading post...</div>
    <div v-else-if="errorMessage" class="error-state">
      <p>{{ errorMessage }}</p>
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
          <!-- Post Actions (Edit/Delete) -->
          <div v-if="isPostOwner" class="post-actions">
            <button @click="editingPost = true" class="action-btn edit-btn" title="Edit post">✏️</button>
            <button @click="deletePost" class="action-btn delete-btn" title="Delete post">🗑️</button>
          </div>
        </div>
        
        <!-- Edit Post Form -->
        <div v-if="editingPost" class="edit-post-form">
          <textarea v-model="editPostForm.content" placeholder="Edit your post..." rows="5"></textarea>
          <div class="form-actions">
            <button @click="savePostEdit" class="submit-btn" :disabled="!editPostForm.content.trim()">Save</button>
            <button @click="cancelPostEdit" class="cancel-btn">Cancel</button>
          </div>
        </div>
        
        <!-- View Post Content -->
        <div v-else>
          <h2 v-if="post.title" class="post-title">{{ post.title }}</h2>
          <p class="post-content">{{ post.content }}</p>
          <img v-if="post.image_path" :src="getImageUrl(post.image_path)" class="post-image" alt="Post image" />
        </div>
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
              <!-- Edit Comment Form -->
              <div v-if="editingComment === comment.id" class="edit-comment-form">
                <textarea v-model="editCommentForm.content" placeholder="Edit your comment..." rows="3"></textarea>
                <div class="form-actions">
                  <button @click="saveCommentEdit(comment.id)" class="submit-btn-sm" :disabled="!editCommentForm.content.trim()">Save</button>
                  <button @click="cancelCommentEdit" class="cancel-btn-sm">Cancel</button>
                </div>
              </div>
              
              <!-- View Comment Content -->
              <div v-else>
                <div class="comment-header">
                  <strong>{{ comment.author.first_name }} {{ comment.author.last_name }}</strong>
                  <small>{{ formatTime(comment.created_at) }}</small>
                  <!-- Comment Actions (Edit/Delete) -->
                  <div v-if="isCommentOwner(comment)" class="comment-actions">
                    <button @click="startEditComment(comment)" class="action-btn-sm edit-btn" title="Edit comment">✏️</button>
                    <button @click="deleteComment(comment.id)" class="action-btn-sm delete-btn" title="Delete comment">🗑️</button>
                  </div>
                </div>
                <p>{{ comment.content }}</p>
                <img v-if="comment.image_path" :src="getImageUrl(comment.image_path)" class="comment-image" alt="Comment image" />
              </div>
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
import { getToken, getUser } from '@/stores/auth'
import { useToast } from '@/composables/useToast'
import {
  getPost,
  getComments,
  uploadImage,
  createComment,
  updatePost,
  deletePost as deletePostService,
  updateComment as updateCommentService,
  deleteComment as deleteCommentService
} from '@/services/postsService'

const route = useRoute()
const router = useRouter()
const { success, error: showError } = useToast()

const post = ref(null)
const comments = ref([])
const loading = ref(true)
const loadingComments = ref(false)
const submittingComment = ref(false)
const errorMessage = ref(null)

const commentForm = ref({
  content: '',
  image: null,
  imagePreview: null
})

const editingPost = ref(false)
const editPostForm = ref({
  content: ''
})

const editingComment = ref(null)
const editCommentForm = ref({
  content: ''
})

const showDeleteModal = ref(false)
const deleteModalMessage = ref('')
const deleteAction = ref(null)

const currentUserInitials = computed(() => {
  const user = getUser()
  if (!user) return '?'
  if (user.first_name && user.last_name) {
    return `${user.first_name[0]}${user.last_name[0]}`.toUpperCase()
  }
  return user.username?.[0]?.toUpperCase() || '?'
})

const isPostOwner = computed(() => {
  const user = getUser()
  return post.value && user && post.value.user_id === user.id
})

function isCommentOwner(comment) {
  const user = getUser()
  return user && comment.user_id === user.id
}

async function loadPost() {
  loading.value = true
  errorMessage.value = null
  try {
    const token = getToken()
    if (!token) {
      errorMessage.value = 'Please log in to view this post'
      loading.value = false
      return
    }

    const postId = route.params.id
    const data = await getPost(postId, token)
    post.value = data.post
    await loadComments()
  } catch (err) {
    console.error('Failed to load post:', err.message)
    errorMessage.value = err.message || 'Failed to load post'
  } finally {
    loading.value = false
  }
}

async function loadComments() {
  loadingComments.value = true
  try {
    const token = getToken()
    const postId = route.params.id
    
    const data = await getComments(postId, token)
    comments.value = data.comments || []
  } catch (err) {
    console.error('Failed to load comments:', err.message)
  } finally {
    loadingComments.value = false
  }
}

function handleImageSelect(event) {
  const file = event.target.files[0]
  if (!file) return

  // Validate file type
  if (!file.type.startsWith('image/')) {
    showError('Please select an image file')
    return
  }

  // Validate file size (5MB max)
  if (file.size > 5 * 1024 * 1024) {
    showError('Image must be less than 5MB')
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
      const uploadData = await uploadImage(commentForm.value.image, token)
      imagePath = uploadData.image_path
    }

    // Create comment
    const commentData = {
      post_id: parseInt(route.params.id),
      content: commentForm.value.content,
      image_path: imagePath
    }

    await createComment(commentData, token)

    // Reset form
    commentForm.value.content = ''
    removeImage()

    // Reload comments
    await loadComments()
  } catch (err) {
    console.error('Failed to submit comment:', err.message)
    showError(err.message || 'Failed to submit comment')
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
  const POSTS_API_URL = import.meta.env.VITE_POSTS_API_URL || 'http://localhost:8083'
  return `${POSTS_API_URL}${path}`
}

// Post edit/delete functions
function cancelPostEdit() {
  editingPost.value = false
  editPostForm.value.content = ''
}

async function savePostEdit() {
  if (!editPostForm.value.content.trim()) return
  
  try {
    const token = getToken()
    await updatePost(post.value.id, {
      content: editPostForm.value.content,
      image_path: post.value.image_path,
      privacy_level: post.value.privacy_level
    }, token)
    
    post.value.content = editPostForm.value.content
    editingPost.value = false
    success('Post updated successfully')
  } catch (err) {
    console.error('Failed to update post:', err.message)
    showError(err.message || 'Failed to update post')
  }
}

function deletePost() {
  deleteModalMessage.value = 'Are you sure you want to delete this post? This action cannot be undone.'
  deleteAction.value = async () => {
    try {
      const token = getToken()
      await deletePostService(post.value.id, token)
      
      router.push('/feed')
    } catch (err) {
      console.error('Failed to delete post:', err.message)
      showError(err.message || 'Failed to delete post')
    }
  }
  showDeleteModal.value = true
}

// Comment edit/delete functions
function startEditComment(comment) {
  editingComment.value = comment.id
  editCommentForm.value.content = comment.content
}

function cancelCommentEdit() {
  editingComment.value = null
  editCommentForm.value.content = ''
}

async function saveCommentEdit(commentId) {
  if (!editCommentForm.value.content.trim()) return
  
  try {
    const token = getToken()
    await updateCommentService(commentId, {
      content: editCommentForm.value.content
    }, token)
    
    // Update local comment
    const comment = comments.value.find(c => c.id === commentId)
    if (comment) {
      comment.content = editCommentForm.value.content
    }
    
    editingComment.value = null
    editCommentForm.value.content = ''
    success('Comment updated successfully')
  } catch (err) {
    console.error('Failed to update comment:', err.message)
    showError(err.message || 'Failed to update comment')
  }
}

function deleteComment(commentId) {
  deleteModalMessage.value = 'Are you sure you want to delete this comment? This action cannot be undone.'
  deleteAction.value = async () => {
    try {
      const token = getToken()
      await deleteCommentService(commentId, token)
      
      comments.value = comments.value.filter(c => c.id !== commentId)
    } catch (err) {
      console.error('Failed to delete comment:', err.message)
      showError(err.message || 'Failed to delete comment')
    }
  }
  showDeleteModal.value = true
}

function closeDeleteModal() {
  showDeleteModal.value = false
  deleteAction.value = null
  deleteModalMessage.value = ''
}

async function confirmDelete() {
  if (deleteAction.value) {
    await deleteAction.value()
  }
  closeDeleteModal()
}

onMounted(() => {
  loadPost()
  // Initialize post edit form when editing starts
  if (post.value) {
    editPostForm.value.content = post.value.content
  }
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
  margin-bottom: 18px;
  position: relative;
}

.post-actions {
  display: flex;
  gap: 8px;
  margin-left: auto;
}

.action-btn {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  padding: 8px 12px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 16px;
  transition: all 0.2s;
}

.action-btn:hover {
  transform: translateY(-2px);
}

.action-btn.edit-btn:hover {
  background: rgba(0, 240, 255, 0.15);
  border-color: var(--neon-cyan);
}

.action-btn.delete-btn:hover {
  background: rgba(255, 0, 100, 0.15);
  border-color: rgba(255, 0, 100, 0.5);
}

.edit-post-form {
  margin-top: 16px;
}

.edit-post-form textarea {
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

.edit-post-form textarea:focus {
  outline: none;
  border-color: var(--neon-cyan);
  background: rgba(255, 255, 255, 0.05);
  box-shadow: 0 0 0 3px rgba(0, 240, 255, 0.1);
}

.cancel-btn {
  padding: 10px 24px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  cursor: pointer;
  font-size: 15px;
  color: rgba(255, 255, 255, 0.8);
  transition: all 0.2s;
}

.cancel-btn:hover {
  background: rgba(255, 255, 255, 0.08);
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
  position: relative;
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

.comment-actions {
  display: flex;
  gap: 4px;
  margin-left: auto;
}

.action-btn-sm {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  padding: 4px 8px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
}

.action-btn-sm:hover {
  transform: translateY(-1px);
}

.action-btn-sm.edit-btn:hover {
  background: rgba(0, 240, 255, 0.15);
  border-color: var(--neon-cyan);
}

.action-btn-sm.delete-btn:hover {
  background: rgba(255, 0, 100, 0.15);
  border-color: rgba(255, 0, 100, 0.5);
}

.edit-comment-form {
  margin-top: 8px;
}

.edit-comment-form textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 10px;
  font-size: 14px;
  font-family: inherit;
  resize: vertical;
  margin-bottom: 8px;
  background: rgba(255, 255, 255, 0.03);
  color: rgba(255, 255, 255, 0.9);
}

.edit-comment-form textarea:focus {
  outline: none;
  border-color: var(--neon-cyan);
  background: rgba(255, 255, 255, 0.05);
  box-shadow: 0 0 0 2px rgba(0, 240, 255, 0.1);
}

.submit-btn-sm, .cancel-btn-sm {
  padding: 6px 14px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
}

.submit-btn-sm {
  background: linear-gradient(135deg, var(--neon-cyan), var(--neon-pink));
  color: white;
  border: none;
  font-weight: 600;
}

.submit-btn-sm:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 240, 255, 0.3);
}

.submit-btn-sm:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.cancel-btn-sm {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.8);
}

.cancel-btn-sm:hover {
  background: rgba(255, 255, 255, 0.08);
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

/* Custom Delete Modal */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.8);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.modal-content {
  background: linear-gradient(135deg, rgba(8, 10, 24, 0.98), rgba(15, 20, 40, 0.98));
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 16px;
  padding: 0;
  max-width: 450px;
  width: 90%;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5), 0 0 0 1px rgba(255, 255, 255, 0.1);
  animation: slideUp 0.3s ease;
}

@keyframes slideUp {
  from { 
    opacity: 0;
    transform: translateY(20px);
  }
  to { 
    opacity: 1;
    transform: translateY(0);
  }
}

.modal-header {
  padding: 24px 24px 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.modal-header h3 {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
  color: rgba(255, 255, 255, 0.95);
  display: flex;
  align-items: center;
  gap: 10px;
}

.modal-body {
  padding: 24px;
}

.modal-body p {
  margin: 0;
  font-size: 15px;
  line-height: 1.6;
  color: rgba(255, 255, 255, 0.8);
}

.modal-actions {
  padding: 16px 24px 24px;
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.confirm-delete-btn,
.cancel-delete-btn {
  padding: 10px 24px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
}

.confirm-delete-btn {
  background: linear-gradient(135deg, #dc3545, #c82333);
  color: white;
  box-shadow: 0 4px 12px rgba(220, 53, 69, 0.3);
}

.confirm-delete-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(220, 53, 69, 0.4);
}

.cancel-delete-btn {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.9);
}

.cancel-delete-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.2);
}
</style>
