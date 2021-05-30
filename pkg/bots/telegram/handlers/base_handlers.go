package handlers

import (
	"github.com/calendar-bot/pkg/bots/telegram/inline_keyboards/baseInlineKeyboards"
	"github.com/calendar-bot/pkg/bots/telegram/keyboards/baseKeyboards"
	"github.com/calendar-bot/pkg/bots/telegram/messages"
	"github.com/calendar-bot/pkg/bots/telegram/messages/baseMessages"
	"github.com/calendar-bot/pkg/customerrors"
	eUseCase "github.com/calendar-bot/pkg/events/usecase"
	uUseCase "github.com/calendar-bot/pkg/users/usecase"
	tb "gopkg.in/tucnak/telebot.v2"
)

type BaseHandlers struct {
	handler      Handler
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
	bot.Handle("/stop", bh.HandleStop)
}

func (bh *BaseHandlers) HandleStart(m *tb.Message) {
	if m.Chat.Type != tb.ChatPrivate {
		_, err := bh.handler.bot.Send(m.Chat, messages.ErrorCommandIsNotAllowedInGroupChat)
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
		return
	}
	isAuth, err := bh.userUseCase.IsUserAuthenticatedByTelegramUserID(int64(m.Sender.ID))
	if err != nil {
		customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		bh.handler.SendError(m.Chat, err)
		return
	}

	if !isAuth {
		link, err := bh.userUseCase.GenOauthLinkForTelegramID(int64(m.Sender.ID))
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
			bh.handler.SendError(m.Chat, err)
			return
		}

		_, err = bh.handler.bot.Send(m.Chat, baseMessages.StartNoRegText(),
			&tb.SendOptions{
				ParseMode: tb.ModeHTML,
				ReplyMarkup: &tb.ReplyMarkup{
					ReplyKeyboardRemove: true,
					InlineKeyboard:      baseInlineKeyboards.StartInlineKeyboard(link),
				},
			})
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}

	} else {
		token, err := bh.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(int64(m.Sender.ID))
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
			bh.handler.SendError(m.Chat, err)
			return
		}

		info, err := bh.userUseCase.GetMailruUserInfo(token)
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
			bh.handler.SendError(m.Chat, err)
			return
		}

		_, err = bh.handler.bot.Send(m.Sender,
			baseMessages.StartRegText(info),
			&tb.SendOptions{
				ParseMode: tb.ModeHTML,
				ReplyMarkup: &tb.ReplyMarkup{
					ReplyKeyboardRemove: true,
				},
			},
		)
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
	}
}

func (bh *BaseHandlers) HandleHelp(m *tb.Message) {
	var replyKeyboard [][]tb.ReplyButton = nil
	if m.Chat.Type == tb.ChatPrivate {
		replyKeyboard = baseKeyboards.HelpCommandKeyboard()
	}
	_, err := bh.handler.bot.Send(m.Chat, baseMessages.HelpInfoText(),

		&tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				OneTimeKeyboard:     true,
				ResizeReplyKeyboard: true,
				ReplyKeyboard:       replyKeyboard,
			},
		})

	if err != nil {
		customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
	}
}

func (bh *BaseHandlers) HandleAbout(m *tb.Message) {
	_, err := bh.handler.bot.Send(m.Chat, baseMessages.AboutText(), &tb.SendOptions{
		ParseMode: tb.ModeHTML,
		ReplyMarkup: &tb.ReplyMarkup{
			ReplyKeyboardRemove: true,
		},
	})

	if err != nil {
		customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
	}
}

func (bh *BaseHandlers) HandleStop(m *tb.Message) {
	if m.Chat.Type != tb.ChatPrivate {
		_, err := bh.handler.bot.Send(m.Chat, messages.ErrorCommandIsNotAllowedInGroupChat)
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
		return
	}
	err := bh.userUseCase.DeleteLocalAuthenticatedUserByTelegramUserID(int64(m.Sender.ID))
	if err != nil {
		_, err = bh.handler.bot.Send(m.Chat, "Вы не авторизованны в боте. Для авторизации воспользуйтесь" +
			" командой /start")
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
	} else {
		_, err = bh.handler.bot.Send(m.Chat, "Вы успешно разлогинились")
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
	}
}
