/**
 * Throttle function - Limits how often a function can execute
 * Ensures function runs at most once per specified time interval
 * @param {Function} func - Function to throttle
 * @param {number} delay - Minimum time between function calls in milliseconds
 * @returns {Function} Throttled function
 */
export function throttle(func, delay) {
  let lastCall = 0
  return function(...args) {
    const now = Date.now()
    if (now - lastCall >= delay) {
      lastCall = now
      return func.apply(this, args)
    }
  }
}

/**
 * Debounce function - Delays function execution until after user stops triggering events
 * Waits for a "quiet period" before executing
 * @param {Function} func - Function to debounce
 * @param {number} delay - Time to wait after last call in milliseconds
 * @returns {Function} Debounced function
 */
export function debounce(func, delay) {
  let timeoutId
  return function(...args) {
    clearTimeout(timeoutId)
    timeoutId = setTimeout(() => func.apply(this, args), delay)
  }
}
