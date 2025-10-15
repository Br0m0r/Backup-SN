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
            // Refresh the user list to update button states
            Users.searchUsers();
        } else {
            window.Utils.showStatus(result.data?.error || 'Follow failed', 'error');
        }
    },

    async unfollowUser(userId) {
        const result = await window.Utils.apiCall(`${window.AppConfig.USER_URL}/follow`, 'DELETE', { user_id: userId }, true);
        
        if (result.ok && result.data.success) {
            window.Utils.showStatus('Unfollowed successfully', 'success');
            // Refresh the user list to update button states
            Users.searchUsers();
        } else {
            window.Utils.showStatus(result.data?.error || 'Unfollow failed', 'error');
        }
    },

    async getFollowStatus(userId) {
        const result = await window.Utils.apiCall(`${window.AppConfig.USER_URL}/follow/status/${userId}`, 'GET', null, true);
        
        if (result.ok && result.data.success) {
            return result.data.data.status; // "none", "pending", or "accepted"
        }
        return "none";
    },

    async getFollowers() {
        const result = await window.Utils.apiCall(`${window.AppConfig.USER_URL}/followers`, 'GET', null, true);
        
        if (result.ok && result.data.success) {
            Users.displayUsersInContainer(result.data.data.followers || [], 'followersContainer', false);
            window.Utils.showResponse(result.data.data, 'usersResponse');
        } else {
            window.Utils.showStatus(result.data?.error || 'Failed to load followers', 'error');
        }
    },

    async getFollowing() {
        const result = await window.Utils.apiCall(`${window.AppConfig.USER_URL}/following`, 'GET', null, true);
        
        if (result.ok && result.data.success) {
            Users.displayUsersInContainer(result.data.data.following || [], 'followingContainer', true);
            window.Utils.showResponse(result.data.data, 'usersResponse');
        } else {
            window.Utils.showStatus(result.data?.error || 'Failed to load following', 'error');
        }
    },

    async getFollowRequests() {
        const result = await window.Utils.apiCall(`${window.AppConfig.USER_URL}/follow/requests`, 'GET', null, true);
        
        if (result.ok && result.data.success) {
            Users.displayFollowRequests(result.data.data.requests || []);
            window.Utils.showResponse(result.data.data, 'usersResponse');
            window.Utils.showStatus(`${result.data.data.count} pending requests`, 'info');
        } else {
            window.Utils.showStatus(result.data?.error || 'Failed to load requests', 'error');
        }
    },

    async respondToFollowRequest(followerId, accept) {
        const result = await window.Utils.apiCall(`${window.AppConfig.USER_URL}/follow/respond`, 'POST', 
            { follower_id: followerId, accept: accept }, true);
        
        if (result.ok && result.data.success) {
            window.Utils.showStatus(accept ? 'Request accepted!' : 'Request rejected', 'success');
            Users.getFollowRequests(); // Refresh the list
        } else {
            window.Utils.showStatus(result.data?.error || 'Failed to respond', 'error');
        }
    },

    async displayUsers(users) {
        const container = document.getElementById('searchUsersContainer');
        if (!users || users.length === 0) {
            container.innerHTML = '<p style="color: #999; padding: 20px;">No users found</p>';
            return;
        }

        // Get follow status for all users
        const usersWithStatus = await Promise.all(users.map(async (user) => {
            const status = await Users.getFollowStatus(user.id);
            return { ...user, followStatus: status };
        }));

        container.innerHTML = usersWithStatus.map(user => `
            <div class="user-card">
                <h4>${user.first_name} ${user.last_name}</h4>
                <p><strong>Username:</strong> ${user.username || 'N/A'}</p>
                <p><strong>Nickname:</strong> ${user.nickname || 'N/A'}</p>
                <p><strong>About:</strong> ${user.about_me || 'No bio'}</p>
                <div style="margin-top: 10px;">
                    ${user.followStatus === 'none' ? `<button onclick="followUser(${user.id})">Follow</button>` : ''}
                    ${user.followStatus === 'pending' ? `<button disabled style="opacity: 0.6; cursor: not-allowed; background: #ffc107;">Request Pending</button>` : ''}
                    ${user.followStatus === 'accepted' ? `<button onclick="unfollowUser(${user.id})" style="background: #28a745;">Following</button>` : ''}
                </div>
            </div>
        `).join('');
    },

    displayUsersInContainer(users, containerId, showUnfollow) {
        const container = document.getElementById(containerId);
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
                ${showUnfollow ? `<button onclick="unfollowUser(${user.id})" style="background: #dc3545;">Unfollow</button>` : ''}
            </div>
        `).join('');
    },

    displayFollowRequests(requests) {
        const container = document.getElementById('followRequestsContainer');
        if (!requests || requests.length === 0) {
            container.innerHTML = '<p style="color: #999; padding: 20px;">No pending follow requests</p>';
            return;
        }

        container.innerHTML = requests.map(user => `
            <div class="user-card">
                <h4>${user.first_name} ${user.last_name}</h4>
                <p><strong>Username:</strong> ${user.username || 'N/A'}</p>
                <p><strong>Nickname:</strong> ${user.nickname || 'N/A'}</p>
                <p><strong>About:</strong> ${user.about_me || 'No bio'}</p>
                <div style="margin-top: 10px; display: flex; gap: 10px;">
                    <button onclick="respondToFollowRequest(${user.id}, true)" style="background: #28a745;">Accept</button>
                    <button onclick="respondToFollowRequest(${user.id}, false)" style="background: #dc3545;">Reject</button>
                </div>
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
window.unfollowUser = Users.unfollowUser;
window.getFollowers = Users.getFollowers;
window.getFollowing = Users.getFollowing;
window.getFollowRequests = Users.getFollowRequests;
window.respondToFollowRequest = Users.respondToFollowRequest;
