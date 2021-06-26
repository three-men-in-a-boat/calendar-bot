package middleware

import (
	"github.com/calendar-bot/pkg/bots/telegram"
	"gopkg.in/tucnak/telebot.v2"
)

func NewMessageHandlerErrorsMetricMiddleware(handler func(*telebot.Message) error) func(*telebot.Message) error {
	return func(message *telebot.Message) error {
		if err := handler(message); err != nil {
			telegram.MetricTotalErrorsCount.Inc()
			return err
		}
		return nil
	}
}

func NewCallbackHandlerErrorsMetricMiddleware(handler func(*telebot.Callback) error) func(*telebot.Callback) error {
	return func(callback *telebot.Callback) error {
		if err := handler(callback); err != nil {
			telegram.MetricTotalErrorsCount.Inc()
			return err
		}
		return nil
	}
}
