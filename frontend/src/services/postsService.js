import axios from 'axios'

const POSTS_BASE_URL = import.meta.env.VITE_POSTS_API_URL 

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
  try {
    const response = await client.post('/posts', postData, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    })

    return unwrapResponse(response)
  } catch (error) {
    // Extract backend error message if available
    if (error.response?.data?.error) {
      throw new Error(error.response.data.error)
    }
    throw error
  }
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
  try {
    const response = await client.put(`/posts/${postId}`, postData, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    })

    return unwrapResponse(response)
  } catch (error) {
    if (error.response?.data?.error) {
      throw new Error(error.response.data.error)
    }
    throw error
  }
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
  try {
    const response = await client.post('/comments', commentData, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    })

    return unwrapResponse(response)
  } catch (error) {
    if (error.response?.data?.error) {
      throw new Error(error.response.data.error)
    }
    throw error
  }
}

export async function updateComment(commentId, commentData, token) {
  try {
    const response = await client.put(`/comments/${commentId}`, commentData, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    })

    return unwrapResponse(response)
  } catch (error) {
    if (error.response?.data?.error) {
      throw new Error(error.response.data.error)
    }
    throw error
  }
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

export async function getUserPosts(userId, token) {
  // Get feed posts and filter by user ID on the frontend
  const response = await client.get('/posts/feed', {
    headers: {
      Authorization: `Bearer ${token}`
    }
  })

  const data = unwrapResponse(response)
  
  // Filter posts by the target user ID
  if (data.posts && Array.isArray(data.posts)) {
    return {
      posts: data.posts.filter(post => post.user_id === userId)
    }
  }
  
  return { posts: [] }
}

export async function getGroupPosts(groupId, token) {
  const response = await client.get(`/posts/group/${groupId}`, {
    headers: {
      Authorization: `Bearer ${token}`
    }
  })

  return unwrapResponse(response)
}

/**
 * Get the full URL for a post image
 * @param {string} imagePath - The image path from the API
 * @returns {string} Full image URL or empty string
 */
export function getPostImageUrl(imagePath) {
  if (!imagePath) return ''
  
  // If it's already a full URL, return it
  if (imagePath.startsWith('http')) return imagePath
  
  // Add leading slash if path doesn't have one
  const fullPath = imagePath.startsWith('/') ? imagePath : `/${imagePath}`
  return `${POSTS_BASE_URL}${fullPath}`
}
