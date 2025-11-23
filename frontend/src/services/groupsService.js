import axios from 'axios'

const GROUPS_API_URL = import.meta.env.VITE_GROUPS_API_URL || 'http://localhost:8084'

// Helper to unwrap response data
function unwrapResponse(response) {
  if (response.data.success) {
    return response.data.data
  }
  throw new Error(response.data.error || 'Request failed')
}

// Respond to group invite
export async function respondToGroupInvite(groupId, accept, token) {
  try {
    const response = await axios.post(
      `${GROUPS_API_URL}/groups/${groupId}/respond`,
      { accept },
      {
        headers: {
          Authorization: `Bearer ${token}`
        }
      }
    )
    return unwrapResponse(response)
  } catch (error) {
    throw new Error(error.response?.data?.error || error.message || 'Failed to respond to group invite')
  }
}
