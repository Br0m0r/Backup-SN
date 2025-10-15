// Utility Functions

function showStatus(message, type, elementId = 'globalStatus') {
    const el = document.getElementById(elementId);
    if (!el) return;
    el.innerHTML = `<div class="status ${type}">${message}</div>`;
    setTimeout(() => el.innerHTML = '', 5000);
}

function showResponse(data, elementId) {
    const el = document.getElementById(elementId);
    if (!el) return;
    el.innerHTML = `<pre>${JSON.stringify(data, null, 2)}</pre>`;
}

async function apiCall(url, method = 'GET', body = null, useToken = false) {
    const options = {
        method,
        headers: {
            'Content-Type': 'application/json',
        }
    };

    if (useToken && window.AppState.getToken()) {
        options.headers['Authorization'] = `Bearer ${window.AppState.getToken()}`;
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

// Tab Switching
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

function switchMainTab(tab) {
    document.querySelectorAll('#mainSection .tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('#mainSection .tab-content').forEach(c => c.classList.remove('active'));
    
    const tabs = ['profile', 'posts', 'users', 'groups'];
    const index = tabs.indexOf(tab);
    document.querySelectorAll('#mainSection .tab')[index].classList.add('active');
    document.getElementById(tab + 'Tab').classList.add('active');
    
    // Auto-load groups when switching to groups tab
    if (tab === 'groups' && window.Groups && window.Groups.browseGroups) {
        window.Groups.browseGroups();
    }
    
    // Auto-load user search when switching to users tab
    if (tab === 'users') {
        switchUserSubTab('search');
    }
}

// Sub-tab switching for Users section
function switchUserSubTab(subTab) {
    document.querySelectorAll('.sub-tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('.sub-tab-content').forEach(c => c.classList.remove('active'));
    
    const subTabs = ['search', 'followers', 'following', 'requests'];
    const index = subTabs.indexOf(subTab);
    
    // Activate the clicked sub-tab button
    document.querySelectorAll('.sub-tab')[index].classList.add('active');
    
    // Activate the corresponding sub-tab content
    const contentIds = {
        'search': 'searchUsersSubTab',
        'followers': 'followersSubTab',
        'following': 'followingSubTab',
        'requests': 'requestsSubTab'
    };
    
    const contentElement = document.getElementById(contentIds[subTab]);
    if (contentElement) {
        contentElement.classList.add('active');
    }
}

// Export to global scope
window.Utils = {
    showStatus,
    showResponse,
    apiCall,
    switchAuthTab,
    switchMainTab
};

// Make tab switching available globally
window.switchUserSubTab = switchUserSubTab;
