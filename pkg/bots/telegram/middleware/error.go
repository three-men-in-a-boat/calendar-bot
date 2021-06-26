package middleware

import (
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/tucnak/telebot.v2"
	"strings"
)

type tgError struct {
	inner         error
	additionalMsg string
	msgID         *int
	chatID        *int64
}

func (te *tgError) Error() string {
	var msg strings.Builder

	// nickeskov: can't fail, see Write method for strings.Builder
	_, _ = fmt.Fprintf(&msg, "TelegramBotError: %s;", te.additionalMsg)

	if te.msgID != nil {
		_, _ = fmt.Fprintf(&msg, "tg_message_id=%d;", *te.msgID)
	}
	if te.chatID != nil {
		_, _ = fmt.Fprintf(&msg, "tg_chat_id=%d;", *te.chatID)
	}
	_, _ = fmt.Fprintf(&msg, "error=%v", te.inner)

	return msg.String()
}

type ErrorHandler struct {
	logger *zap.SugaredLogger
}

func NewErrorHandler(logger *zap.SugaredLogger) ErrorHandler {
	return ErrorHandler{
		logger: logger,
	}
}

func (eh ErrorHandler) HandleMsg(message *telebot.Message, err error) {
	// TODO(nickeskov): add chatID, msgID and eh.logger
	tge := tgError{
		inner:         err,
		additionalMsg: "error from message handler",
		msgID:         &message.ID,
	}
	if message.Chat != nil {
		tge.chatID = &message.Chat.ID
	}
	eh.LogErr(&tge)
}

func (eh ErrorHandler) HandleCallback(callback *telebot.Callback, err error) {
	tge := tgError{
		inner:         err,
		additionalMsg: "error from callback handler",
	}
	msg := callback.Message

	if msg != nil {
		tge.msgID = &msg.ID
		if msg.Chat != nil {
			tge.chatID = &msg.Chat.ID
		}
	}
	eh.LogErr(&tge)
}

func (eh ErrorHandler) LogErr(err error) {
	eh.logger.Error(err)
}

func (eh ErrorHandler) WrapMessageHandler(handler func(*telebot.Message) error) func(*telebot.Message) {
	return func(message *telebot.Message) {
		if err := handler(message); err != nil {
			eh.HandleMsg(message, err)
		}
	}
}

func (eh ErrorHandler) WrapCallbackHandler(handler func(*telebot.Callback) error) func(*telebot.Callback) {
	return func(callback *telebot.Callback) {
		if err := handler(callback); err != nil {
			eh.HandleCallback(callback, err)
		}
	}
}
