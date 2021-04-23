package handlers

import (
	eUseCase "github.com/calendar-bot/pkg/events/usecase"
	uUseCase "github.com/calendar-bot/pkg/users/usecase"
	tb "gopkg.in/tucnak/telebot.v2"
)

type BaseHandlers struct {
	eventUseCase eUseCase.EventUseCase
	userUseCase  uUseCase.UserUseCase
	bot *tb.Bot
}

func NewBaseHandlers(eventUC eUseCase.EventUseCase, userUC uUseCase.UserUseCase) BaseHandlers {
	return BaseHandlers{eventUseCase: eventUC, userUseCase: userUC, bot: nil}
}

func (bh *BaseHandlers) InitHandlers(bot *tb.Bot) {
	bh.bot = bot
	bot.Handle("/test", bh.HandleTest)
}

func (bh *BaseHandlers) HandleTest (m *tb.Message) {
	bh.bot.Send(m.Sender, m.Text)
}