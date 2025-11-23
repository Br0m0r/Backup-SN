import axios from 'axios'

const POSTS_BASE_URL = import.meta.env.VITE_POSTS_API_URL || 'http://localhost:8083'

const client = axios.create({
  baseURL: POSTS_BASE_URL,
  headers: {
    'Content-Type': 'application/json'
  },
  withCredentials: false,
  timeout: 10000
})

function unwrapResponse(response) {
  const payload = response?.data

  if (!payload) {
    throw new Error('No response from posts service')
  }

  if (payload.success === false) {
    throw new Error(payload.error || 'Posts service error')
  }

  if (payload.data !== undefined) {
    return payload.data
  }

  return payload
}

export async function uploadImage(imageFile, token) {
  const formData = new FormData()
  formData.append('image', imageFile)

  const response = await client.post('/upload/image', formData, {
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'multipart/form-data'
    }
  })

  return unwrapResponse(response)
}

export async function createPost(postData, token) {
  const response = await client.post('/posts', postData, {
    headers: {
      Authorization: `Bearer ${token}`
    }
  })

  return unwrapResponse(response)
}

export async function getPost(postId, token) {
  const response = await client.get(`/posts/${postId}`, {
    headers: {
      Authorization: `Bearer ${token}`
    }
  })

  return unwrapResponse(response)
}

export async function updatePost(postId, postData, token) {
  const response = await client.put(`/posts/${postId}`, postData, {
    headers: {
      Authorization: `Bearer ${token}`
    }
  })

  return unwrapResponse(response)
}

export async function deletePost(postId, token) {
  const response = await client.delete(`/posts/${postId}`, {
    headers: {
      Authorization: `Bearer ${token}`
    }
  })

  return unwrapResponse(response)
}

export async function getComments(postId, token) {
  const response = await client.get('/comments', {
    params: { post_id: postId },
    headers: {
      Authorization: `Bearer ${token}`
    }
  })

  return unwrapResponse(response)
}

export async function createComment(commentData, token) {
  const response = await client.post('/comments', commentData, {
    headers: {
      Authorization: `Bearer ${token}`
    }
  })

  return unwrapResponse(response)
}

export async function updateComment(commentId, commentData, token) {
  const response = await client.put(`/comments/${commentId}`, commentData, {
    headers: {
      Authorization: `Bearer ${token}`
    }
  })

  return unwrapResponse(response)
}

export async function deleteComment(commentId, token) {
  const response = await client.delete(`/comments/${commentId}`, {
    headers: {
      Authorization: `Bearer ${token}`
    }
  })

  return unwrapResponse(response)
}

export async function getFeedPosts(token) {
  const response = await client.get('/posts/feed', {
    headers: {
      Authorization: `Bearer ${token}`
    }
  })

  return unwrapResponse(response)
}

export async function searchPosts(searchTerm, token) {
  const response = await client.get('/posts/search', {
    params: { q: searchTerm },
    headers: {
      Authorization: `Bearer ${token}`
    }
  })

  return unwrapResponse(response)
}
