package handlers

import (
	"github.com/calendar-bot/pkg/bots/telegram/inline_keyboards/baseInlineKeyboards"
	"github.com/calendar-bot/pkg/bots/telegram/keyboards/baseKeyboards"
	"github.com/calendar-bot/pkg/bots/telegram/messages/baseMessages"
	"github.com/calendar-bot/pkg/customerrors"
	eUseCase "github.com/calendar-bot/pkg/events/usecase"
	uUseCase "github.com/calendar-bot/pkg/users/usecase"
	tb "gopkg.in/tucnak/telebot.v2"
)

type BaseHandlers struct {
	handler Handler
	eventUseCase eUseCase.EventUseCase
	userUseCase  uUseCase.UserUseCase
}

func NewBaseHandlers(eventUC eUseCase.EventUseCase, userUC uUseCase.UserUseCase, parseAddress string) BaseHandlers {
	return BaseHandlers{eventUseCase: eventUC, userUseCase: userUC,
		handler: Handler{bot: nil, parseAddress: parseAddress}}
}

func (bh *BaseHandlers) InitHandlers(bot *tb.Bot) {
	bh.handler.bot = bot
	bot.Handle("/start", bh.HandleStart)
	bot.Handle("/help", bh.HandleHelp)
	bot.Handle("/about", bh.HandleAbout)
}

func (bh *BaseHandlers) HandleStart(m *tb.Message) {
	isAuth, err := bh.userUseCase.IsUserAuthenticatedByTelegramUserID(int64(m.Sender.ID))
	if err != nil {
		customerrors.HandlerError(err)
		bh.handler.SendError(m.Sender, err)
		return
	}

	if !isAuth {
		link, err := bh.userUseCase.GenOauthLinkForTelegramID(int64(m.Sender.ID))
		if err != nil {
			customerrors.HandlerError(err)
			bh.handler.SendError(m.Sender, err)
			return
		}

		_, err = bh.handler.bot.Send(m.Sender, baseMessages.StartNoRegText(),
			&tb.SendOptions{
				ParseMode: tb.ModeHTML,
				ReplyMarkup: &tb.ReplyMarkup{
					ReplyKeyboardRemove: true,
					InlineKeyboard:      baseInlineKeyboards.StartInlineKeyboard(link),
				},
			})
		if err != nil {
			customerrors.HandlerError(err)
		}

	} else {
		token, err := bh.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(int64(m.Sender.ID))
		if err != nil {
			customerrors.HandlerError(err)
			bh.handler.SendError(m.Sender, err)
			return
		}

		info, err := bh.userUseCase.GetMailruUserInfo(token)
		if err != nil {
			customerrors.HandlerError(err)
			bh.handler.SendError(m.Sender, err)
			return
		}

		_ ,err = bh.handler.bot.Send(m.Sender,
			baseMessages.StartRegText(info),
			&tb.SendOptions{
				ParseMode: tb.ModeHTML,
				ReplyMarkup: &tb.ReplyMarkup{
					ReplyKeyboardRemove: true,
				},
			},
		)
		if err != nil {
			customerrors.HandlerError(err)
		}
	}
}

func (bh *BaseHandlers) HandleHelp(m *tb.Message) {
	_, err := bh.handler.bot.Send(m.Sender, baseMessages.HelpInfoText(),
		&tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				OneTimeKeyboard:     true,
				ResizeReplyKeyboard: true,
				ReplyKeyboard:       baseKeyboards.HelpCommandKeyboard(),
			},
		})

	if err != nil {
		customerrors.HandlerError(err)
	}
}

func (bh *BaseHandlers) HandleAbout(m *tb.Message) {
	_, err := bh.handler.bot.Send(m.Sender, baseMessages.AboutText(), &tb.SendOptions{
		ParseMode: tb.ModeHTML,
		ReplyMarkup: &tb.ReplyMarkup{
			ReplyKeyboardRemove: true,
		},
	})

	if err != nil {
		customerrors.HandlerError(err)
	}
}
