package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// Config
var notificationServiceURL string

func init() {
	notificationServiceURL = os.Getenv("NOTIFICATION_SERVICE_URL")
	if notificationServiceURL == "" {
		// Default to localhost for local development
		// Use http://notification-service:8086 when running in Docker
		notificationServiceURL = "http://localhost:8086"
	}
}

// ============================================
// CORE FUNCTION (called by all helpers below)
// ============================================

// createNotification makes an HTTP call to the notification service
func createNotification(userID int, notifType, content string, relatedID int) error {
	payload := map[string]interface{}{
		"user_id":    userID,
		"type":       notifType,
		"related_id": relatedID,
		"content":    content,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[Notify] Failed to marshal notification: %v", err)
		return err
	}

	resp, err := http.Post(
		notificationServiceURL+"/notifications",
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		log.Printf("[Notify] HTTP error calling notification service: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[Notify] Notification service returned status: %d", resp.StatusCode)
		return fmt.Errorf("notification service error: %d", resp.StatusCode)
	}

	log.Printf("[Notify] Created notification: type=%s, user=%d", notifType, userID)
	return nil
}

// ============================================
// FOLLOW SYSTEM NOTIFICATIONS
// ============================================

// FollowRequest notifies user about a follow request (private profile)
func FollowRequest(targetUserID, followerID int, followerName string) {
	content := fmt.Sprintf("%s sent you a follow request", followerName)
	createNotification(targetUserID, "follow_request", content, followerID)
}

// FollowAccepted notifies user their follow request was accepted
func FollowAccepted(requesterID, accepterID int, accepterName string) {
	content := fmt.Sprintf("%s accepted your follow request", accepterName)
	createNotification(requesterID, "follow", content, accepterID)
}

// NewFollower notifies user about a new follower (public profile)
func NewFollower(targetUserID, followerID int, followerName string) {
	content := fmt.Sprintf("%s started following you", followerName)
	createNotification(targetUserID, "follow", content, followerID)
}

// ============================================
// GROUP NOTIFICATIONS
// ============================================

// GroupInvite notifies user about group invitation
func GroupInvite(invitedUserID, groupID int, inviterName, groupName string) {
	content := fmt.Sprintf("%s invited you to join %s", inviterName, groupName)
	createNotification(invitedUserID, "group_invite", content, groupID)
}

// GroupJoinRequest notifies group creator about join request
func GroupJoinRequest(creatorID, groupID int, requesterName, groupName string) {
	content := fmt.Sprintf("%s wants to join your group %s", requesterName, groupName)
	createNotification(creatorID, "group_request", content, groupID)
}

// GroupRequestAccepted notifies user their join request was accepted
func GroupRequestAccepted(requesterID, groupID int, groupName string) {
	content := fmt.Sprintf("Your request to join %s was accepted", groupName)
	createNotification(requesterID, "group_invite", content, groupID)
}

// GroupRequestRejected notifies user their join request was rejected
func GroupRequestRejected(requesterID, groupID int, groupName string) {
	content := fmt.Sprintf("Your request to join %s was declined", groupName)
	createNotification(requesterID, "group_invite", content, groupID)
}

// NewGroupMember notifies creator when someone joins group
func NewGroupMember(creatorID, groupID int, memberName, groupName string) {
	content := fmt.Sprintf("%s joined your group %s", memberName, groupName)
	createNotification(creatorID, "group_invite", content, groupID)
}

// GroupPost notifies members about new group post
func GroupPost(memberIDs []int, postID int, authorName, groupName string) {
	content := fmt.Sprintf("%s posted in %s", authorName, groupName)

	// Send to all members except author
	for _, memberID := range memberIDs {
		createNotification(memberID, "post", content, postID)
	}
}

// ============================================
// EVENT NOTIFICATIONS
// ============================================

// EventCreated notifies group members about new event
func EventCreated(memberIDs []int, eventID int, creatorName, eventTitle, groupName string) {
	content := fmt.Sprintf("%s created event %s in %s", creatorName, eventTitle, groupName)

	for _, memberID := range memberIDs {
		createNotification(memberID, "event", content, eventID)
	}
}

// EventResponse notifies event creator about response
func EventResponse(creatorID, eventID int, responderName, eventTitle, response string) {
	content := fmt.Sprintf("%s is %s to %s", responderName, response, eventTitle)
	createNotification(creatorID, "event", content, eventID)
}

// ============================================
// POST & COMMENT NOTIFICATIONS
// ============================================

// NewComment notifies post author about comment
func NewComment(postAuthorID, commentID int, commenterName, commentPreview string) {
	// Truncate preview if needed
	if len(commentPreview) > 50 {
		commentPreview = commentPreview[:50] + "..."
	}
	content := fmt.Sprintf("%s commented on your post: '%s'", commenterName, commentPreview)
	createNotification(postAuthorID, "comment", content, commentID)
}

// NewPost notifies followers about new post
func NewPost(followerIDs []int, postID int, authorName string) {
	content := fmt.Sprintf("%s shared a new post", authorName)

	for _, followerID := range followerIDs {
		createNotification(followerID, "post", content, postID)
	}
}

// ============================================
// MESSAGE NOTIFICATIONS
// ============================================

// NewMessage notifies user about private message
func NewMessage(receiverID, messageID int, senderName string) {
	content := fmt.Sprintf("New message from %s", senderName)
	createNotification(receiverID, "message", content, messageID)
}

// NewGroupMessage notifies group members about group chat message
func NewGroupMessage(memberIDs []int, messageID, senderID int, senderName, groupName string) {
	content := fmt.Sprintf("%s sent a message in %s", senderName, groupName)

	// Send to all members except sender
	for _, memberID := range memberIDs {
		if memberID != senderID {
			createNotification(memberID, "message", content, messageID)
		}
	}
}
