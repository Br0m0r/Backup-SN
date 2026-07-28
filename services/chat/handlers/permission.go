package handlers

import (
	"context"
	"database/sql"

	"social-network/services/chat/db"
	"social-network/services/chat/usersclient"
)

func canChat(ctx context.Context, database *sql.DB, directory usersclient.Directory, senderID, receiverID int) (bool, error) {
	hasHistory, err := db.HasChatHistory(database, senderID, receiverID)
	if err != nil || hasHistory {
		return hasHistory, err
	}
	return directory.CanStartConversation(ctx, senderID, receiverID)
}
