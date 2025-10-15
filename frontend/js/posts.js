// Post Management Functions

const Posts = {
    showCreatePost() {
        document.getElementById('createPostForm').classList.remove('hidden');
    },

    hideCreatePost() {
        document.getElementById('createPostForm').classList.add('hidden');
    },

    async createPost() {
        const title = document.getElementById('postTitle').value.trim();
        const content = document.getElementById('postContent').value;
        const privacy = document.getElementById('postPrivacy').value;

        if (!content) {
            window.Utils.showStatus('Please enter post content', 'error');
            return;
        }

        const payload = {
            content,
            privacy_level: privacy,
            viewers: []
        };
        
        if (title) {
            payload.title = title;
        }

        const result = await window.Utils.apiCall(`${window.AppConfig.POST_URL}/posts`, 'POST', payload, true);

        if (result.ok && result.data.success) {
            window.Utils.showStatus('Post created!', 'success');
            window.Utils.showResponse(result.data.data, 'postsResponse');
            Posts.hideCreatePost();
            document.getElementById('postTitle').value = '';
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

    async createComment(postId) {
        const textareaId = `commentContent-${postId}`;
        const content = document.getElementById(textareaId).value.trim();

        if (!content) {
            window.Utils.showStatus('Please enter comment content', 'error');
            return;
        }

        const result = await window.Utils.apiCall(`${window.AppConfig.POST_URL}/comments`, 'POST', {
            post_id: postId,
            content: content
        }, true);

        if (result.ok && result.data.success) {
            window.Utils.showStatus('Comment added!', 'success');
            document.getElementById(textareaId).value = '';
            Posts.loadCommentsForPost(postId);
        } else {
            window.Utils.showStatus(result.data?.error || 'Failed to add comment', 'error');
        }
    },

    async loadCommentsForPost(postId) {
        const result = await window.Utils.apiCall(`${window.AppConfig.POST_URL}/comments?post_id=${postId}`, 'GET', null, true);
        
        const commentsContainer = document.getElementById(`comments-${postId}`);
        if (result.ok && result.data.success) {
            const comments = result.data.data.comments || [];
            if (comments.length === 0) {
                commentsContainer.innerHTML = '<p style="color: #999; margin: 10px 0;">No comments yet</p>';
            } else {
                commentsContainer.innerHTML = comments.map(comment => `
                    <div class="comment-card">
                        <p>${comment.content}</p>
                        <small>By User ID: ${comment.user_id} • ${new Date(comment.created_at).toLocaleString()}</small>
                    </div>
                `).join('');
            }
        } else {
            commentsContainer.innerHTML = '<p style="color: #999;">Failed to load comments</p>';
        }
    },

    toggleComments(postId) {
        const commentsSection = document.getElementById(`commentsSection-${postId}`);
        if (commentsSection.style.display === 'none' || !commentsSection.style.display) {
            commentsSection.style.display = 'block';
            Posts.loadCommentsForPost(postId);
        } else {
            commentsSection.style.display = 'none';
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
                <h4>${post.title || 'Untitled Post'} <small>(${post.privacy_level})</small></h4>
                <p>${post.content}</p>
                <p><small>By User ID: ${post.user_id} • ${new Date(post.created_at).toLocaleString()}</small></p>
                <div class="post-actions">
                    <button onclick="Posts.toggleComments(${post.id})">💬 Comments</button>
                    ${post.user_id === currentUser?.id ? `<button onclick="Posts.deletePost(${post.id})">Delete</button>` : ''}
                </div>
                <div id="commentsSection-${post.id}" style="display: none; margin-top: 15px; border-top: 1px solid #ddd; padding-top: 15px;">
                    <div class="comment-form" style="margin-bottom: 15px;">
                        <textarea id="commentContent-${post.id}" placeholder="Write a comment..." rows="2" style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px;"></textarea>
                        <button onclick="Posts.createComment(${post.id})" style="margin-top: 5px;">Post Comment</button>
                    </div>
                    <div id="comments-${post.id}" class="comments-list"></div>
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
