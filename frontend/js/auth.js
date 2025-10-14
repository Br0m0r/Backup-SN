// Authentication Functions

const Auth = {
    async register() {
        const username = document.getElementById('regUsername').value;
        const email = document.getElementById('regEmail').value;
        const password = document.getElementById('regPassword').value;
        const firstName = document.getElementById('regFirstName').value;
        const lastName = document.getElementById('regLastName').value;

        if (!username || !email || !password || !firstName || !lastName) {
            window.Utils.showStatus('Please fill all fields', 'error', 'authStatus');
            return;
        }

        const result = await window.Utils.apiCall(`${window.AppConfig.AUTH_URL}/register`, 'POST', {
            username, email, password, first_name: firstName, last_name: lastName
        });

        if (result.ok && result.data.success) {
            window.AppState.setToken(result.data.data.token);
            window.Utils.showStatus('Registration successful!', 'success', 'authStatus');
            setTimeout(() => {
                document.getElementById('authSection').classList.add('hidden');
                document.getElementById('mainSection').classList.remove('hidden');
                Auth.getSession();
            }, 1000);
        } else {
            window.Utils.showStatus(result.data?.error || 'Registration failed', 'error', 'authStatus');
        }
    },

    async login() {
        const email = document.getElementById('loginEmail').value;
        const password = document.getElementById('loginPassword').value;

        if (!email || !password) {
            window.Utils.showStatus('Please fill all fields', 'error', 'authStatus');
            return;
        }

        const result = await window.Utils.apiCall(`${window.AppConfig.AUTH_URL}/login`, 'POST', {
            email, password
        });

        if (result.ok && result.data.success) {
            window.AppState.setToken(result.data.data.token);
            window.Utils.showStatus('Login successful!', 'success', 'authStatus');
            setTimeout(() => {
                document.getElementById('authSection').classList.add('hidden');
                document.getElementById('mainSection').classList.remove('hidden');
                Auth.getSession();
            }, 1000);
        } else {
            window.Utils.showStatus(result.data?.error || 'Login failed', 'error', 'authStatus');
        }
    },

    async logout() {
        const result = await window.Utils.apiCall(`${window.AppConfig.AUTH_URL}/logout`, 'POST', null, true);
        
        window.AppState.setToken('');
        window.AppState.setCurrentUser(null);
        window.AppState.setCurrentGroupId(null);
        
        // Clear all UI displays
        document.getElementById('profileResponse').innerHTML = '';
        document.getElementById('postsResponse').innerHTML = '';
        document.getElementById('usersResponse').innerHTML = '';
        document.getElementById('groupsResponse').innerHTML = '';
        document.getElementById('feedContainer').innerHTML = '';
        document.getElementById('usersContainer').innerHTML = '';
        document.getElementById('groupsListContainer').innerHTML = '';
        document.getElementById('groupDetailView').classList.add('hidden');
        
        document.getElementById('authSection').classList.remove('hidden');
        document.getElementById('mainSection').classList.add('hidden');
        window.Utils.showStatus('Logged out successfully', 'info', 'authStatus');
    },

    async getSession() {
        const result = await window.Utils.apiCall(`${window.AppConfig.AUTH_URL}/session`, 'GET', null, true);
        
        if (result.ok && result.data.success) {
            const user = result.data.data.user;
            window.AppState.setCurrentUser(user);
            document.getElementById('userInfo').innerHTML = `
                <p><strong>Logged in as:</strong> ${user.username}</p>
                <p><strong>ID:</strong> ${user.id}</p>
                <p><strong>Email:</strong> ${user.email}</p>
            `;
            document.getElementById('authSection').classList.add('hidden');
            document.getElementById('mainSection').classList.remove('hidden');
        } else {
            Auth.logout();
        }
    }
};

// Export to global scope
window.Auth = Auth;

// Make functions available globally for onclick handlers
window.register = Auth.register;
window.login = Auth.login;
window.logout = Auth.logout;
window.getSession = Auth.getSession;
