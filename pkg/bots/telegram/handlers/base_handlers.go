package handlers

import (
	"github.com/calendar-bot/pkg/bots/telegram/inline_keyboards"
	"github.com/calendar-bot/pkg/bots/telegram/text"
	eUseCase "github.com/calendar-bot/pkg/events/usecase"
	uUseCase "github.com/calendar-bot/pkg/users/usecase"
	"go.uber.org/zap"
	tb "gopkg.in/tucnak/telebot.v2"
)

type BaseHandlers struct {
	eventUseCase eUseCase.EventUseCase
	userUseCase  uUseCase.UserUseCase
	bot          *tb.Bot
}

func NewBaseHandlers(eventUC eUseCase.EventUseCase, userUC uUseCase.UserUseCase) BaseHandlers {
	return BaseHandlers{eventUseCase: eventUC, userUseCase: userUC, bot: nil}
}

func (bh *BaseHandlers) InitHandlers(bot *tb.Bot) {
	bh.bot = bot
	bot.Handle("/start", bh.HandleStart)
}

func (bh *BaseHandlers) HandleStart(m *tb.Message) {
	isAuth, err := bh.userUseCase.IsUserAuthenticatedByTelegramUserID(int64(m.Sender.ID))
	if err != nil {
		zap.S().Fatal(err)
	}

	if !isAuth {
		link, err := bh.userUseCase.GenOauthLinkForTelegramID(int64(m.Sender.ID))
		if err != nil {
			zap.S().Fatal(err)
		}

		bh.bot.Send(m.Sender, text.StartNoRegText(), inline_keyboards.Start(link))
	} else {
		bh.bot.Send(m.Sender, text.StartRegText())
	}
}
