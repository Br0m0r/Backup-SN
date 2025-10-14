// Main Application Entry Point
// This file loads after all modules and initializes the application

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
    console.log('Social Network App Initialized');
    
    // Check if user has a saved token and try to restore session
    const token = window.AppState.getToken();
    if (token) {
        window.Auth.getSession();
    }
});

// Make tab switching available globally
window.switchAuthTab = window.Utils.switchAuthTab;
window.switchMainTab = window.Utils.switchMainTab;
