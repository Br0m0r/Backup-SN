import axios from 'axios'
import { apiEndpoints } from './endpoints'

const CHAT_API_URL = apiEndpoints.chat

// Helper to unwrap response data
function unwrapResponse(response) {
  if (response.data.success) {
    return response.data.data
  }
  throw new Error(response.data.error || 'Request failed')
}

// Get contacts list
export async function getContacts(token) {
  try {
    const response = await axios.get(`${CHAT_API_URL}/chat/contacts`, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    })
    return unwrapResponse(response)
  } catch (error) {
    throw new Error(error.response?.data?.error || error.message || 'Failed to load contacts')
  }
}

// Get chat history with a user
export async function getChatHistory(userId, token, limit = 50) {
  try {
    const response = await axios.get(`${CHAT_API_URL}/chat/history/${userId}`, {
      params: { limit },
      headers: {
        Authorization: `Bearer ${token}`
      }
    })
    return unwrapResponse(response)
  } catch (error) {
    throw new Error(error.response?.data?.error || error.message || 'Failed to load chat history')
  }
}

// Get group-chat history from Chat, the sole message owner.
export async function getGroupMessages(groupId, token, limit = 50) {
  try {
    const response = await axios.get(`${CHAT_API_URL}/chat/groups/${groupId}/history`, {
      params: { limit },
      headers: {
        Authorization: `Bearer ${token}`
      }
    })
    const data = unwrapResponse(response)
    return Array.isArray(data.messages) ? data.messages : []
  } catch (error) {
    throw new Error(error.response?.data?.error || error.message || 'Failed to load group messages')
  }
}

// Send a group-chat message through Chat when WebSocket delivery is unavailable.
export async function createGroupMessage(groupId, content, token) {
  try {
    const response = await axios.post(
      `${CHAT_API_URL}/chat/groups/${groupId}/messages`,
      { content },
      {
        headers: {
          Authorization: `Bearer ${token}`
        }
      }
    )
    return unwrapResponse(response)
  } catch (error) {
    throw new Error(error.response?.data?.error || error.message || 'Failed to send group message')
  }
}

// Upload image
export async function uploadImage(file, token) {
  try {
    const formData = new FormData()
    formData.append('image', file)

    const response = await axios.post(`${CHAT_API_URL}/upload/image`, formData, {
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'multipart/form-data'
      }
    })
    return unwrapResponse(response)
  } catch (error) {
    throw new Error(error.response?.data?.error || error.message || 'Failed to upload image')
  }
}

// Mark messages as read
export async function markAsRead(userId, token) {
  try {
    const response = await axios.post(`${CHAT_API_URL}/chat/read/${userId}`, {}, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    })
    return unwrapResponse(response)
  } catch (error) {
    throw new Error(error.response?.data?.error || error.message || 'Failed to mark as read')
  }
}

// Get image URL helper
export function getImageUrl(path) {
  if (!path) return ''
  if (path.startsWith('http')) return path
  if (path.startsWith('/media/')) return path
  // Add leading slash if path doesn't have one
  const fullPath = path.startsWith('/') ? path : `/${path}`
  return `${CHAT_API_URL}${fullPath}`
}
