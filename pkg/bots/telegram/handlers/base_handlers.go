package handlers

import (
	"github.com/calendar-bot/pkg/bots/telegram/inline_keyboards/baseInlineKeyboards"
	"github.com/calendar-bot/pkg/bots/telegram/messages"
	"github.com/calendar-bot/pkg/bots/telegram/messages/baseMessages"
	eUseCase "github.com/calendar-bot/pkg/events/usecase"
	uUseCase "github.com/calendar-bot/pkg/users/usecase"
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
		bh.bot.Send(m.Sender, messages.MessageUnexpectedError(err.Error()))
	}

	if !isAuth {
		link, err := bh.userUseCase.GenOauthLinkForTelegramID(int64(m.Sender.ID))
		if err != nil {
			bh.bot.Send(m.Sender, messages.MessageUnexpectedError(err.Error()))
			return
		}

		bh.bot.Send(m.Sender, baseMessages.StartNoRegText(), baseInlineKeyboards.Start(link))
	} else {
		token, err := bh.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(int64(m.Sender.ID))
		if err != nil {
			bh.bot.Send(m.Sender, messages.MessageUnexpectedError(err.Error()))
			return
		}

		info, err := bh.userUseCase.GetMailruUserInfo(token)
		if err != nil {
			bh.bot.Send(m.Sender, messages.MessageUnexpectedError(err.Error()))
			return
		}

		bh.bot.Send(m.Sender, baseMessages.StartRegText(info))
	}
}
