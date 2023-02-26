package discord

type updateMessage struct {
	roomID    uint64
	sessionID string
	messageID string
	message   *WebhookMessage
}
