/**
 * Composable for handling avatar URLs
 */
export function useAvatar() {
  const USERS_API_URL = import.meta.env.VITE_USERS_API_URL || 'http://localhost:8082'

  /**
   * Get the full avatar URL from an avatar_path
   * @param {string} avatarPath - The avatar path from the API
   * @param {string} fallbackText - Text to show in placeholder (default: '?')
   * @returns {string} Full avatar URL or placeholder
   */
  function getAvatarUrl(avatarPath, fallbackText = '?') {
    if (!avatarPath) {
      return `https://placehold.co/140x140/161832/fff?text=${encodeURIComponent(fallbackText)}`
    }
    
    // If it's already a full URL, return it
    if (avatarPath.startsWith('http')) {
      return avatarPath
    }
    
    // Otherwise, construct the full URL to the users service
    return `${USERS_API_URL}${avatarPath}`
  }

  /**
   * Get avatar URL for a user object
   * @param {Object} user - User object with avatar_path
   * @param {number} size - Size for placeholder (default: 140)
   * @returns {string} Full avatar URL or placeholder with initials
   */
  function getUserAvatarUrl(user, size = 140) {
    if (!user) {
      return `https://placehold.co/${size}x${size}/161832/fff?text=?`
    }

    if (user.avatar_path) {
      return getAvatarUrl(user.avatar_path)
    }

    // Generate initials for placeholder
    let initials = '?'
    if (user.first_name && user.last_name) {
      initials = (user.first_name[0] + user.last_name[0]).toUpperCase()
    } else if (user.username) {
      initials = user.username.substring(0, 2).toUpperCase()
    } else if (user.nickname) {
      initials = user.nickname.substring(0, 2).toUpperCase()
    }

    return `https://placehold.co/${size}x${size}/161832/fff?text=${encodeURIComponent(initials)}`
  }

  return {
    getAvatarUrl,
    getUserAvatarUrl
  }
}
