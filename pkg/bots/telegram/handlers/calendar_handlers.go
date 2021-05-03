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
	bot.Handle(tb.OnText, ch.HandleText)
}

func (ch *CalendarHandlers) HandleToday(m *tb.Message) {
	if !ch.AuthMiddleware(m.Sender) {
		return
	}
	token, err := ch.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(int64(m.Sender.ID))
	if err != nil {
		customerrors.HandlerError(err)
		ch.handler.SendError(m.Sender, err)
		return
	}

	events, err := ch.eventUseCase.GetEventsToday(token)
	if err != nil {
		customerrors.HandlerError(err)
		ch.handler.SendError(m.Sender, err)
		return
	}

	if events != nil {
		_, err := ch.handler.bot.Send(m.Sender, calendarMessages.GetTodayTitle(), &tb.SendOptions{
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
		_, err := ch.handler.bot.Send(m.Sender, calendarMessages.GetTodayNotFound(), &tb.SendOptions{
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
	if !ch.AuthMiddleware(m.Sender) {
		return
	}
	token, err := ch.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(int64(m.Sender.ID))
	if err != nil {
		customerrors.HandlerError(err)
		ch.handler.SendError(m.Sender, err)
		return
	}

	event, err := ch.eventUseCase.GetClosestEvent(token)
	if err != nil {
		customerrors.HandlerError(err)
		ch.handler.SendError(m.Sender, err)
		return
	}

	if event != nil {
		_, err := ch.handler.bot.Send(m.Sender, calendarMessages.GetNextTitle(), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
			},
		})
		if err != nil {
			customerrors.HandlerError(err)
		}

		inlineKeyboard, err := calendarInlineKeyboards.EventShowMoreInlineKeyboard(event, ch.redisDB)
		if err != nil {
			customerrors.HandlerError(err)
		}
		_, err = ch.handler.bot.Send(m.Sender, calendarMessages.SingleEventShortText(event), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: inlineKeyboard,
			},
		})
		if err != nil {
			customerrors.HandlerError(err)
		}
	} else {
		_, err = ch.handler.bot.Send(m.Sender, calendarMessages.NoClosestEvents(), &tb.SendOptions{
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
	if !ch.AuthMiddleware(m.Sender) {
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
	_, err = ch.handler.bot.Send(m.Sender, calendarMessages.GetInitDateMessage(), &tb.SendOptions{
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
	if !ch.AuthMiddleware(m.Sender) {
		return
	}
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

			_, err = ch.handler.bot.Send(m.Sender, calendarMessages.GetCancelDate(), &tb.SendOptions{
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
			ch.handler.SendError(m.Sender, err)
			return
		}

		client := &http.Client{}
		req, err := http.NewRequest(http.MethodPut, ch.handler.parseAddress+"/parse/date", bytes.NewBuffer(b))
		if err != nil {
			customerrors.HandlerError(err)
			ch.handler.SendError(m.Sender, err)
			return
		}
		req.Header.Add("Content-Type", "application/json")
		resp, err := client.Do(req)

		if err != nil {
			customerrors.HandlerError(err)
			ch.handler.SendError(m.Sender, err)
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
			ch.handler.SendError(m.Sender, err)
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
				ch.handler.SendError(m.Sender, err)
				return
			}
			events, err := ch.eventUseCase.GetEventsByDate(token, parseDate.Date)
			if err != nil {
				customerrors.HandlerError(err)
				ch.handler.SendError(m.Sender, err)
				return
			}
			if events != nil {
				_, err := ch.handler.bot.Send(m.Sender, calendarMessages.GetDateTitle(parseDate.Date), &tb.SendOptions{
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
				_, err := ch.handler.bot.Send(m.Sender, calendarMessages.GetDateEventsNotFound(), &tb.SendOptions{
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
			_, err = ch.handler.bot.Send(m.Sender, calendarMessages.GetDateNotParsed(), &tb.SendOptions{
				ParseMode: tb.ModeHTML,
			})
			if err != nil {
				customerrors.HandlerError(err)
			}
		}

	} else {
		_, err = ch.handler.bot.Send(m.Sender, calendarMessages.RedisSessionNotFound(), &tb.SendOptions{
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
func (ch *CalendarHandlers) HandleShowMore(c *tb.Callback) {
	if !ch.AuthMiddleware(c.Sender) {
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
	if !ch.AuthMiddleware(c.Sender) {
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
		keyboard, err := calendarInlineKeyboards.EventShowMoreInlineKeyboard(&event, ch.redisDB)
		if err != nil {
			zap.S().Errorf("Can't set calendarId=%v for eventId=%v. Err: %v",
				event.Calendar.UID, event.Uid, err)
		}
		if chat.Type != tb.ChatPrivate {
			keyboard = nil
		}
		_, err = ch.handler.bot.Send(user, calendarMessages.SingleEventShortText(&event), &tb.SendOptions{
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
		ch.handler.SendAuthError(c.Sender, err)

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
		ch.handler.SendError(c.Sender, err)

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


func (ch *CalendarHandlers) AuthMiddleware(u *tb.User) bool {
	isAuth, err := ch.userUseCase.IsUserAuthenticatedByTelegramUserID(int64(u.ID))
	if err != nil {
		customerrors.HandlerError(err)
		return false
	}

	if !isAuth {
		_, err = ch.handler.bot.Send(u, calendarMessages.GetUserNotAuth(), &tb.SendOptions{
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
func (ch *CalendarHandlers) GroupMiddleware(c *tb.Chat) bool {
	return true
}
