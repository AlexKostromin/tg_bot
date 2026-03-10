package middleware

func IsAdmin(chatID int64, adminChatID int64) bool {
	return chatID == adminChatID
}
