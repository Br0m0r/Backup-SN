// Configuration and Global State
const API_BASE = 'http://localhost';
const AUTH_URL = `${API_BASE}:8081`;
const USER_URL = `${API_BASE}:8082`;
const POST_URL = `${API_BASE}:8083`;
const GROUP_URL = `${API_BASE}:8084`;

// Global state
let token = localStorage.getItem('token') || '';
let currentUser = null;
let currentGroupId = null;

// Export for use in other modules
window.AppConfig = {
    API_BASE,
    AUTH_URL,
    USER_URL,
    POST_URL,
    GROUP_URL
};

window.AppState = {
    getToken: () => token,
    setToken: (newToken) => {
        token = newToken;
        if (newToken) {
            localStorage.setItem('token', newToken);
        } else {
            localStorage.removeItem('token');
        }
    },
    getCurrentUser: () => currentUser,
    setCurrentUser: (user) => { currentUser = user; },
    getCurrentGroupId: () => currentGroupId,
    setCurrentGroupId: (id) => { currentGroupId = id; }
};
