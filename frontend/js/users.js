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
        const isPublic = document.getElementById('updateIsPublic').checked;

        // Build payload with only non-empty values
        const payload = {};
        if (nickname) payload.nickname = nickname;
        if (aboutMe) payload.about_me = aboutMe;
        payload.is_public_profile = isPublic;

        console.log('Updating profile with payload:', payload);

        const result = await window.Utils.apiCall(`${window.AppConfig.USER_URL}/profile`, 'PUT', payload, true);

        if (result.ok && result.data.success) {
            window.Utils.showResponse(result.data.data, 'profileResponse');
            window.Utils.showStatus('Profile updated!', 'success');
            Users.hideUpdateProfile();
            // Refresh profile display
            const currentUser = window.AppState.getCurrentUser();
            if (currentUser) {
                Users.showOwnProfile();
            }
        } else {
            window.Utils.showStatus(result.data?.error || 'Failed to update profile', 'error');
            console.error('Update profile error:', result);
        }
    },

    async showOwnProfile() {
        const currentUser = window.AppState.getCurrentUser();
        if (!currentUser) {
            window.Utils.showStatus('Not logged in', 'error');
            return;
        }
        await Users.getUserProfile(currentUser.id);
    },

    async getUserProfile(userId) {
        const result = await window.Utils.apiCall(`${window.AppConfig.USER_URL}/users/${userId}/profile`, 'GET', null, true);
        
        if (result.ok && result.data.success) {
            console.log('Profile data received:', result.data.data);
            // Backend returns {profile: {...}}, so we need to extract it
            const profileData = result.data.data.profile || result.data.data;
            Users.displayUserProfile(profileData);
            window.Utils.showResponse(result.data.data, 'profileResponse');
        } else {
            window.Utils.showStatus(result.data?.error || 'Failed to load profile', 'error');
            console.error('Profile fetch error:', result);
        }
    },

    displayUserProfile(profile) {
        const container = document.getElementById('profileDisplay');
        
        // Validate profile data
        if (!profile) {
            container.innerHTML = '<div class="profile-no-access"><h3>Error</h3><p>No profile data received</p></div>';
            console.error('Profile is null or undefined');
            return;
        }

        if (!profile.user) {
            container.innerHTML = '<div class="profile-no-access"><h3>Error</h3><p>User data is missing from profile</p></div>';
            console.error('profile.user is missing:', profile);
            return;
        }

        const currentUser = window.AppState.getCurrentUser();
        const isOwnProfile = currentUser && currentUser.id === profile.user.id;

        if (!profile.can_view) {
            container.innerHTML = `
                <div class="profile-no-access">
                    <h3>🔒 Private Profile</h3>
                    <p>This profile is private. You need to follow this user to view their profile.</p>
                </div>
            `;
            return;
        }

        const user = profile.user;
        const privacyBadge = user.is_public_profile 
            ? '<span class="profile-privacy-badge">🌐 Public</span>' 
            : '<span class="profile-privacy-badge">🔒 Private</span>';

        container.innerHTML = `
            <div class="profile-header">
                <h3>
                    ${user.first_name || ''} ${user.last_name || ''} 
                    ${user.nickname ? `(${user.nickname})` : ''}
                    ${privacyBadge}
                </h3>
                <p><strong>@${user.username}</strong></p>
                ${user.about_me ? `<p style="margin-top: 10px;">${user.about_me}</p>` : ''}
                
                <div class="profile-stats">
                    <div class="stat">
                        <span class="stat-value">${profile.post_count}</span>
                        <span class="stat-label">Posts</span>
                    </div>
                    <div class="stat">
                        <span class="stat-value">${profile.follower_count}</span>
                        <span class="stat-label">Followers</span>
                    </div>
                    <div class="stat">
                        <span class="stat-value">${profile.following_count}</span>
                        <span class="stat-label">Following</span>
                    </div>
                </div>
            </div>

            ${isOwnProfile ? '' : `
                <div style="margin-bottom: 20px;">
                    <button onclick="Users.viewUserProfile(${user.id})" style="width: auto;">
                        View Full Profile
                    </button>
                </div>
            `}

            <div class="profile-section">
                <h4>📝 Posts (${profile.post_count})</h4>
                <div class="profile-posts-grid">
                    ${profile.posts && profile.posts.length > 0 
                        ? profile.posts.map(post => `
                            <div class="post-card">
                                <h5>${post.title || 'Untitled Post'}</h5>
                                <p>${post.content}</p>
                                <small>
                                    <span class="privacy-badge">${post.privacy_level}</span>
                                    • ${new Date(post.created_at).toLocaleString()}
                                </small>
                            </div>
                        `).join('')
                        : '<p style="color: #999;">No posts yet</p>'
                    }
                </div>
            </div>

            <div class="profile-section">
                <h4>👥 Followers (${profile.follower_count})</h4>
                <div class="profile-users-grid">
                    ${profile.followers && profile.followers.length > 0
                        ? profile.followers.map(follower => `
                            <div class="user-card">
                                <h5>${follower.first_name || ''} ${follower.last_name || ''}</h5>
                                <p><strong>@${follower.username}</strong></p>
                                ${follower.nickname ? `<p>${follower.nickname}</p>` : ''}
                                <button onclick="Users.getUserProfile(${follower.id})" style="margin-top: 10px;">
                                    View Profile
                                </button>
                            </div>
                        `).join('')
                        : '<p style="color: #999;">No followers yet</p>'
                    }
                </div>
            </div>

            <div class="profile-section">
                <h4>👤 Following (${profile.following_count})</h4>
                <div class="profile-users-grid">
                    ${profile.following && profile.following.length > 0
                        ? profile.following.map(following => `
                            <div class="user-card">
                                <h5>${following.first_name || ''} ${following.last_name || ''}</h5>
                                <p><strong>@${following.username}</strong></p>
                                ${following.nickname ? `<p>${following.nickname}</p>` : ''}
                                <button onclick="Users.getUserProfile(${following.id})" style="margin-top: 10px;">
                                    View Profile
                                </button>
                            </div>
                        `).join('')
                        : '<p style="color: #999;">Not following anyone yet</p>'
                    }
                </div>
            </div>
        `;
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
                <div style="margin-top: 10px; display: flex; gap: 10px;">
                    <button onclick="Users.getUserProfile(${user.id})">View Profile</button>
                    ${user.followStatus === 'none' ? `<button onclick="Users.followUser(${user.id})">Follow</button>` : ''}
                    ${user.followStatus === 'pending' ? `<button disabled style="opacity: 0.6; cursor: not-allowed; background: #ffc107;">Request Pending</button>` : ''}
                    ${user.followStatus === 'accepted' ? `<button onclick="Users.unfollowUser(${user.id})" style="background: #28a745;">Following</button>` : ''}
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
                <div style="margin-top: 10px; display: flex; gap: 10px;">
                    <button onclick="Users.getUserProfile(${user.id})">View Profile</button>
                    ${showUnfollow ? `<button onclick="Users.unfollowUser(${user.id})" style="background: #dc3545;">Unfollow</button>` : ''}
                </div>
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
window.getUserProfile = Users.getUserProfile;
window.showOwnProfile = Users.showOwnProfile;
