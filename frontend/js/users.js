// User Management Functions

const Users = {
    async getProfile() {
        const result = await window.Utils.apiCall(`${window.AppConfig.USER_URL}/profile`, 'GET', null, true);
        
        if (result.ok && result.data.success) {
            window.Utils.showResponse(result.data.data, 'profileResponse');
            window.Utils.showStatus('Profile loaded', 'success');
        } else {
            window.Utils.showStatus(result.data?.error || 'Failed to load profile', 'error');
        }
    },

    showUpdateProfile() {
        document.getElementById('updateProfileForm').classList.remove('hidden');
    },

    hideUpdateProfile() {
        document.getElementById('updateProfileForm').classList.add('hidden');
    },

    async updateProfile() {
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

        const result = await window.Utils.apiCall(`${window.AppConfig.USER_URL}/profile`, 'PUT', payload, true);

        if (result.ok && result.data.success) {
            window.Utils.showResponse(result.data.data, 'profileResponse');
            window.Utils.showStatus('Profile updated!', 'success');
            Users.hideUpdateProfile();
        } else {
            window.Utils.showStatus(result.data?.error || 'Failed to update profile', 'error');
            console.error('Update profile error:', result);
        }
    },

    async searchUsers() {
        const query = document.getElementById('searchQuery').value;
        
        if (!query) {
            window.Utils.showStatus('Please enter a search query', 'error');
            return;
        }

        const result = await window.Utils.apiCall(`${window.AppConfig.USER_URL}/search?q=${encodeURIComponent(query)}`, 'GET', null, true);
        
        if (result.ok && result.data.success) {
            Users.displayUsers(result.data.data.users || []);
            window.Utils.showResponse(result.data.data, 'usersResponse');
        } else {
            window.Utils.showStatus(result.data?.error || 'Search failed', 'error');
        }
    },

    async followUser(userId) {
        const result = await window.Utils.apiCall(`${window.AppConfig.USER_URL}/follow`, 'POST', { user_id: userId }, true);
        
        if (result.ok && result.data.success) {
            window.Utils.showStatus('Follow request sent!', 'success');
        } else {
            window.Utils.showStatus(result.data?.error || 'Follow failed', 'error');
        }
    },

    async getFollowers() {
        const result = await window.Utils.apiCall(`${window.AppConfig.USER_URL}/followers`, 'GET', null, true);
        
        if (result.ok && result.data.success) {
            Users.displayUsers(result.data.data.followers || []);
            window.Utils.showResponse(result.data.data, 'usersResponse');
        } else {
            window.Utils.showStatus(result.data?.error || 'Failed to load followers', 'error');
        }
    },

    async getFollowing() {
        const result = await window.Utils.apiCall(`${window.AppConfig.USER_URL}/following`, 'GET', null, true);
        
        if (result.ok && result.data.success) {
            Users.displayUsers(result.data.data.following || []);
            window.Utils.showResponse(result.data.data, 'usersResponse');
        } else {
            window.Utils.showStatus(result.data?.error || 'Failed to load following', 'error');
        }
    },

    displayUsers(users) {
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
};

// Export to global scope
window.Users = Users;

// Make functions available globally for onclick handlers
window.getProfile = Users.getProfile;
window.showUpdateProfile = Users.showUpdateProfile;
window.hideUpdateProfile = Users.hideUpdateProfile;
window.updateProfile = Users.updateProfile;
window.searchUsers = Users.searchUsers;
window.followUser = Users.followUser;
window.getFollowers = Users.getFollowers;
window.getFollowing = Users.getFollowing;
