// Chat Module - WebSocket and REST API for 1-on-1 and Group Chat

const CHAT_URL = `${AppConfig.API_BASE}:8085`;
const CHAT_WS_URL = `ws://localhost:8085/ws`;

// WebSocket connection
let chatWebSocket = null;
let reconnectAttempts = 0;
const MAX_RECONNECT_ATTEMPTS = 5;

// Chat state
let activeChatType = null; // 'private' or 'group'
let activeChatId = null; // user_id or group_id
let activeChatName = null;

// Initialize WebSocket connection
function connectChatWebSocket() {
    const token = AppState.getToken();
    if (!token) {
        console.error('No token available for WebSocket connection');
        return;
    }

    console.log('Connecting to chat WebSocket...');
    chatWebSocket = new WebSocket(`${CHAT_WS_URL}?token=${token}`);

    chatWebSocket.onopen = () => {
        console.log('✅ Chat WebSocket connected');
        reconnectAttempts = 0;
        updateConnectionStatus(true);
    };

    chatWebSocket.onmessage = (event) => {
        console.log('📨 Message received:', event.data);
        try {
            const message = JSON.parse(event.data);
            handleIncomingMessage(message);
        } catch (error) {
            console.error('Failed to parse message:', error);
        }
    };

    chatWebSocket.onerror = (error) => {
        console.error('❌ WebSocket error:', error);
        updateConnectionStatus(false);
    };

    chatWebSocket.onclose = () => {
        console.log('🔌 WebSocket closed');
        updateConnectionStatus(false);
        chatWebSocket = null;

        // Attempt to reconnect
        if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
            reconnectAttempts++;
            console.log(`Reconnecting... attempt ${reconnectAttempts}/${MAX_RECONNECT_ATTEMPTS}`);
            setTimeout(connectChatWebSocket, 2000 * reconnectAttempts);
        }
    };
}

// Disconnect WebSocket
function disconnectChatWebSocket() {
    if (chatWebSocket) {
        chatWebSocket.close();
        chatWebSocket = null;
    }
    updateConnectionStatus(false);
}

// Update connection status indicator
function updateConnectionStatus(connected) {
    const statusEl = document.getElementById('chatConnectionStatus');
    if (statusEl) {
        if (connected) {
            statusEl.innerHTML = '🟢 Connected';
            statusEl.className = 'connection-status connected';
        } else {
            statusEl.innerHTML = '🔴 Disconnected';
            statusEl.className = 'connection-status disconnected';
        }
    }
}

// Handle incoming WebSocket message
function handleIncomingMessage(message) {
    console.log('Handling message:', message);

    if (message.type === 'error') {
        showChatError(message.content || 'An error occurred');
        return;
    }

    // Check if message is for active chat
    let isForActiveChat = false;

    if (message.type === 'message' && activeChatType === 'private') {
        // 1-on-1 message
        if (message.sender_id === activeChatId || message.receiver_id === activeChatId) {
            isForActiveChat = true;
        }
    } else if (message.type === 'group_message' && activeChatType === 'group') {
        // Group message
        if (message.group_id === activeChatId) {
            isForActiveChat = true;
        }
    }

    if (isForActiveChat) {
        displayMessage(message);
    } else {
        // Show notification for other chats
        showChatNotification(message);
    }
}

// Display message in chat window
function displayMessage(message) {
    const messagesContainer = document.getElementById('chatMessages');
    if (!messagesContainer) return;

    const currentUserId = AppState.getCurrentUser()?.id;
    const isOwnMessage = message.sender_id === currentUserId;

    const messageEl = document.createElement('div');
    messageEl.className = `chat-message ${isOwnMessage ? 'own-message' : 'other-message'}`;

    const time = new Date(message.created_at || Date.now()).toLocaleTimeString('en-US', {
        hour: '2-digit',
        minute: '2-digit'
    });

    let senderInfo = '';
    if (!isOwnMessage) {
        senderInfo = `<div class="message-sender">User ${message.sender_id}</div>`;
    }

    messageEl.innerHTML = `
        ${senderInfo}
        <div class="message-content">${Utils.escapeHtml(message.content)}</div>
        <div class="message-time">${time}</div>
    `;

    messagesContainer.appendChild(messageEl);
    messagesContainer.scrollTop = messagesContainer.scrollHeight;
}

// Show notification for messages in other chats
function showChatNotification(message) {
    const notifEl = document.getElementById('chatNotifications');
    if (!notifEl) return;

    const notif = document.createElement('div');
    notif.className = 'chat-notification';
    
    let chatInfo = '';
    if (message.type === 'group_message') {
        chatInfo = `Group ${message.group_id}`;
    } else {
        chatInfo = `User ${message.sender_id}`;
    }

    notif.innerHTML = `
        📩 New message from ${chatInfo}
        <button onclick="Chat.dismissNotification(this)">×</button>
    `;

    notifEl.appendChild(notif);

    // Auto-dismiss after 5 seconds
    setTimeout(() => {
        notif.remove();
    }, 5000);
}

// Send message via WebSocket
function sendMessage() {
    const input = document.getElementById('chatMessageInput');
    const content = input.value.trim();

    if (!content) {
        showChatError('Message cannot be empty');
        return;
    }

    if (!chatWebSocket || chatWebSocket.readyState !== WebSocket.OPEN) {
        showChatError('Not connected to chat server');
        return;
    }

    const message = {
        type: activeChatType === 'group' ? 'group_message' : 'message',
        content: content
    };

    if (activeChatType === 'group') {
        message.group_id = activeChatId;
    } else {
        message.receiver_id = activeChatId;
    }

    console.log('Sending message:', message);
    chatWebSocket.send(JSON.stringify(message));

    // Clear input
    input.value = '';
}

// Load chat conversations list
async function loadConversations() {
    try {
        const response = await fetch(`${CHAT_URL}/chat/conversations`, {
            headers: {
                'Authorization': `Bearer ${AppState.getToken()}`
            }
        });

        const result = await response.json();
        console.log('Conversations:', result);

        if (result.data && result.data.conversations) {
            displayConversationsList(result.data.conversations);
        } else {
            document.getElementById('conversationsList').innerHTML = '<p style="color: #666;">No conversations yet</p>';
        }
    } catch (error) {
        console.error('Error loading conversations:', error);
        showChatError('Failed to load conversations');
    }
}

// Display conversations list
function displayConversationsList(conversations) {
    const container = document.getElementById('conversationsList');
    container.innerHTML = '';

    if (conversations.length === 0) {
        container.innerHTML = '<p style="color: #666;">No conversations yet</p>';
        return;
    }

    conversations.forEach(conv => {
        const convEl = document.createElement('div');
        convEl.className = 'conversation-item';
        convEl.onclick = () => openPrivateChat(conv.user_id, `User ${conv.user_id}`);

        const time = new Date(conv.last_message_time).toLocaleTimeString('en-US', {
            hour: '2-digit',
            minute: '2-digit'
        });

        convEl.innerHTML = `
            <div class="conversation-info">
                <div class="conversation-name">👤 User ${conv.user_id}</div>
                <div class="conversation-preview">${Utils.escapeHtml(conv.last_message.substring(0, 50))}${conv.last_message.length > 50 ? '...' : ''}</div>
            </div>
            <div class="conversation-time">${time}</div>
        `;

        container.appendChild(convEl);
    });
}

// Load groups for group chat
async function loadGroupsForChat() {
    try {
        const response = await fetch(`${AppConfig.GROUP_URL}/groups`, {
            headers: {
                'Authorization': `Bearer ${AppState.getToken()}`
            }
        });

        const result = await response.json();
        console.log('Groups for chat:', result);

        if (result.data && result.data.groups) {
            displayGroupsList(result.data.groups);
        } else {
            document.getElementById('groupChatsList').innerHTML = '<p style="color: #666;">No groups yet. Create or join a group first!</p>';
        }
    } catch (error) {
        console.error('Error loading groups:', error);
        showChatError('Failed to load groups');
    }
}

// Display groups list for chat
function displayGroupsList(groups) {
    const container = document.getElementById('groupChatsList');
    container.innerHTML = '';

    // Filter only groups where user is a member
    const memberGroups = groups.filter(g => g.is_member);

    if (memberGroups.length === 0) {
        container.innerHTML = '<p style="color: #666;">You are not a member of any groups yet</p>';
        return;
    }

    memberGroups.forEach(group => {
        const groupEl = document.createElement('div');
        groupEl.className = 'conversation-item';
        groupEl.onclick = () => openGroupChat(group.id, group.name);

        groupEl.innerHTML = `
            <div class="conversation-info">
                <div class="conversation-name">👥 ${Utils.escapeHtml(group.name)}</div>
                <div class="conversation-preview">${group.member_count} members</div>
            </div>
        `;

        container.appendChild(groupEl);
    });
}

// Open private chat
async function openPrivateChat(userId, userName) {
    activeChatType = 'private';
    activeChatId = userId;
    activeChatName = userName;

    // Show chat window
    document.getElementById('chatWindow').classList.remove('hidden');
    document.getElementById('chatSidebar').classList.remove('hidden');
    document.getElementById('chatHeader').innerHTML = `
        <button onclick="Chat.closeChat()">← Back</button>
        <span>💬 Chat with ${Utils.escapeHtml(userName)} (ID: ${userId})</span>
    `;

    // Clear messages
    document.getElementById('chatMessages').innerHTML = '<p style="color: #666; text-align: center;">Loading messages...</p>';

    // Load message history
    await loadPrivateChatHistory(userId);
}

// Load private chat history
async function loadPrivateChatHistory(userId) {
    try {
        const response = await fetch(`${CHAT_URL}/chat/history/${userId}`, {
            headers: {
                'Authorization': `Bearer ${AppState.getToken()}`
            }
        });

        const result = await response.json();
        console.log('Chat history:', result);

        const messagesContainer = document.getElementById('chatMessages');
        messagesContainer.innerHTML = '';

        if (result.data && result.data.messages && result.data.messages.length > 0) {
            result.data.messages.forEach(msg => displayMessage(msg));
        } else {
            messagesContainer.innerHTML = '<p style="color: #666; text-align: center;">No messages yet. Start the conversation!</p>';
        }
    } catch (error) {
        console.error('Error loading chat history:', error);
        showChatError('Failed to load chat history');
    }
}

// Open group chat
async function openGroupChat(groupId, groupName) {
    activeChatType = 'group';
    activeChatId = groupId;
    activeChatName = groupName;

    // Show chat window
    document.getElementById('chatWindow').classList.remove('hidden');
    document.getElementById('chatSidebar').classList.remove('hidden');
    document.getElementById('chatHeader').innerHTML = `
        <button onclick="Chat.closeChat()">← Back</button>
        <span>👥 ${Utils.escapeHtml(groupName)} (Group Chat)</span>
    `;

    // Clear messages
    document.getElementById('chatMessages').innerHTML = '<p style="color: #666; text-align: center;">Loading messages...</p>';

    // Load message history
    await loadGroupChatHistory(groupId);
}

// Load group chat history
async function loadGroupChatHistory(groupId) {
    try {
        const response = await fetch(`${CHAT_URL}/chat/groups/${groupId}/history?limit=50`, {
            headers: {
                'Authorization': `Bearer ${AppState.getToken()}`
            }
        });

        const result = await response.json();
        console.log('Group chat history:', result);

        const messagesContainer = document.getElementById('chatMessages');
        messagesContainer.innerHTML = '';

        if (result.data && result.data.messages && result.data.messages.length > 0) {
            result.data.messages.forEach(msg => displayMessage(msg));
        } else {
            messagesContainer.innerHTML = '<p style="color: #666; text-align: center;">No messages yet. Start the conversation!</p>';
        }
    } catch (error) {
        console.error('Error loading group chat history:', error);
        showChatError('Failed to load group chat history');
    }
}

// Close chat window
function closeChat() {
    activeChatType = null;
    activeChatId = null;
    activeChatName = null;

    document.getElementById('chatWindow').classList.add('hidden');
    document.getElementById('chatMessages').innerHTML = '';
}

// Show new chat dialog
function showNewChatDialog() {
    const userId = prompt('Enter User ID to start a chat:');
    if (userId) {
        const userIdNum = parseInt(userId);
        if (!isNaN(userIdNum) && userIdNum > 0) {
            openPrivateChat(userIdNum, `User ${userIdNum}`);
        } else {
            showChatError('Invalid user ID');
        }
    }
}

// Show error in chat
function showChatError(message) {
    const errorDiv = document.getElementById('chatError');
    if (errorDiv) {
        errorDiv.textContent = `❌ ${message}`;
        errorDiv.style.display = 'block';
        setTimeout(() => {
            errorDiv.style.display = 'none';
        }, 5000);
    }
}

// Dismiss notification
function dismissNotification(button) {
    button.parentElement.remove();
}

// Handle Enter key in message input
function handleChatKeyPress(event) {
    if (event.key === 'Enter' && !event.shiftKey) {
        event.preventDefault();
        sendMessage();
    }
}

// Export Chat namespace
window.Chat = {
    connectWebSocket: connectChatWebSocket,
    disconnectWebSocket: disconnectChatWebSocket,
    loadConversations: loadConversations,
    loadGroupsForChat: loadGroupsForChat,
    openPrivateChat: openPrivateChat,
    openGroupChat: openGroupChat,
    closeChat: closeChat,
    sendMessage: sendMessage,
    showNewChatDialog: showNewChatDialog,
    dismissNotification: dismissNotification,
    handleChatKeyPress: handleChatKeyPress
};

// Initialize chat when user logs in
document.addEventListener('userLoggedIn', () => {
    console.log('User logged in, connecting to chat WebSocket...');
    connectChatWebSocket();
});

// Cleanup on logout
document.addEventListener('userLoggedOut', () => {
    console.log('User logged out, disconnecting chat WebSocket...');
    disconnectChatWebSocket();
});
