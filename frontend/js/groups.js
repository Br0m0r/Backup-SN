// Group Management Functions

const Groups = {
    // Show/Hide Group Forms
    showCreateGroup() {
        document.getElementById('createGroupForm').classList.remove('hidden');
        document.getElementById('groupsListContainer').classList.add('hidden');
    },

    hideCreateGroup() {
        document.getElementById('createGroupForm').classList.add('hidden');
        document.getElementById('groupsListContainer').classList.remove('hidden');
    },

    showInviteMember() {
        document.getElementById('inviteMemberForm').classList.remove('hidden');
    },

    hideInviteMember() {
        document.getElementById('inviteMemberForm').classList.add('hidden');
    },

    hideGroupDetail() {
        document.getElementById('groupDetailView').classList.add('hidden');
        document.getElementById('groupsListContainer').classList.remove('hidden');
        window.AppState.setCurrentGroupId(null);
    },

    // Create Group
    async createGroup() {
        const name = document.getElementById('groupName').value;
        const description = document.getElementById('groupDescription').value;
        const imageURL = document.getElementById('groupImageURL').value;

        if (!name.trim()) {
            window.Utils.showStatus('Group name is required', 'error', 'groupsResponse');
            return;
        }

        const body = {
            name: name,
            description: description || null,
            image_url: imageURL || null
        };

        const result = await window.Utils.apiCall(`${window.AppConfig.GROUP_URL}/groups`, 'POST', body, true);
        window.Utils.showResponse(result, 'groupsResponse');

        if (result.ok && result.data && result.data.success) {
            window.Utils.showStatus('Group created successfully!', 'success', 'groupsResponse');
            Groups.hideCreateGroup();
            // Clear form
            document.getElementById('groupName').value = '';
            document.getElementById('groupDescription').value = '';
            document.getElementById('groupImageURL').value = '';
            // Refresh groups list
            Groups.browseGroups();
        } else {
            const errorMsg = result.data?.error || result.error || 'Failed to create group';
            window.Utils.showStatus(errorMsg, 'error', 'groupsResponse');
        }
    },

    // Browse All Groups
    async browseGroups() {
        const result = await window.Utils.apiCall(`${window.AppConfig.GROUP_URL}/groups`, 'GET', null, true);
        window.Utils.showResponse(result, 'groupsResponse');

        if (result.ok && result.data && result.data.success && result.data.data) {
            Groups.displayGroups(result.data.data);
        } else {
            const errorMsg = result.data?.error || result.error || 'Failed to load groups';
            window.Utils.showStatus(errorMsg, 'error', 'groupsResponse');
        }
    },

    // Get My Groups (where user is a member)
    async getMyGroups() {
        const result = await window.Utils.apiCall(`${window.AppConfig.GROUP_URL}/groups`, 'GET', null, true);
        window.Utils.showResponse(result, 'groupsResponse');

        if (result.ok && result.data && result.data.success && result.data.data) {
            const myGroups = result.data.data.filter(g => g.is_member);
            Groups.displayGroups(myGroups);
            if (myGroups.length === 0) {
                document.getElementById('groupsListContainer').innerHTML = 
                    '<p style="color: #999; padding: 20px;">You are not a member of any groups yet.</p>';
            }
        } else {
            const errorMsg = result.data?.error || result.error || 'Failed to load groups';
            window.Utils.showStatus(errorMsg, 'error', 'groupsResponse');
        }
    },

    // Display Groups List
    displayGroups(groups) {
        const container = document.getElementById('groupsListContainer');
        container.classList.remove('hidden');
        
        if (!groups || groups.length === 0) {
            container.innerHTML = '<p style="color: #999; padding: 20px;">No groups found</p>';
            return;
        }

        container.innerHTML = groups.map(group => `
            <div class="group-card">
                <h3>${group.name}</h3>
                <p>${group.description || 'No description'}</p>
                <div class="group-meta">
                    <span>👥 ${group.member_count} member${group.member_count !== 1 ? 's' : ''}</span>
                    <span>📅 Created ${new Date(group.created_at).toLocaleDateString()}</span>
                </div>
                <div style="margin-top: 15px;">
                    ${group.is_creator ? '<span class="group-badge creator">👑 Creator</span>' : ''}
                    ${group.is_member && !group.is_creator ? '<span class="group-badge member">✓ Member</span>' : ''}
                    ${group.has_pending_request ? '<span class="group-badge pending">⏳ Request Pending</span>' : ''}
                    ${!group.is_member && !group.has_pending_request ? '<span class="group-badge not-member">Not a member</span>' : ''}
                </div>
                <div style="margin-top: 15px;">
                    <button onclick="viewGroupDetail(${group.id})">View Details</button>
                    ${!group.is_member && !group.has_pending_request ? `<button onclick="requestToJoinGroup(${group.id})">Request to Join</button>` : ''}
                    ${group.has_pending_request ? `<button disabled style="opacity: 0.6; cursor: not-allowed;">Request Pending</button>` : ''}
                </div>
            </div>
        `).join('');
    },

    // View Group Detail
    async viewGroupDetail(groupId) {
        window.AppState.setCurrentGroupId(groupId);
        const result = await window.Utils.apiCall(`${window.AppConfig.GROUP_URL}/groups/${groupId}`, 'GET', null, true);
        window.Utils.showResponse(result, 'groupsResponse');

        if (result.ok && result.data && result.data.success && result.data.data) {
            const group = result.data.data;
            document.getElementById('groupsListContainer').classList.add('hidden');
            document.getElementById('groupDetailView').classList.remove('hidden');
            
            document.getElementById('groupDetailContent').innerHTML = `
                <h2>${group.name}</h2>
                <p>${group.description || 'No description'}</p>
                ${group.image_url ? `<img src="${group.image_url}" alt="${group.name}" style="max-width: 200px; border-radius: 8px; margin: 10px 0;">` : ''}
                <div class="group-meta" style="margin-top: 15px;">
                    <span>👥 ${group.member_count} member${group.member_count !== 1 ? 's' : ''}</span>
                    <span>📅 Created ${new Date(group.created_at).toLocaleDateString()}</span>
                </div>
                <div style="margin-top: 10px;">
                    ${group.is_creator ? '<span class="group-badge creator">👑 You are the creator</span>' : ''}
                    ${group.is_member && !group.is_creator ? '<span class="group-badge member">✓ You are a member</span>' : ''}
                </div>
            `;

            // Show appropriate actions
            if (group.is_member) {
                document.getElementById('groupMemberActions').classList.remove('hidden');
            } else {
                document.getElementById('groupMemberActions').classList.add('hidden');
            }

            if (group.is_creator) {
                document.getElementById('groupCreatorActions').classList.remove('hidden');
            } else {
                document.getElementById('groupCreatorActions').classList.add('hidden');
            }

            // Auto-load events
            if (window.Events && window.Events.viewGroupEvents) {
                window.Events.viewGroupEvents();
            }
        } else {
            const errorMsg = result.data?.error || result.error || 'Failed to load group details';
            window.Utils.showStatus(errorMsg, 'error', 'groupsResponse');
        }
    },

    // Request to Join Group
    async requestToJoinGroup(groupId) {
        const result = await window.Utils.apiCall(`${window.AppConfig.GROUP_URL}/groups/${groupId}/request`, 'POST', {}, true);
        window.Utils.showResponse(result, 'groupsResponse');

        if (result.ok && result.data && result.data.success) {
            window.Utils.showStatus('Join request sent! Waiting for approval.', 'success', 'groupsResponse');
            Groups.browseGroups(); // Refresh list
        } else {
            const errorMsg = result.data?.error || result.error || 'Failed to send join request';
            window.Utils.showStatus(errorMsg, 'error', 'groupsResponse');
        }
    },

    // Invite Member
    async inviteMember() {
        const currentGroupId = window.AppState.getCurrentGroupId();
        if (!currentGroupId) return;

        const userId = parseInt(document.getElementById('inviteUserId').value);
        if (!userId) {
            window.Utils.showStatus('Please enter a valid user ID', 'error', 'groupsResponse');
            return;
        }

        const result = await window.Utils.apiCall(`${window.AppConfig.GROUP_URL}/groups/${currentGroupId}/invite`, 'POST', { user_id: userId }, true);
        window.Utils.showResponse(result, 'groupsResponse');

        if (result.ok && result.data && result.data.success) {
            window.Utils.showStatus('Invitation sent!', 'success', 'groupsResponse');
            document.getElementById('inviteUserId').value = '';
            Groups.hideInviteMember();
        } else {
            const errorMsg = result.data?.error || result.error || 'Failed to send invitation';
            window.Utils.showStatus(errorMsg, 'error', 'groupsResponse');
        }
    },

    // View Group Members
    async viewGroupMembers() {
        const currentGroupId = window.AppState.getCurrentGroupId();
        if (!currentGroupId) return;

        const result = await window.Utils.apiCall(`${window.AppConfig.GROUP_URL}/groups/${currentGroupId}/members`, 'GET', null, true);
        window.Utils.showResponse(result, 'groupsResponse');

        if (result.ok && result.data && result.data.success && result.data.data) {
            document.getElementById('groupMembersList').classList.remove('hidden');
            document.getElementById('groupEventsList').classList.add('hidden');
            document.getElementById('pendingRequestsList').classList.add('hidden');
            
            const container = document.getElementById('membersContainer');
            if (result.data.data.length === 0) {
                container.innerHTML = '<p style="color: #999;">No members found</p>';
                return;
            }

            container.innerHTML = result.data.data.map(member => `
                <div class="member-card">
                    <div class="member-info">
                        <strong>User ID: ${member.user_id}</strong>
                        <span class="member-role ${member.role}">${member.role.toUpperCase()}</span>
                        <br><small>Joined: ${new Date(member.joined_at).toLocaleDateString()}</small>
                    </div>
                </div>
            `).join('');
        } else {
            const errorMsg = result.data?.error || result.error || 'Failed to load members';
            window.Utils.showStatus(errorMsg, 'error', 'groupsResponse');
        }
    },

    // View Pending Requests (Creator only)
    async viewPendingRequests() {
        const currentGroupId = window.AppState.getCurrentGroupId();
        if (!currentGroupId) return;

        const result = await window.Utils.apiCall(`${window.AppConfig.GROUP_URL}/groups/${currentGroupId}/requests`, 'GET', null, true);
        window.Utils.showResponse(result, 'groupsResponse');

        if (result.ok && result.data && result.data.success && result.data.data) {
            document.getElementById('pendingRequestsList').classList.remove('hidden');
            document.getElementById('groupMembersList').classList.add('hidden');
            document.getElementById('groupEventsList').classList.add('hidden');
            
            const container = document.getElementById('requestsContainer');
            if (result.data.data.length === 0) {
                container.innerHTML = '<p style="color: #999;">No pending requests</p>';
                return;
            }

            container.innerHTML = result.data.data.map(request => `
                <div class="request-card">
                    <div class="request-info">
                        <strong>User ID: ${request.user_id}</strong>
                        <br><small>Requested: ${new Date(request.joined_at).toLocaleDateString()}</small>
                    </div>
                    <div class="request-actions">
                        <button onclick="respondToRequest(${request.id}, true)" style="background: #28a745;">Accept</button>
                        <button onclick="respondToRequest(${request.id}, false)" style="background: #dc3545;">Reject</button>
                    </div>
                </div>
            `).join('');
        } else {
            const errorMsg = result.data?.error || result.error || 'Failed to load requests';
            window.Utils.showStatus(errorMsg, 'error', 'groupsResponse');
        }
    },

    // Respond to Join Request
    async respondToRequest(memberId, accept) {
        const currentGroupId = window.AppState.getCurrentGroupId();
        if (!currentGroupId) return;

        const result = await window.Utils.apiCall(`${window.AppConfig.GROUP_URL}/groups/${currentGroupId}/requests/respond`, 'POST', 
            { member_id: memberId, accept: accept }, true);
        window.Utils.showResponse(result, 'groupsResponse');

        if (result.ok && result.data && result.data.success) {
            window.Utils.showStatus(`Request ${accept ? 'accepted' : 'rejected'}!`, 'success', 'groupsResponse');
            Groups.viewPendingRequests(); // Refresh list
        } else {
            const errorMsg = result.data?.error || result.error || 'Failed to respond to request';
            window.Utils.showStatus(errorMsg, 'error', 'groupsResponse');
        }
    }
};

// Export to global scope
window.Groups = Groups;

// Make functions available globally for onclick handlers
window.showCreateGroup = Groups.showCreateGroup;
window.hideCreateGroup = Groups.hideCreateGroup;
window.showInviteMember = Groups.showInviteMember;
window.hideInviteMember = Groups.hideInviteMember;
window.hideGroupDetail = Groups.hideGroupDetail;
window.createGroup = Groups.createGroup;
window.browseGroups = Groups.browseGroups;
window.getMyGroups = Groups.getMyGroups;
window.viewGroupDetail = Groups.viewGroupDetail;
window.requestToJoinGroup = Groups.requestToJoinGroup;
window.inviteMember = Groups.inviteMember;
window.viewGroupMembers = Groups.viewGroupMembers;
window.viewPendingRequests = Groups.viewPendingRequests;
window.respondToRequest = Groups.respondToRequest;
