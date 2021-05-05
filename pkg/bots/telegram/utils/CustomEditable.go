package utils

type CustomEditable struct {
	MessageID string
	ChatID int64
}

func (ce *CustomEditable) MessageSig() (messageID string, chatID int64) {
	return ce.MessageID, ce.ChatID
}

func InitCustomEditable(messageID string, chatID int64) CustomEditable {
	return CustomEditable{
		MessageID: messageID,
		ChatID:    chatID,
	}
}