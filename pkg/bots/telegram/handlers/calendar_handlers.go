package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/calendar-bot/pkg/bots/telegram"
	"github.com/calendar-bot/pkg/bots/telegram/inline_keyboards/calendarInlineKeyboards"
	"github.com/calendar-bot/pkg/bots/telegram/keyboards/calendarKeyboards"
	"github.com/calendar-bot/pkg/bots/telegram/messages/calendarMessages"
	"github.com/calendar-bot/pkg/customerrors"
	eUseCase "github.com/calendar-bot/pkg/events/usecase"
	"github.com/calendar-bot/pkg/types"
	uUseCase "github.com/calendar-bot/pkg/users/usecase"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	tb "gopkg.in/tucnak/telebot.v2"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type CalendarHandlers struct {
	handler      Handler
	eventUseCase eUseCase.EventUseCase
	userUseCase  uUseCase.UserUseCase
	redisDB      *redis.Client
}

func NewCalendarHandlers(eventUC eUseCase.EventUseCase, userUC uUseCase.UserUseCase, redis *redis.Client,
	parseAddress string) CalendarHandlers {
	return CalendarHandlers{eventUseCase: eventUC, userUseCase: userUC,
		handler: Handler{bot: nil, parseAddress: parseAddress}, redisDB: redis}
}

func (ch *CalendarHandlers) InitHandlers(bot *tb.Bot) {
	ch.handler.bot = bot
	bot.Handle("/today", ch.HandleToday)
	bot.Handle("/next", ch.HandleNext)
	bot.Handle("/date", ch.HandleDate)
	bot.Handle("\f"+telegram.ShowFullEvent, ch.HandleShowMore)
	bot.Handle("\f"+telegram.ShowShortEvent, ch.HandleShowLess)
	bot.Handle("\f"+telegram.AlertCallbackYes, ch.HandleAlertYes)
	bot.Handle("\f"+telegram.AlertCallbackNo, ch.HandleAlertNo)
	bot.Handle(tb.OnText, ch.HandleText)
}

func (ch *CalendarHandlers) HandleToday(m *tb.Message) {
	if !ch.AuthMiddleware(m.Sender, m.Chat) {
		return
	}
	if ch.GroupMiddleware(m) {
		return
	}
	token, err := ch.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(int64(m.Sender.ID))
	if err != nil {
		customerrors.HandlerError(err)
		ch.handler.SendError(m.Chat, err)
		return
	}

	events, err := ch.eventUseCase.GetEventsToday(token)
	if err != nil {
		customerrors.HandlerError(err)
		ch.handler.SendError(m.Chat, err)
		return
	}

	if events != nil {
		_, err := ch.handler.bot.Send(m.Chat, calendarMessages.GetTodayTitle(), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
			},
		})
		if err != nil {
			customerrors.HandlerError(err)
		}

		ch.sendShortEvents(&events.Data.Events, m.Chat, m.Chat)
	} else {
		_, err := ch.handler.bot.Send(m.Chat, calendarMessages.GetTodayNotFound(), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
			},
		})
		if err != nil {
			customerrors.HandlerError(err)
		}
	}

}
func (ch *CalendarHandlers) HandleNext(m *tb.Message) {
	if !ch.AuthMiddleware(m.Sender, m.Chat) {
		return
	}
	if ch.GroupMiddleware(m) {
		return
	}
	token, err := ch.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(int64(m.Sender.ID))
	if err != nil {
		customerrors.HandlerError(err)
		ch.handler.SendError(m.Chat, err)
		return
	}

	event, err := ch.eventUseCase.GetClosestEvent(token)
	if err != nil {
		customerrors.HandlerError(err)
		ch.handler.SendError(m.Chat, err)
		return
	}

	if event != nil {
		_, err := ch.handler.bot.Send(m.Chat, calendarMessages.GetNextTitle(), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
			},
		})
		if err != nil {
			customerrors.HandlerError(err)
		}

		var inlineKeyboard [][]tb.InlineButton = nil
		if m.Chat.Type == tb.ChatPrivate {
			inlineKeyboard, err = calendarInlineKeyboards.EventShowMoreInlineKeyboard(event, ch.redisDB)
			if err != nil {
				customerrors.HandlerError(err)
			}
		}
		_, err = ch.handler.bot.Send(m.Chat, calendarMessages.SingleEventShortText(event), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: inlineKeyboard,
			},
		})
		if err != nil {
			customerrors.HandlerError(err)
		}
	} else {
		_, err = ch.handler.bot.Send(m.Chat, calendarMessages.NoClosestEvents(), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
			},
		})
		if err != nil {
			customerrors.HandlerError(err)
		}
	}
}
func (ch *CalendarHandlers) HandleDate(m *tb.Message) {
	if !ch.AuthMiddleware(m.Sender, m.Chat) {
		return
	}
	if ch.GroupMiddleware(m) {
		return
	}
	currSession, err := ch.getSession(m.Sender)
	if err != nil {
		return
	}

	currSession.IsDate = true
	err = ch.setSession(currSession, m.Sender)
	if err != nil {
		return
	}
	_, err = ch.handler.bot.Send(m.Chat, calendarMessages.GetInitDateMessage(), &tb.SendOptions{
		ParseMode: tb.ModeHTML,
		ReplyMarkup: &tb.ReplyMarkup{
			ReplyKeyboard: calendarKeyboards.GetDateFastCommand(),
		},
	})
	if err != nil {
		customerrors.HandlerError(err)
	}
}
func (ch *CalendarHandlers) HandleText(m *tb.Message) {
	session, err := ch.getSession(m.Sender)
	if err != nil {
		return
	}

	if session.IsDate {
		if calendarMessages.GetCancelDateReplyButton() == m.Text {
			session.IsDate = false
			err := ch.setSession(session, m.Sender)
			if err != nil {
				return
			}

			_, err = ch.handler.bot.Send(m.Chat, calendarMessages.GetCancelDate(), &tb.SendOptions{
				ParseMode: tb.ModeHTML,
				ReplyMarkup: &tb.ReplyMarkup{
					ReplyKeyboardRemove: true,
				},
			})

			if err != nil {
				customerrors.HandlerError(err)
			}

			return
		}

		reqData := &types.ParseDateReq{Timezone: "Europe/Moscow", Text: m.Text}
		b, err := json.Marshal(reqData)
		if err != nil {
			customerrors.HandlerError(err)
			ch.handler.SendError(m.Chat, err)
			return
		}

		client := &http.Client{}
		req, err := http.NewRequest(http.MethodPut, ch.handler.parseAddress+"/parse/date", bytes.NewBuffer(b))
		if err != nil {
			customerrors.HandlerError(err)
			ch.handler.SendError(m.Chat, err)
			return
		}
		req.Header.Add("Content-Type", "application/json")
		resp, err := client.Do(req)

		if err != nil {
			customerrors.HandlerError(err)
			ch.handler.SendError(m.Chat, err)
			return
		}

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				customerrors.HandlerError(err)
			}
		}(resp.Body)

		body, err := ioutil.ReadAll(resp.Body)

		parseDate := &types.ParseDateResp{}
		err = json.Unmarshal(body, parseDate)
		if err != nil {
			customerrors.HandlerError(err)
			ch.handler.SendError(m.Chat, err)
			return
		}

		if !parseDate.Date.IsZero() {
			session.IsDate = false
			err := ch.setSession(session, m.Sender)
			if err != nil {
				return
			}

			token, err := ch.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(int64(m.Sender.ID))
			if err != nil {
				customerrors.HandlerError(err)
				ch.handler.SendError(m.Chat, err)
				return
			}
			events, err := ch.eventUseCase.GetEventsByDate(token, parseDate.Date)
			if err != nil {
				customerrors.HandlerError(err)
				ch.handler.SendError(m.Chat, err)
				return
			}
			if events != nil {
				_, err := ch.handler.bot.Send(m.Chat, calendarMessages.GetDateTitle(parseDate.Date), &tb.SendOptions{
					ParseMode: tb.ModeHTML,
					ReplyMarkup: &tb.ReplyMarkup{
						ReplyKeyboardRemove: true,
					},
				})
				if err != nil {
					customerrors.HandlerError(err)
				}
				ch.sendShortEvents(&events.Data.Events, m.Sender, m.Chat)
			} else {
				_, err := ch.handler.bot.Send(m.Chat, calendarMessages.GetDateEventsNotFound(), &tb.SendOptions{
					ParseMode: tb.ModeHTML,
					ReplyMarkup: &tb.ReplyMarkup{
						ReplyKeyboardRemove: true,
					},
				})

				if err != nil {
					customerrors.HandlerError(err)
				}
			}

		} else {
			_, err = ch.handler.bot.Send(m.Chat, calendarMessages.GetDateNotParsed(), &tb.SendOptions{
				ParseMode: tb.ModeHTML,
			})
			if err != nil {
				customerrors.HandlerError(err)
			}
		}

	} else {
		if m.Chat.Type == tb.ChatPrivate {
			_, err = ch.handler.bot.Send(m.Chat, calendarMessages.RedisSessionNotFound(), &tb.SendOptions{
				ParseMode: tb.ModeHTML,
				ReplyMarkup: &tb.ReplyMarkup{
					ReplyKeyboardRemove: true,
				},
			})
			if err != nil {
				customerrors.HandlerError(err)
			}
		}
	}
}
func (ch *CalendarHandlers) HandleShowMore(c *tb.Callback) {
	if !ch.AuthMiddleware(c.Sender, c.Message.Chat) {
		err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
		})
		if err != nil {
			customerrors.HandlerError(err)
		}
		return
	}
	event := ch.getEventByIdForCallback(c)
	if event == nil {
		return
	}

	err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
		CallbackID: c.ID,
		Text:       calendarMessages.CallbackResponseHeader(event),
	})
	if err != nil {
		customerrors.HandlerError(err)
	}

	_, err = ch.handler.bot.Edit(c.Message, calendarMessages.SingleEventFullText(event),
		&tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: calendarInlineKeyboards.EventShowLessInlineKeyboard(event),
			},
		})
	if err != nil {
		customerrors.HandlerError(err)
	}
}
func (ch *CalendarHandlers) HandleShowLess(c *tb.Callback) {
	if !ch.AuthMiddleware(c.Sender, c.Message.Chat) {
		err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
		})
		if err != nil {
			customerrors.HandlerError(err)
		}
		return
	}
	event := ch.getEventByIdForCallback(c)
	if event == nil {
		return
	}

	err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
		CallbackID: c.ID,
	})
	if err != nil {
		customerrors.HandlerError(err)
	}

	inlineKeyboard, err := calendarInlineKeyboards.EventShowMoreInlineKeyboard(event, ch.redisDB)
	if err != nil {
		customerrors.HandlerError(err)
	}
	_, err = ch.handler.bot.Edit(c.Message, calendarMessages.SingleEventShortText(event),
		&tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: inlineKeyboard,
			},
		})
	if err != nil {
		customerrors.HandlerError(err)
	}
}
func (ch *CalendarHandlers) HandleAlertYes(c *tb.Callback) {
	err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
		CallbackID: c.ID,
	})
	if err != nil {
		customerrors.HandlerError(err)
	}

	c.Message.Sender = c.Message.ReplyTo.Sender
	c.Message.ReplyTo = nil

	err = ch.handler.bot.Delete(c.Message)
	if err != nil {
		customerrors.HandlerError(err)
	}

	switch c.Data {
	case telegram.Today:
		ch.HandleToday(c.Message)
		break
	case telegram.Next:
		ch.HandleNext(c.Message)
		break
	case telegram.Date:
		ch.HandleDate(c.Message)
		break
	}

}
func (ch *CalendarHandlers) HandleAlertNo(c *tb.Callback) {
	err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
		CallbackID: c.ID,
	})
	if err != nil {
		customerrors.HandlerError(err)
	}
	err = ch.handler.bot.Delete(c.Message)
	if err != nil {
		customerrors.HandlerError(err)
	}
}

func (ch *CalendarHandlers) getSession(user *tb.User) (*types.BotRedisSession, error) {
	resp, err := ch.redisDB.Get(context.TODO(), strconv.Itoa(user.ID)).Result()
	if err != nil {
		newSession := &types.BotRedisSession{
			IsDate: false,
		}
		err = ch.setSession(newSession, user)

		if err != nil {
			customerrors.HandlerError(err)
			_, err := ch.handler.bot.Send(user, calendarMessages.RedisSessionNotFound(), &tb.SendOptions{
				ParseMode: tb.ModeHTML,
				ReplyMarkup: &tb.ReplyMarkup{
					ReplyKeyboardRemove: true,
				},
			})
			if err != nil {
				customerrors.HandlerError(err)
			}
			return nil, err
		}

		return newSession, nil
	}

	session := &types.BotRedisSession{}
	err = json.Unmarshal([]byte(resp), session)
	if err != nil {
		customerrors.HandlerError(err)
		_, err := ch.handler.bot.Send(user, calendarMessages.RedisSessionNotFound(), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
			},
		})
		if err != nil {
			customerrors.HandlerError(err)
		}
		return nil, err
	}

	return session, nil
}
func (ch *CalendarHandlers) setSession(session *types.BotRedisSession, user *tb.User) error {
	b, err := json.Marshal(session)
	if err != nil {
		customerrors.HandlerError(err)
		ch.handler.SendError(user, err)
		return err
	}
	err = ch.redisDB.Set(context.TODO(), strconv.Itoa(user.ID), b, 0).Err()
	if err != nil {
		customerrors.HandlerError(err)
		ch.handler.SendError(user, err)
		return err
	}

	return nil
}
func (ch *CalendarHandlers) sendShortEvents(events *types.Events, user tb.Recipient, chat *tb.Chat) {
	for _, event := range *events {
		var err error = nil
		var keyboard [][]tb.InlineButton = nil
		if chat.Type == tb.ChatPrivate {
			keyboard, err = calendarInlineKeyboards.EventShowMoreInlineKeyboard(&event, ch.redisDB)
			if err != nil {
				zap.S().Errorf("Can't set calendarId=%v for eventId=%v. Err: %v",
					event.Calendar.UID, event.Uid, err)
			}
		}
		if chat.Type != tb.ChatPrivate {
			keyboard = nil
		}
		_, err = ch.handler.bot.Send(chat, calendarMessages.SingleEventShortText(&event), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: keyboard,
			},
		})
		if err != nil {
			customerrors.HandlerError(err)
		}
	}
}
func (ch *CalendarHandlers) getEventByIdForCallback(c *tb.Callback) *types.Event {
	calUid, err := ch.redisDB.Get(context.TODO(), c.Data).Result()
	if err != nil {
		customerrors.HandlerError(err)
		err = ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
			Text:       calendarMessages.RedisNotFoundMessage(),
			ShowAlert:  true,
		})
		if err != nil {
			customerrors.HandlerError(err)
		}
		return nil
	}

	token, err := ch.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(int64(c.Sender.ID))

	if err != nil {
		customerrors.HandlerError(err)
		ch.handler.SendAuthError(c.Message.Chat, err)

		err = ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
		})
		if err != nil {
			customerrors.HandlerError(err)
		}
		return nil
	}
	resp, err := ch.eventUseCase.GetEventByEventID(token, calUid, c.Data)
	if err != nil {
		customerrors.HandlerError(err)
		ch.handler.SendError(c.Message.Chat, err)

		err = ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
		})
		if err != nil {
			customerrors.HandlerError(err)
		}
		return nil
	}

	return &resp.Data.Event
}

func (ch *CalendarHandlers) AuthMiddleware(u *tb.User, c *tb.Chat) bool {
	isAuth, err := ch.userUseCase.IsUserAuthenticatedByTelegramUserID(int64(u.ID))
	if err != nil {
		customerrors.HandlerError(err)
		return false
	}

	if !isAuth {
		_, err = ch.handler.bot.Send(c, calendarMessages.GetUserNotAuth(), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
			},
		})
		if err != nil {
			customerrors.HandlerError(err)
		}
	}

	return isAuth
}
func (ch *CalendarHandlers) GroupMiddleware(m *tb.Message) bool {
	if strings.Contains(m.Text, calendarMessages.GetMessageAlertBase()) {
		return false
	}
	if m.Chat.Type != tb.ChatPrivate {
		_, err := ch.handler.bot.Send(m.Chat, calendarMessages.GetGroupAlertMessage(m.Text), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyTo:   m,
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: calendarInlineKeyboards.GroupAlertsButtons(m.Text),
			},
		})
		if err != nil {
			customerrors.HandlerError(err)
		}
		return true
	}

	return false
}
