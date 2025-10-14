// Event Management Functions (for Group Events)

const Events = {
    // Show/Hide Event Forms
    showCreateEvent() {
        document.getElementById('createEventForm').classList.remove('hidden');
    },

    hideCreateEvent() {
        document.getElementById('createEventForm').classList.add('hidden');
    },

    // Create Event
    async createGroupEvent() {
        console.log('createGroupEvent called!'); // Debug log
        const currentGroupId = window.AppState.getCurrentGroupId();
        console.log('Current Group ID:', currentGroupId); // Debug log
        
        if (!currentGroupId) {
            console.error('No group ID found!');
            window.Utils.showStatus('Please select a group first', 'error', 'groupsResponse');
            return;
        }

        const title = document.getElementById('eventTitle').value;
        const description = document.getElementById('eventDescription').value;
        const eventTime = document.getElementById('eventTime').value;

        console.log('Event data:', { title, description, eventTime }); // Debug log

        if (!title.trim() || !eventTime) {
            window.Utils.showStatus('Title and event time are required', 'error', 'groupsResponse');
            return;
        }

        const body = {
            group_id: currentGroupId,
            title: title,
            description: description || null,
            event_time: new Date(eventTime).toISOString()
        };

        const result = await window.Utils.apiCall(`${window.AppConfig.GROUP_URL}/events`, 'POST', body, true);
        window.Utils.showResponse(result, 'groupsResponse');

        if (result.ok && result.data && result.data.success) {
            window.Utils.showStatus('Event created!', 'success', 'groupsResponse');
            document.getElementById('eventTitle').value = '';
            document.getElementById('eventDescription').value = '';
            document.getElementById('eventTime').value = '';
            Events.hideCreateEvent();
            Events.viewGroupEvents(); // Refresh events
        } else {
            const errorMsg = result.data?.error || result.error || 'Failed to create event';
            window.Utils.showStatus(errorMsg, 'error', 'groupsResponse');
        }
    },

    // View Group Events
    async viewGroupEvents() {
        const currentGroupId = window.AppState.getCurrentGroupId();
        if (!currentGroupId) return;

        const result = await window.Utils.apiCall(`${window.AppConfig.GROUP_URL}/groups/${currentGroupId}/events`, 'GET', null, true);
        window.Utils.showResponse(result, 'groupsResponse');

        if (result.ok && result.data && result.data.success && result.data.data) {
            document.getElementById('groupEventsList').classList.remove('hidden');
            document.getElementById('groupMembersList').classList.add('hidden');
            document.getElementById('pendingRequestsList').classList.add('hidden');
            
            const container = document.getElementById('eventsContainer');
            if (result.data.data.length === 0) {
                container.innerHTML = '<p style="color: #999;">No events yet</p>';
                return;
            }

            container.innerHTML = result.data.data.map(event => `
                <div class="event-card">
                    <h4>${event.title}</h4>
                    <p>${event.description || 'No description'}</p>
                    <p><small>📅 ${new Date(event.event_time).toLocaleString()}</small></p>
                    <div class="event-responses">
                        <span class="going">✓ Going: ${event.going_count}</span>
                        <span class="not-going">✗ Not Going: ${event.not_going_count}</span>
                        <span class="interested">★ Interested: ${event.interested_count}</span>
                    </div>
                    ${event.user_response ? `<p style="margin-top: 10px;"><strong>Your response: ${event.user_response}</strong></p>` : ''}
                    <div style="margin-top: 10px;">
                        <button onclick="respondToEvent(${event.id}, 'going')">Going</button>
                        <button onclick="respondToEvent(${event.id}, 'not_going')">Not Going</button>
                        <button onclick="respondToEvent(${event.id}, 'interested')">Interested</button>
                    </div>
                </div>
            `).join('');
        } else {
            const errorMsg = result.data?.error || result.error || 'Failed to load events';
            window.Utils.showStatus(errorMsg, 'error', 'groupsResponse');
        }
    },

    // Respond to Event (RSVP)
    async respondToEvent(eventId, response) {
        const result = await window.Utils.apiCall(`${window.AppConfig.GROUP_URL}/events/${eventId}/respond`, 'POST', 
            { event_id: eventId, response: response }, true);
        window.Utils.showResponse(result, 'groupsResponse');

        if (result.ok && result.data && result.data.success) {
            window.Utils.showStatus(`Responded: ${response.replace('_', ' ')}`, 'success', 'groupsResponse');
            Events.viewGroupEvents(); // Refresh events
        } else {
            const errorMsg = result.data?.error || result.error || 'Failed to respond to event';
            window.Utils.showStatus(errorMsg, 'error', 'groupsResponse');
        }
    }
};

// Export to global scope
window.Events = Events;

// Make functions available globally for onclick handlers
window.showCreateEvent = Events.showCreateEvent;
window.hideCreateEvent = Events.hideCreateEvent;
window.createGroupEvent = Events.createGroupEvent;
window.viewGroupEvents = Events.viewGroupEvents;
window.respondToEvent = Events.respondToEvent;
