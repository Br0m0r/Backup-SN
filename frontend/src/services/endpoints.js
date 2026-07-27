export const apiEndpoints = Object.freeze({
  auth: import.meta.env.VITE_AUTH_API_URL || '/api/auth',
  users: import.meta.env.VITE_USERS_API_URL || '/api/users',
  posts: import.meta.env.VITE_POSTS_API_URL || '/api/posts',
  groups: import.meta.env.VITE_GROUPS_API_URL || '/api/groups',
  chat: import.meta.env.VITE_CHAT_API_URL || '/api/chat',
  notifications: import.meta.env.VITE_NOTIFICATIONS_API_URL || '/api/notifications'
})

export function websocketEndpoint(configuredURL, gatewayPath) {
  if (configuredURL) return configuredURL

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${protocol}//${window.location.host}${gatewayPath}`
}
