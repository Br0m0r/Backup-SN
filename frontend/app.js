// Configuration
const API_BASE = 'http://localhost';
const AUTH_URL = `${API_BASE}:8081`;
const USER_URL = `${API_BASE}:8082`;
const POST_URL = `${API_BASE}:8083`;

let token = localStorage.getItem('token') || '';
let currentUser = null;

// Initialize
if (token) {
    getSession();
}

// Auth Tab Switching
function switchAuthTab(tab) {
    document.querySelectorAll('#authSection .tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('#authSection .tab-content').forEach(c => c.classList.remove('active'));
    
    if (tab === 'register') {
        document.querySelector('#authSection .tab:first-child').classList.add('active');
        document.getElementById('registerForm').classList.add('active');
    } else {
        document.querySelector('#authSection .tab:last-child').classList.add('active');
        document.getElementById('loginForm').classList.add('active');
    }
}

// Main Tab Switching
function switchMainTab(tab) {
    document.querySelectorAll('#mainSection .tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('#mainSection .tab-content').forEach(c => c.classList.remove('active'));
    
    const tabs = ['profile', 'posts', 'users'];
    const index = tabs.indexOf(tab);
    document.querySelectorAll('#mainSection .tab')[index].classList.add('active');
    document.getElementById(tab + 'Tab').classList.add('active');
}

// Utility Functions
function showStatus(message, type, elementId = 'globalStatus') {
    const el = document.getElementById(elementId);
    el.innerHTML = `<div class="status ${type}">${message}</div>`;
    setTimeout(() => el.innerHTML = '', 5000);
}

function showResponse(data, elementId) {
    document.getElementById(elementId).innerHTML = `<pre>${JSON.stringify(data, null, 2)}</pre>`;
}

async function apiCall(url, method = 'GET', body = null, useToken = false) {
    const options = {
        method,
        headers: {
            'Content-Type': 'application/json',
        }
    };

    if (useToken && token) {
        options.headers['Authorization'] = `Bearer ${token}`;
    }

    if (body) {
        options.body = JSON.stringify(body);
    }

    try {
        const response = await fetch(url, options);
        const data = await response.json();
        return { ok: response.ok, status: response.status, data };
    } catch (error) {
        return { ok: false, error: error.message };
    }
}

// AUTH SERVICE
async function register() {
    const username = document.getElementById('regUsername').value;
    const email = document.getElementById('regEmail').value;
    const password = document.getElementById('regPassword').value;
    const firstName = document.getElementById('regFirstName').value;
    const lastName = document.getElementById('regLastName').value;

    if (!username || !email || !password || !firstName || !lastName) {
        showStatus('Please fill all fields', 'error', 'authStatus');
        return;
    }

    const result = await apiCall(`${AUTH_URL}/register`, 'POST', {
        username, email, password, first_name: firstName, last_name: lastName
    });

    if (result.ok && result.data.success) {
        token = result.data.data.token;
        localStorage.setItem('token', token);
        showStatus('Registration successful!', 'success', 'authStatus');
        setTimeout(() => {
            document.getElementById('authSection').classList.add('hidden');
            document.getElementById('mainSection').classList.remove('hidden');
            getSession();
        }, 1000);
    } else {
        showStatus(result.data.error || 'Registration failed', 'error', 'authStatus');
    }
}

async function login() {
    const email = document.getElementById('loginEmail').value;
    const password = document.getElementById('loginPassword').value;

    if (!email || !password) {
        showStatus('Please fill all fields', 'error', 'authStatus');
        return;
    }

    const result = await apiCall(`${AUTH_URL}/login`, 'POST', {
        email, password
    });

    if (result.ok && result.data.success) {
        token = result.data.data.token;
        localStorage.setItem('token', token);
        showStatus('Login successful!', 'success', 'authStatus');
        setTimeout(() => {
            document.getElementById('authSection').classList.add('hidden');
            document.getElementById('mainSection').classList.remove('hidden');
            getSession();
        }, 1000);
    } else {
        showStatus(result.data.error || 'Login failed', 'error', 'authStatus');
    }
}

async function logout() {
    const result = await apiCall(`${AUTH_URL}/logout`, 'POST', null, true);
    
    token = '';
    localStorage.removeItem('token');
    currentUser = null;
    
    document.getElementById('authSection').classList.remove('hidden');
    document.getElementById('mainSection').classList.add('hidden');
    showStatus('Logged out successfully', 'info', 'authStatus');
}

async function getSession() {
    const result = await apiCall(`${AUTH_URL}/session`, 'GET', null, true);
    
    if (result.ok && result.data.success) {
        currentUser = result.data.data.user;
        document.getElementById('userInfo').innerHTML = `
            <p><strong>Logged in as:</strong> ${currentUser.username}</p>
            <p><strong>ID:</strong> ${currentUser.id}</p>
            <p><strong>Email:</strong> ${currentUser.email}</p>
        `;
        document.getElementById('authSection').classList.add('hidden');
        document.getElementById('mainSection').classList.remove('hidden');
    } else {
        logout();
    }
}

// USER SERVICE
async function getProfile() {
    const result = await apiCall(`${USER_URL}/profile`, 'GET', null, true);
    
    if (result.ok && result.data.success) {
        showResponse(result.data.data, 'profileResponse');
        showStatus('Profile loaded', 'success');
    } else {
        showStatus(result.data.error || 'Failed to load profile', 'error');
    }
}

function showUpdateProfile() {
    document.getElementById('updateProfileForm').classList.remove('hidden');
}

function hideUpdateProfile() {
    document.getElementById('updateProfileForm').classList.add('hidden');
}

async function updateProfile() {
    const nickname = document.getElementById('updateNickname').value.trim();
    const aboutMe = document.getElementById('updateAboutMe').value.trim();
    const isPublic = parseInt(document.getElementById('updateIsPublic').value);

    // Build payload with only non-empty values
    const payload = {};
    if (nickname) payload.nickname = nickname;
    if (aboutMe) payload.about_me = aboutMe;
    // Convert number to boolean (1 -> true, 0 -> false)
    payload.is_public_profile = isPublic === 1;

    console.log('Updating profile with payload:', payload);

    const result = await apiCall(`${USER_URL}/profile`, 'PUT', payload, true);

    if (result.ok && result.data.success) {
        showResponse(result.data.data, 'profileResponse');
        showStatus('Profile updated!', 'success');
        hideUpdateProfile();
    } else {
        showStatus(result.data.error || 'Failed to update profile', 'error');
        console.error('Update profile error:', result);
    }
}

async function searchUsers() {
    const query = document.getElementById('searchQuery').value;
    
    if (!query) {
        showStatus('Please enter a search query', 'error');
        return;
    }

    const result = await apiCall(`${USER_URL}/search?q=${encodeURIComponent(query)}`, 'GET', null, true);
    
    if (result.ok && result.data.success) {
        displayUsers(result.data.data.users || []);
        showResponse(result.data.data, 'usersResponse');
    } else {
        showStatus(result.data.error || 'Search failed', 'error');
    }
}

async function followUser(userId) {
    const result = await apiCall(`${USER_URL}/follow`, 'POST', { user_id: userId }, true);
    
    if (result.ok && result.data.success) {
        showStatus('Follow request sent!', 'success');
    } else {
        showStatus(result.data.error || 'Follow failed', 'error');
    }
}

async function getFollowers() {
    const result = await apiCall(`${USER_URL}/followers`, 'GET', null, true);
    
    if (result.ok && result.data.success) {
        displayUsers(result.data.data.followers || []);
        showResponse(result.data.data, 'usersResponse');
    } else {
        showStatus(result.data.error || 'Failed to load followers', 'error');
    }
}

async function getFollowing() {
    const result = await apiCall(`${USER_URL}/following`, 'GET', null, true);
    
    if (result.ok && result.data.success) {
        displayUsers(result.data.data.following || []);
        showResponse(result.data.data, 'usersResponse');
    } else {
        showStatus(result.data.error || 'Failed to load following', 'error');
    }
}

function displayUsers(users) {
    const container = document.getElementById('usersContainer');
    if (!users || users.length === 0) {
        container.innerHTML = '<p style="color: #999; padding: 20px;">No users found</p>';
        return;
    }

    container.innerHTML = users.map(user => `
        <div class="user-card">
            <h4>${user.first_name} ${user.last_name}</h4>
            <p><strong>Username:</strong> ${user.username || 'N/A'}</p>
            <p><strong>Nickname:</strong> ${user.nickname || 'N/A'}</p>
            <p><strong>About:</strong> ${user.about_me || 'No bio'}</p>
            <button onclick="followUser(${user.id})">Follow</button>
        </div>
    `).join('');
}

// POST SERVICE
function showCreatePost() {
    document.getElementById('createPostForm').classList.remove('hidden');
}

function hideCreatePost() {
    document.getElementById('createPostForm').classList.add('hidden');
}

async function createPost() {
    const content = document.getElementById('postContent').value;
    const privacy = document.getElementById('postPrivacy').value;

    if (!content) {
        showStatus('Please enter post content', 'error');
        return;
    }

    const result = await apiCall(`${POST_URL}/posts`, 'POST', {
        content,
        privacy_level: privacy,
        viewers: []
    }, true);

    if (result.ok && result.data.success) {
        showStatus('Post created!', 'success');
        showResponse(result.data.data, 'postsResponse');
        hideCreatePost();
        document.getElementById('postContent').value = '';
        getFeed();
    } else {
        showStatus(result.data.error || 'Failed to create post', 'error');
    }
}

async function getFeed() {
    const result = await apiCall(`${POST_URL}/posts`, 'GET', null, true);
    
    if (result.ok && result.data.success) {
        displayPosts(result.data.data.posts || []);
        showResponse(result.data.data, 'postsResponse');
    } else {
        showStatus(result.data.error || 'Failed to load feed', 'error');
    }
}

async function deletePost(postId) {
    if (!confirm('Delete this post?')) return;

    const result = await apiCall(`${POST_URL}/posts/${postId}`, 'DELETE', null, true);
    
    if (result.ok && result.data.success) {
        showStatus('Post deleted!', 'success');
        getFeed();
    } else {
        showStatus(result.data.error || 'Failed to delete post', 'error');
    }
}

async function getComments(postId) {
    const result = await apiCall(`${POST_URL}/comments?post_id=${postId}`, 'GET', null, true);
    
    if (result.ok && result.data.success) {
        showResponse(result.data.data, 'postsResponse');
        const comments = result.data.data.comments || [];
        showStatus(`Loaded ${comments.length} comments`, 'info');
    } else {
        showStatus(result.data.error || 'Failed to load comments', 'error');
    }
}

function displayPosts(posts) {
    const container = document.getElementById('feedContainer');
    if (!posts || posts.length === 0) {
        container.innerHTML = '<p style="color: #999; padding: 20px;">No posts found</p>';
        return;
    }

    container.innerHTML = posts.map(post => `
        <div class="post-card">
            <h4>Post #${post.id} <small>(${post.privacy_level})</small></h4>
            <p>${post.content}</p>
            <p><small>By User ID: ${post.user_id} • ${new Date(post.created_at).toLocaleString()}</small></p>
            <div class="post-actions">
                <button onclick="getComments(${post.id})">View Comments</button>
                ${post.user_id === currentUser?.id ? `<button onclick="deletePost(${post.id})">Delete</button>` : ''}
            </div>
        </div>
    `).join('');
}
