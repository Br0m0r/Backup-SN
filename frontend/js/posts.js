// Post Management Functions

const Posts = {
    showCreatePost() {
        document.getElementById('createPostForm').classList.remove('hidden');
    },

    hideCreatePost() {
        document.getElementById('createPostForm').classList.add('hidden');
    },

    async createPost() {
        const content = document.getElementById('postContent').value;
        const privacy = document.getElementById('postPrivacy').value;

        if (!content) {
            window.Utils.showStatus('Please enter post content', 'error');
            return;
        }

        const result = await window.Utils.apiCall(`${window.AppConfig.POST_URL}/posts`, 'POST', {
            content,
            privacy_level: privacy,
            viewers: []
        }, true);

        if (result.ok && result.data.success) {
            window.Utils.showStatus('Post created!', 'success');
            window.Utils.showResponse(result.data.data, 'postsResponse');
            Posts.hideCreatePost();
            document.getElementById('postContent').value = '';
            Posts.getFeed();
        } else {
            window.Utils.showStatus(result.data?.error || 'Failed to create post', 'error');
        }
    },

    async getFeed() {
        const result = await window.Utils.apiCall(`${window.AppConfig.POST_URL}/posts`, 'GET', null, true);
        
        if (result.ok && result.data.success) {
            Posts.displayPosts(result.data.data.posts || []);
            window.Utils.showResponse(result.data.data, 'postsResponse');
        } else {
            window.Utils.showStatus(result.data?.error || 'Failed to load feed', 'error');
        }
    },

    async deletePost(postId) {
        if (!confirm('Delete this post?')) return;

        const result = await window.Utils.apiCall(`${window.AppConfig.POST_URL}/posts/${postId}`, 'DELETE', null, true);
        
        if (result.ok && result.data.success) {
            window.Utils.showStatus('Post deleted!', 'success');
            Posts.getFeed();
        } else {
            window.Utils.showStatus(result.data?.error || 'Failed to delete post', 'error');
        }
    },

    async getComments(postId) {
        const result = await window.Utils.apiCall(`${window.AppConfig.POST_URL}/comments?post_id=${postId}`, 'GET', null, true);
        
        if (result.ok && result.data.success) {
            window.Utils.showResponse(result.data.data, 'postsResponse');
            const comments = result.data.data.comments || [];
            window.Utils.showStatus(`Loaded ${comments.length} comments`, 'info');
        } else {
            window.Utils.showStatus(result.data?.error || 'Failed to load comments', 'error');
        }
    },

    displayPosts(posts) {
        const container = document.getElementById('feedContainer');
        if (!posts || posts.length === 0) {
            container.innerHTML = '<p style="color: #999; padding: 20px;">No posts found</p>';
            return;
        }

        const currentUser = window.AppState.getCurrentUser();
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
};

// Export to global scope
window.Posts = Posts;

// Make functions available globally for onclick handlers
window.showCreatePost = Posts.showCreatePost;
window.hideCreatePost = Posts.hideCreatePost;
window.createPost = Posts.createPost;
window.getFeed = Posts.getFeed;
window.deletePost = Posts.deletePost;
window.getComments = Posts.getComments;
