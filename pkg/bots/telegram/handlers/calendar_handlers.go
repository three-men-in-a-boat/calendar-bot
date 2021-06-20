package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/calendar-bot/pkg/bots/telegram"
	"github.com/calendar-bot/pkg/bots/telegram/inline_keyboards/calendarInlineKeyboards"
	"github.com/calendar-bot/pkg/bots/telegram/keyboards/calendarKeyboards"
	"github.com/calendar-bot/pkg/bots/telegram/messages/calendarMessages"
	"github.com/calendar-bot/pkg/bots/telegram/utils"
	"github.com/calendar-bot/pkg/customerrors"
	eUseCase "github.com/calendar-bot/pkg/events/usecase"
	"github.com/calendar-bot/pkg/types"
	uUseCase "github.com/calendar-bot/pkg/users/usecase"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	tb "gopkg.in/tucnak/telebot.v2"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
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
	bot.Handle("/create", ch.HandleCreate)

	bot.Handle(calendarMessages.CreateEventAddTitleButton, ch.HandleTitleChange)
	bot.Handle(calendarMessages.CreateEventChangeTitleButton, ch.HandleTitleChange)
	bot.Handle(calendarMessages.CreateEventAddDescButton, ch.HandleDescChange)
	bot.Handle(calendarMessages.CreateEventChangeDescButton, ch.HandleDescChange)
	bot.Handle(calendarMessages.CreateEventAddLocationButton, ch.HandleLocationChange)
	bot.Handle(calendarMessages.CreateEventChangeLocationButton, ch.HandleLocationChange)
	bot.Handle(calendarMessages.CreateEventChangeStopTimeButton, ch.HandleStopTimeChange)
	bot.Handle(calendarMessages.CreateEventChangeStartTimeButton, ch.HandleStartTimeChange)
	bot.Handle(calendarMessages.CreateEventAddUser, ch.HandleUserChange)
	bot.Handle(calendarMessages.GetCreateFullDay(), ch.HandleFullDayChange)

	bot.Handle("\f"+telegram.ShowFullEvent, ch.HandleShowMore)
	bot.Handle("\f"+telegram.ShowShortEvent, ch.HandleShowLess)
	bot.Handle("\f"+telegram.AlertCallbackYes, ch.HandleAlertYes)
	bot.Handle("\f"+telegram.AlertCallbackNo, ch.HandleAlertNo)
	bot.Handle("\f"+telegram.CancelCreateEvent, ch.HandleCancelCreateEvent)
	bot.Handle("\f"+telegram.CreateEvent, ch.HandleCreateEvent)
	bot.Handle("\f"+telegram.GroupGo, ch.HandleGroupGo)
	bot.Handle("\f"+telegram.GroupNotGo, ch.HandleGroupNotGo)
	bot.Handle("\f"+telegram.GroupFindTimeNo, ch.HandleGroupFindTimeNo)
	bot.Handle("\f"+telegram.GroupFindTimeYes, ch.HandleGroupFindTimeYes)
	bot.Handle("\f"+telegram.FindTimeDayPart, ch.HandleFindTimeDayPart)
	bot.Handle("\f"+telegram.FindTimeLength, ch.HandleFindTimeLength)
	bot.Handle("\f"+telegram.FindTimeAdd, ch.FindTimeAdd)
	bot.Handle("\f"+telegram.FindTimeCreate, ch.FindTimeCreate)
	bot.Handle("\f"+telegram.HandleGroupText, ch.HandleGroupText)
	bot.Handle("\f"+telegram.FindTimeFind, ch.HandleFindTimeFind)
	bot.Handle("\f"+telegram.FindTimeBack, ch.HandleFindTimeBack)
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
		customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		ch.handler.SendError(m.Chat, err)
		return
	}

	events, err := ch.eventUseCase.GetEventsToday(token)
	if err != nil {
		customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		ch.handler.SendError(m.Chat, err)
		return
	}

	if events != nil {
		i := 0
		for _, event := range events.Data.Events {
			if !event.FullDay || event.From.Day() != time.Now().Day()-1 || event.To.Day() != time.Now().Day() {
				events.Data.Events[i] = event
				i++
			}
		}
		events.Data.Events = events.Data.Events[:i]
	}

	title := calendarMessages.GetTodayTitle()
	if m.Chat.Type != tb.ChatPrivate {
		title += calendarMessages.AddNameBold(m.Sender.FirstName + " " + m.Sender.LastName)
	}

	if events != nil || len(events.Data.Events) > 0 {
		_, err := ch.handler.bot.Send(m.Chat, title, &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
			},
		})
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}

		ch.sendShortEvents(&events.Data.Events, m.Chat)
	} else {
		title := calendarMessages.GetTodayNotFound()
		if m.Chat.Type != tb.ChatPrivate {
			title += calendarMessages.AddName(m.Sender.FirstName + " " + m.Sender.LastName)
		}
		_, err := ch.handler.bot.Send(m.Chat, title, &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
			},
		})
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
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
		customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		ch.handler.SendError(m.Chat, err)
		return
	}

	event, err := ch.eventUseCase.GetClosestEvent(token)
	if err != nil {
		customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		ch.handler.SendError(m.Chat, err)
		return
	}

	if event != nil {
		title := calendarMessages.GetNextTitle()
		if m.Chat.Type != tb.ChatPrivate {
			title += calendarMessages.AddNameBold(m.Sender.FirstName + " " + m.Sender.LastName)
		}
		_, err := ch.handler.bot.Send(m.Chat, title, &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
			},
		})
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}

		var inlineKeyboard [][]tb.InlineButton = nil
		if m.Chat.Type == tb.ChatPrivate {
			inlineKeyboard, err = calendarInlineKeyboards.EventShowMoreInlineKeyboard(event, ch.redisDB)
			if err != nil {
				customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
			}
		}
		_, err = ch.handler.bot.Send(m.Chat, calendarMessages.SingleEventShortText(event), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: inlineKeyboard,
			},
		})
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
	} else {
		title := calendarMessages.NoClosestEvents()
		if m.Chat.Type != tb.ChatPrivate {
			title += calendarMessages.AddName(m.Sender.FirstName + " " + m.Sender.LastName)
		}
		_, err = ch.handler.bot.Send(m.Chat, title, &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
			},
		})
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
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
	currSession, err := ch.getSession(m.Sender, m.Chat)
	if err != nil {
		return
	}

	if currSession.PollMsg.ChatID != 0 {
		err = ch.handler.bot.Delete(&currSession.PollMsg)
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
	}

	if currSession.InlineMsg.ChatID != 0 {
		_, err = ch.handler.bot.EditReplyMarkup(&currSession.InlineMsg, nil)
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
	}

	if currSession.InfoMsg.ChatID != 0 {
		err = ch.handler.bot.Delete(&currSession.InfoMsg)
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
	}

	currSession = &types.BotRedisSession{}
	currSession.IsDate = true
	currSession.IsCreate = false

	replyMarkup := tb.ReplyMarkup{}
	var replyTo *tb.Message = nil
	if m.Chat.Type != tb.ChatPrivate {
		replyMarkup = tb.ReplyMarkup{
			InlineKeyboard: calendarInlineKeyboards.GetDateFastCommand(false),
		}
		replyTo = m.ReplyTo
	} else {
		replyMarkup = tb.ReplyMarkup{
			ReplyKeyboard: calendarKeyboards.GetDateFastCommand(false),
		}
	}

	msg, err := ch.handler.bot.Send(m.Chat, calendarMessages.GetInitDateMessage(), &tb.SendOptions{
		ParseMode:   tb.ModeHTML,
		ReplyMarkup: &replyMarkup,
		ReplyTo:     replyTo,
	})
	if err != nil {
		customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
	}
	if m.Chat.Type != tb.ChatPrivate {
		currSession.InlineMsg = utils.InitCustomEditable(msg.MessageSig())
	}
	err = ch.setSession(currSession, m.Sender, m.Chat)
	if err != nil {
		return
	}
}
func (ch *CalendarHandlers) HandleCreate(m *tb.Message) {
	if !ch.AuthMiddleware(m.Sender, m.Chat) {
		return
	}
	session, err := ch.getSession(m.Sender, m.Chat)
	if err != nil {
		customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		ch.handler.SendError(m.Chat, err)
		return
	}

	if session.PollMsg.ChatID != 0 {
		session.FindTimeDone = false
		err = ch.handler.bot.Delete(&session.PollMsg)
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
	}

	if session.InlineMsg.ChatID != 0 {
		_, err = ch.handler.bot.EditReplyMarkup(&session.InlineMsg, nil)
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
	}

	if session.InfoMsg.ChatID != 0 {
		session.FindTimeDone = false
		err = ch.handler.bot.Delete(&session.InfoMsg)
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
	}

	if (m.Chat.Type == tb.ChatGroup || m.Chat.Type == tb.ChatSuperGroup) && !session.FindTimeDone {
		session = &types.BotRedisSession{}
		err = ch.setSession(session, m.Sender, m.Chat)
		if err != nil {
			ch.handler.SendError(m.Chat, err)
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
			return
		}

		_, err = ch.handler.bot.Send(m.Chat, calendarMessages.CreateEventFindTimeMessage, &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: calendarInlineKeyboards.GroupFindTimeButtons(),
			},
			ReplyTo: m,
		})
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		} else {
			return
		}
	}

	session = &types.BotRedisSession{}
	session.IsCreate = true
	session.FindTimeDone = true

	token, err := ch.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(int64(m.Sender.ID))
	if err != nil {
		customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		ch.handler.SendError(m.Chat, err)
		return
	} else {
		userInfo, err := ch.userUseCase.GetMailruUserInfo(token)
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		} else {
			organizerAttendee := types.AttendeeEvent{
				Email:  userInfo.Email,
				Name:   userInfo.Name,
				Role:   telegram.RoleRequired,
				Status: telegram.StatusAccepted,
			}
			session.Event.Organizer = organizerAttendee
			session.Event.Attendees = append(session.Event.Attendees, organizerAttendee)
		}
	}

	session.Step = telegram.StepCreateFrom

	replyMarkup := tb.ReplyMarkup{}
	var replyTo *tb.Message = nil
	if m.Chat.Type != tb.ChatPrivate {
		replyMarkup = tb.ReplyMarkup{
			InlineKeyboard: calendarInlineKeyboards.GetCreateFastCommand(),
		}
		replyTo = m
	} else {
		replyMarkup = tb.ReplyMarkup{
			ReplyKeyboard: calendarKeyboards.GetCreateFastCommand(),
		}
	}

	msg, err := ch.handler.bot.Send(m.Chat, calendarMessages.GetCreateInitText(), &tb.SendOptions{
		ParseMode:   tb.ModeHTML,
		ReplyMarkup: &replyMarkup,
		ReplyTo:     replyTo,
	})
	if err != nil {
		customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
	}
	if m.Chat.Type != tb.ChatPrivate {
		session.InlineMsg = utils.InitCustomEditable(msg.MessageSig())
	}

	err = ch.setSession(session, m.Sender, m.Chat)
	if err != nil {
		customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		ch.handler.SendError(m.Chat, err)
		return
	}

}
func (ch *CalendarHandlers) HandleText(m *tb.Message) {

	if strings.ToLower(m.Text) == calendarMessages.ShowTodayTasks ||
		strings.ToLower(m.Text) == calendarMessages.ShowTodayPhrase {
		ch.HandleToday(m)
		return
	}

	if strings.ToLower(m.Text) == calendarMessages.ShowNextTask ||
		strings.ToLower(m.Text) == calendarMessages.ShowNextPhrase {
		ch.HandleNext(m)
		return
	}

	session, err := ch.getSession(m.Sender, m.Chat)
	if err != nil {
		return
	}

	if session.IsDate {
		// Удаляет текст Сегодня, Завтра из даты
		m.Text = strings.Split(m.Text, ",")[0]
		ch.handleDateText(m, session)
	} else if session.IsCreate && session.FindTimeDone {
		ch.handleCreateText(m, session)
	} else if session.IsCreate && !session.FindTimeDone {
		ch.handleFindTimeText(m, session)
	} else {
		if m.Chat.Type == tb.ChatPrivate {

			data := ch.ParseEvent(m)

			if data == nil || data.EventStart.IsZero() {
				_, err = ch.handler.bot.Send(m.Chat, calendarMessages.EventNoEventDataFound, &tb.SendOptions{
					ParseMode: tb.ModeHTML,
					ReplyMarkup: &tb.ReplyMarkup{
						ReplyKeyboardRemove: true,
					},
				})
				if err != nil {
					customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
				}

				return
			}

			session.IsCreate = true
			session.FromTextCreate = true
			session.FindTimeDone = true
			session.Event = types.Event{}
			session.Step = telegram.StepCreateInit
			session.Event.From = data.EventStart
			session.Event.To = data.EventEnd
			session.Event.Title = data.EventName

			token, err := ch.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(int64(m.Sender.ID))
			if err != nil {
				customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
				ch.handler.SendError(m.Chat, err)
				return
			} else {
				userInfo, err := ch.userUseCase.GetMailruUserInfo(token)
				if err != nil {
					customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
				} else {
					organizerAttendee := types.AttendeeEvent{
						Email:  userInfo.Email,
						Name:   userInfo.Name,
						Role:   telegram.RoleRequired,
						Status: telegram.StatusAccepted,
					}
					session.Event.Organizer = organizerAttendee
					session.Event.Attendees = append(session.Event.Attendees, organizerAttendee)
				}
			}

			newMsg, err := ch.handler.bot.Send(m.Chat,
				calendarMessages.GetCreateEventHeader()+calendarMessages.SingleEventFullText(&session.Event),
				&tb.SendOptions{
					ParseMode: tb.ModeHTML,
					ReplyTo:   m,
					ReplyMarkup: &tb.ReplyMarkup{
						InlineKeyboard: calendarInlineKeyboards.CreateEventButtons(session.Event),
					},
				})

			if err != nil {
				customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
			}

			session.InfoMsg = utils.InitCustomEditable(newMsg.MessageSig())

			if data.EventEnd.IsZero() {
				session.Step = telegram.StepCreateTo

				_, err = ch.handler.bot.Send(m.Chat, calendarMessages.GetCreateEventToText(), &tb.SendOptions{
					ParseMode: tb.ModeHTML,
					ReplyMarkup: &tb.ReplyMarkup{
						ReplyKeyboard:       calendarKeyboards.GetCreateDuration(),
						ResizeReplyKeyboard: true,
					},
				})

				if err != nil {
					customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
				}

			} else {
				if data.EventName == "" {
					session.Step = telegram.StepCreateTitle

					_, err = ch.handler.bot.Send(m.Chat, calendarMessages.GetCreateEventTitle(), &tb.SendOptions{
						ParseMode: tb.ModeHTML,
						ReplyMarkup: &tb.ReplyMarkup{
							ReplyKeyboard:   calendarKeyboards.GetCreateOptionButtons(session),
							OneTimeKeyboard: true,
						},
					})

					if err != nil {
						customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
					}
				} else {
					session.Step = telegram.StepCreateDesc

					_, err = ch.handler.bot.Send(m.Chat, calendarMessages.CreateEventDescText, &tb.SendOptions{
						ParseMode: tb.ModeHTML,
						ReplyMarkup: &tb.ReplyMarkup{
							ReplyKeyboard:   calendarKeyboards.GetCreateOptionButtons(session),
							OneTimeKeyboard: true,
						},
					})

					if err != nil {
						customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
					}
				}
			}

			err = ch.setSession(session, m.Sender, m.Chat)
			if err == nil {
				return
			}

		}
	}
}

func (ch *CalendarHandlers) HandleDescChange(m *tb.Message) {
	session, err := ch.getSession(m.Sender, m.Chat)
	if err != nil {
		return
	}

	if session.IsCreate {
		session.Step = telegram.StepCreateDesc
		var replyMarkup *tb.ReplyMarkup = nil
		var replyTo *tb.Message = nil
		if m.Chat.Type != tb.ChatPrivate {
			replyMarkup = &tb.ReplyMarkup{
				InlineKeyboard: calendarInlineKeyboards.GetCreateOptionButtons(session),
			}
			replyTo = m
		}
		msg, err := ch.handler.bot.Send(m.Chat, calendarMessages.CreateEventDescText, &tb.SendOptions{
			ParseMode:   tb.ModeHTML,
			ReplyMarkup: replyMarkup,
			ReplyTo:     replyTo,
		})

		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}

		if m.Chat.Type != tb.ChatPrivate {
			session.InlineMsg = utils.InitCustomEditable(msg.MessageSig())
		}

		err = ch.setSession(session, m.Sender, m.Chat)
		if err != nil {
			ch.handler.SendError(m.Chat, err)
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
	} else {
		ch.HandleText(m)
	}
}
func (ch *CalendarHandlers) HandleTitleChange(m *tb.Message) {
	session, err := ch.getSession(m.Sender, m.Chat)
	if err != nil {
		return
	}

	if session.IsCreate {
		session.Step = telegram.StepCreateTitle
		var replyMarkup *tb.ReplyMarkup = nil
		var replyTo *tb.Message = nil
		if m.Chat.Type != tb.ChatPrivate {
			replyMarkup = &tb.ReplyMarkup{
				InlineKeyboard: calendarInlineKeyboards.GetCreateOptionButtons(session),
			}
			replyTo = m
		}
		msg, err := ch.handler.bot.Send(m.Chat, calendarMessages.GetCreateEventTitle(), &tb.SendOptions{
			ParseMode:   tb.ModeHTML,
			ReplyMarkup: replyMarkup,
			ReplyTo:     replyTo,
		})

		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}

		if m.Chat.Type != tb.ChatPrivate {
			session.InlineMsg = utils.InitCustomEditable(msg.MessageSig())
		}

		err = ch.setSession(session, m.Sender, m.Chat)
		if err != nil {
			ch.handler.SendError(m.Chat, err)
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
	} else {
		ch.HandleText(m)
	}
}
func (ch *CalendarHandlers) HandleUserChange(m *tb.Message) {
	session, err := ch.getSession(m.Sender, m.Chat)
	if err != nil {
		return
	}

	if session.IsCreate {
		session.Step = telegram.StepCreateUser
		var replyMarkup *tb.ReplyMarkup = nil
		var replyTo *tb.Message = nil
		if m.Chat.Type != tb.ChatPrivate {
			replyMarkup = &tb.ReplyMarkup{
				InlineKeyboard: calendarInlineKeyboards.GetCreateOptionButtons(session),
			}
			replyTo = m
		}

		msg, err := ch.handler.bot.Send(m.Chat, calendarMessages.CreateEventUserText, &tb.SendOptions{
			ParseMode:   tb.ModeHTML,
			ReplyMarkup: replyMarkup,
			ReplyTo:     replyTo,
		})

		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}

		if m.Chat.Type != tb.ChatPrivate {
			session.InlineMsg = utils.InitCustomEditable(msg.MessageSig())
		}

		err = ch.setSession(session, m.Sender, m.Chat)
		if err != nil {
			ch.handler.SendError(m.Chat, err)
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
	} else {
		ch.HandleText(m)
	}
}
func (ch *CalendarHandlers) HandleStartTimeChange(m *tb.Message) {
	session, err := ch.getSession(m.Sender, m.Chat)
	if err != nil {
		return
	}

	if session.IsCreate {
		session.Step = telegram.StepCreateFrom
		var replyMarkup *tb.ReplyMarkup = nil
		var replyTo *tb.Message = nil
		if m.Chat.Type != tb.ChatPrivate {
			replyMarkup = &tb.ReplyMarkup{
				InlineKeyboard: calendarInlineKeyboards.GetCreateOptionButtons(session),
			}
			replyTo = m
		}
		msg, err := ch.handler.bot.Send(m.Chat, calendarMessages.GetCreateInitText(), &tb.SendOptions{
			ParseMode:   tb.ModeHTML,
			ReplyMarkup: replyMarkup,
			ReplyTo:     replyTo,
		})

		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}

		if m.Chat.Type != tb.ChatPrivate {
			session.InlineMsg = utils.InitCustomEditable(msg.MessageSig())
		}

		err = ch.setSession(session, m.Sender, m.Chat)
		if err != nil {
			ch.handler.SendError(m.Chat, err)
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
	} else {
		ch.HandleText(m)
	}
}
func (ch *CalendarHandlers) HandleStopTimeChange(m *tb.Message) {
	session, err := ch.getSession(m.Sender, m.Chat)
	if err != nil {
		return
	}

	if session.IsCreate {
		session.Step = telegram.StepCreateTo
		var replyMarkup *tb.ReplyMarkup = nil
		var replyTo *tb.Message = nil
		if m.Chat.Type != tb.ChatPrivate {
			replyMarkup = &tb.ReplyMarkup{
				InlineKeyboard: calendarInlineKeyboards.GetCreateOptionButtons(session),
			}
			replyTo = m
		}

		msg, err := ch.handler.bot.Send(m.Chat, calendarMessages.GetCreateEventToText(), &tb.SendOptions{
			ParseMode:   tb.ModeHTML,
			ReplyMarkup: replyMarkup,
			ReplyTo:     replyTo,
		})

		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}

		if m.Chat.Type != tb.ChatPrivate {
			session.InlineMsg = utils.InitCustomEditable(msg.MessageSig())
		}

		err = ch.setSession(session, m.Sender, m.Chat)
		if err != nil {
			ch.handler.SendError(m.Chat, err)
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
	} else {
		ch.HandleText(m)
	}
}
func (ch *CalendarHandlers) HandleLocationChange(m *tb.Message) {
	session, err := ch.getSession(m.Sender, m.Chat)
	if err != nil {
		return
	}

	if session.IsCreate {
		session.Step = telegram.StepCreateLocation
		var replyMarkup *tb.ReplyMarkup = nil
		var replyTo *tb.Message = nil
		if m.Chat.Type != tb.ChatPrivate {
			replyMarkup = &tb.ReplyMarkup{
				InlineKeyboard: calendarInlineKeyboards.GetCreateOptionButtons(session),
			}
			replyTo = m
		}

		msg, err := ch.handler.bot.Send(m.Chat, calendarMessages.CreateEventLocationText, &tb.SendOptions{
			ParseMode:   tb.ModeHTML,
			ReplyMarkup: replyMarkup,
			ReplyTo:     replyTo,
		})

		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}

		if m.Chat.Type != tb.ChatPrivate {
			session.InlineMsg = utils.InitCustomEditable(msg.MessageSig())
		}

		err = ch.setSession(session, m.Sender, m.Chat)
		if err != nil {
			ch.handler.SendError(m.Chat, err)
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
	} else {
		ch.HandleText(m)
	}
}
func (ch *CalendarHandlers) HandleFullDayChange(m *tb.Message) {
	session, err := ch.getSession(m.Sender, m.Chat)
	if err != nil {
		return
	}

	if session.IsCreate {
		session.Event.FullDay = true
		session.Event.To = session.Event.From.Add(24 * time.Hour)
		if session.InfoMsg.ChatID != 0 {
			err := ch.handler.bot.Delete(&session.InfoMsg)
			if err != nil {
				customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
			}
		}

		newMsg, err := ch.handler.bot.Send(m.Chat,
			calendarMessages.GetCreateEventHeader()+calendarMessages.SingleEventFullText(&session.Event),
			&tb.SendOptions{
				ParseMode: tb.ModeHTML,
				ReplyTo:   m,
				ReplyMarkup: &tb.ReplyMarkup{
					InlineKeyboard: calendarInlineKeyboards.CreateEventButtons(session.Event),
				},
			})

		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}

		if session.Event.Title == "" {
			replyMarkup := tb.ReplyMarkup{}
			var replyTo *tb.Message = nil
			if m.Chat.Type != tb.ChatPrivate {
				replyMarkup = tb.ReplyMarkup{
					InlineKeyboard: calendarInlineKeyboards.GetCreateOptionButtons(session),
				}
				replyTo = m
			} else {
				replyMarkup = tb.ReplyMarkup{
					ReplyKeyboard:   calendarKeyboards.GetCreateOptionButtons(session),
					OneTimeKeyboard: true,
				}
			}

			msg, err := ch.handler.bot.Send(m.Chat, calendarMessages.GetCreateEventTitle(), &tb.SendOptions{
				ParseMode:   tb.ModeHTML,
				ReplyMarkup: &replyMarkup,
				ReplyTo:     replyTo,
			})

			if err != nil {
				customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
			}

			if m.Chat.Type != tb.ChatPrivate {
				session.InlineMsg = utils.InitCustomEditable(msg.MessageSig())
				err = ch.setSession(session, m.Sender, m.Chat)
				if err != nil {
					return
				}
			}
		}

		session.InfoMsg = utils.InitCustomEditable(newMsg.MessageSig())

		err = ch.setSession(session, m.Sender, m.Chat)
		if err != nil {
			ch.handler.SendError(m.Chat, err)
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
	} else {
		ch.HandleText(m)
	}
}

func (ch *CalendarHandlers) HandleShowMore(c *tb.Callback) {
	if !ch.AuthMiddleware(c.Sender, c.Message.Chat) {
		err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
		})
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
		return
	}
	event := ch.getEventByIdForCallback(c, c.Sender.ID)
	if event == nil {
		return
	}

	err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
		CallbackID: c.ID,
		Text:       calendarMessages.CallbackResponseHeader(event),
	})
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	_, err = ch.handler.bot.Edit(c.Message, calendarMessages.SingleEventFullText(event),
		&tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: calendarInlineKeyboards.EventShowLessInlineKeyboard(event),
			},
		})
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

}
func (ch *CalendarHandlers) HandleShowLess(c *tb.Callback) {
	if !ch.AuthMiddleware(c.Sender, c.Message.Chat) {
		err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
		})
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
		return
	}
	event := ch.getEventByIdForCallback(c, c.Sender.ID)
	if event == nil {
		return
	}

	err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
		CallbackID: c.ID,
	})
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	inlineKeyboard, err := calendarInlineKeyboards.EventShowMoreInlineKeyboard(event, ch.redisDB)
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}
	_, err = ch.handler.bot.Edit(c.Message, calendarMessages.SingleEventShortText(event),
		&tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: inlineKeyboard,
			},
		})
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}
}
func (ch *CalendarHandlers) HandleAlertYes(c *tb.Callback) {
	if c.Sender.ID != c.Message.ReplyTo.Sender.ID {
		err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
			Text:       calendarMessages.GetUserNotAllow(),
			ShowAlert:  true,
		})
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
		return
	}

	err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
		CallbackID: c.ID,
	})
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	c.Message.Sender = c.Message.ReplyTo.Sender

	err = ch.handler.bot.Delete(c.Message)
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	switch c.Data {
	case telegram.Today:
		ch.HandleToday(c.Message)
	case telegram.Next:
		ch.HandleNext(c.Message)
	case telegram.Date:
		ch.HandleDate(c.Message)
	}

}
func (ch *CalendarHandlers) HandleAlertNo(c *tb.Callback) {
	if c.Sender.ID != c.Message.ReplyTo.Sender.ID {
		err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
			Text:       calendarMessages.GetUserNotAllow(),
			ShowAlert:  true,
		})
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
		return
	}

	err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
		CallbackID: c.ID,
	})
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}
	err = ch.handler.bot.Delete(c.Message)
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}
}
func (ch *CalendarHandlers) HandleCancelCreateEvent(c *tb.Callback) {

	if c.Sender.ID != c.Message.ReplyTo.Sender.ID {
		err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
			Text:       calendarMessages.GetUserNotAllow(),
			ShowAlert:  true,
		})
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
		return
	}

	session, err := ch.getSession(c.Sender, c.Message.Chat)
	if err != nil {
		return
	}

	if session.InlineMsg.ChatID != 0 {
		_, err := ch.handler.bot.EditReplyMarkup(&session.InlineMsg, nil)
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
	}

	err = ch.handler.bot.Respond(c, &tb.CallbackResponse{
		CallbackID: c.ID,
		Text:       calendarMessages.GetCreateCanceledText(),
	})
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}
	session.IsCreate = false
	session.IsDate = false
	session.FindTimeDone = false
	session.Step = telegram.StepCreateInit
	session.Event = types.Event{}

	if session.InfoMsg.ChatID != 0 {
		err := ch.handler.bot.Delete(&session.InfoMsg)
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
	}

	session.InfoMsg = utils.InitCustomEditable("", 0)

	err = ch.setSession(session, c.Sender, c.Message.Chat)
	if err != nil {
		return
	}

	_, err = ch.handler.bot.Send(c.Message.Chat, calendarMessages.GetCreateCanceledText(), &tb.SendOptions{
		ParseMode: tb.ModeHTML,
		ReplyMarkup: &tb.ReplyMarkup{
			ReplyKeyboardRemove: true,
		},
	})

	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}
}
func (ch *CalendarHandlers) HandleCreateEvent(c *tb.Callback) {
	if c.Sender.ID != c.Message.ReplyTo.Sender.ID {
		err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
			Text:       calendarMessages.GetUserNotAllow(),
			ShowAlert:  true,
		})
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
		return
	}

	session, err := ch.getSession(c.Sender, c.Message.Chat)
	if err != nil {
		return
	}

	token, err := ch.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(int64(c.Sender.ID))
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		return
	}

	session.Event.Uid = uuid.NewString()

	inpEvent := EventToEventInput(session.Event)
	info, err := ch.eventUseCase.CreateEvent(token, inpEvent)

	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		return
	}

	respInfo := &types.CreateEventResp{}
	err = json.Unmarshal(info, respInfo)
	if err != nil {
		return
	}

	err = ch.handler.bot.Respond(c, &tb.CallbackResponse{
		CallbackID: c.ID,
		Text:       calendarMessages.GetEventCreatedText(),
	})
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	session.Event.Calendar = respInfo.Data.CreateEvent.Calendar

	if len(session.Users) > 1 {
		for _, userId := range session.Users {
			if int64(c.Sender.ID) == userId {
				continue
			}
			userToken, err := ch.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(userId)
			if err != nil {
				customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
				continue
			}

			userInfo, err := ch.userUseCase.GetMailruUserInfo(userToken)
			if err != nil {
				customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
				continue
			}

			events, err := ch.eventUseCase.GetEventsByDate(userToken, session.Event.From)
			if err != nil {
				customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
				continue
			}

			userCalId := ""

			if events != nil {
				for _, userEvent := range events.Data.Events {
					if userEvent.Uid == *inpEvent.Uid {
						userCalId = userEvent.Calendar.UID
					}
				}
			}

			if userCalId == "" {
				continue
			}

			_, err = ch.eventUseCase.ChangeStatus(userToken, types.ChangeStatus{
				EventID:    *inpEvent.Uid,
				CalendarID: userCalId,
				Status:     telegram.StatusAccepted,
			})

			if err != nil {
				customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
				continue
			}

			for idx, attendee := range session.Event.Attendees {
				if attendee.Email == userInfo.Email {
					session.Event.Attendees[idx].Name = userInfo.Name
					session.Event.Attendees[idx].Status = telegram.StatusAccepted
					break
				}
			}
		}
	}

	_, err = ch.handler.bot.Send(c.Message.Chat,
		calendarMessages.GetCreatedEventHeader(), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
			},
		})

	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	var groupButtons [][]tb.InlineButton = nil
	if c.Message.Chat.Type == tb.ChatGroup || c.Message.Chat.Type == tb.ChatSuperGroup {
		groupButtons, err = calendarInlineKeyboards.GroupChatButtons(&session.Event, ch.redisDB, c.Sender.ID)
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
	}

	_, err = ch.handler.bot.Send(c.Message.Chat,
		calendarMessages.SingleEventFullText(&session.Event),
		&tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: groupButtons,
			},
		})

	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	session.IsCreate = false
	session.IsDate = false
	session.FindTimeDone = false
	session.Step = telegram.StepCreateInit
	session.Event = types.Event{}

	if session.InfoMsg.ChatID != 0 {
		err = ch.handler.bot.Delete(&session.InfoMsg)
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
	}

	if session.InlineMsg.ChatID != 0 {
		_, err = ch.handler.bot.EditReplyMarkup(&session.InlineMsg, nil)
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
	}

	session.InfoMsg = utils.InitCustomEditable("", 0)
	session.InlineMsg = utils.InitCustomEditable("", 0)

	err = ch.setSession(session, c.Sender, c.Message.Chat)
	if err != nil {
		return
	}
}
func (ch *CalendarHandlers) HandleGroupGo(c *tb.Callback) {
	ch.handleGroup(c, telegram.StatusAccepted)
}
func (ch *CalendarHandlers) HandleGroupNotGo(c *tb.Callback) {
	ch.handleGroup(c, telegram.StatusDeclined)
}
func (ch *CalendarHandlers) HandleGroupFindTimeYes(c *tb.Callback) {
	if c.Sender.ID != c.Message.ReplyTo.Sender.ID {
		err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
			Text:       calendarMessages.GetUserNotAllow(),
			ShowAlert:  true,
		})
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
		return
	}

	session, err := ch.getSession(c.Sender, c.Message.Chat)
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		return
	}

	session.IsCreate = true
	session.IsDate = false

	err = ch.handler.bot.Delete(c.Message)
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	msg, err := ch.handler.bot.Send(c.Message.Chat, calendarMessages.GetFindTimeStartText(), &tb.SendOptions{
		ParseMode: tb.ModeHTML,
		ReplyMarkup: &tb.ReplyMarkup{
			InlineKeyboard: calendarInlineKeyboards.GetDateFastCommand(true),
		},
		ReplyTo: c.Message.ReplyTo,
	})
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	session.InlineMsg = utils.InitCustomEditable(msg.MessageSig())

	err = ch.setSession(session, c.Sender, c.Message.Chat)
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		return
	}
}
func (ch *CalendarHandlers) HandleGroupFindTimeNo(c *tb.Callback) {
	if c.Sender.ID != c.Message.ReplyTo.Sender.ID {
		err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
			Text:       calendarMessages.GetUserNotAllow(),
			ShowAlert:  true,
		})
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
		return
	}

	session, err := ch.getSession(c.Sender, c.Message.Chat)
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		ch.handler.SendError(c.Message.Chat, err)
		return
	}

	session.FindTimeDone = true
	err = ch.setSession(session, c.Sender, c.Message.Chat)
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		return
	}

	err = ch.handler.bot.Respond(c, &tb.CallbackResponse{
		CallbackID: c.ID,
	})
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	err = ch.handler.bot.Delete(c.Message)
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	ch.HandleCreate(c.Message.ReplyTo)
}
func (ch *CalendarHandlers) HandleFindTimeDayPart(c *tb.Callback) {
	if c.Sender.ID != c.Message.ReplyTo.Sender.ID {
		err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
			Text:       calendarMessages.GetUserNotAllow(),
			ShowAlert:  true,
		})
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
		return
	}

	err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
		CallbackID: c.ID,
	})
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	session, err := ch.getSession(c.Sender, c.Message.Chat)
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	if c.Data == calendarMessages.GetCreateCancelText() {
		session.IsCreate = false
		session.IsDate = false
		session.FindTimeDone = false
		session.Step = telegram.StepCreateInit
		session.Event = types.Event{}
		session.FreeBusy = types.FreeBusy{}
		session.FindTimeDayPart = nil

		if session.InfoMsg.ChatID != 0 {
			err := ch.handler.bot.Delete(&session.InfoMsg)
			if err != nil {
				customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
			}
		}

		session.InfoMsg = utils.InitCustomEditable("", 0)

		err = ch.setSession(session, c.Sender, c.Message.Chat)
		if err != nil {
			return
		}

		_, err = ch.handler.bot.Send(c.Message.Chat, calendarMessages.GetCreateCanceledText(), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
			},
		})

		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}

		err = ch.handler.bot.Delete(c.Message)
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
		return
	}

	if c.Data == "All day" {
		session.FindTimeDayPart = nil
		err = ch.setSession(session, c.Sender, c.Message.Chat)
		if err != nil {
			ch.handler.SendError(c.Message.Chat, err)
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}

		err = ch.handler.bot.Delete(c.Message)
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}

		_, err = ch.handler.bot.Send(c.Message.Chat, calendarMessages.FindTimeChooseLengthHeader,
			&tb.SendOptions{
				ParseMode: tb.ModeHTML,
				ReplyMarkup: &tb.ReplyMarkup{
					InlineKeyboard: calendarInlineKeyboards.FindTimeLengthButtons(),
				},
				ReplyTo: c.Message.ReplyTo,
			})
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}

		return
	}

	t, err := time.Parse(time.RFC3339, c.Data)
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		return
	}

	var d time.Duration

	if t.Hour() == 9 {
		d, _ = time.ParseDuration("9h")
	} else {
		d, _ = time.ParseDuration("7h")
	}

	session.FindTimeDayPart = &types.DayPart{
		Start:    t,
		Duration: d,
	}

	err = ch.setSession(session, c.Sender, c.Message.Chat)
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		return
	}

	err = ch.handler.bot.Delete(c.Message)
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	_, err = ch.handler.bot.Send(c.Message.Chat, calendarMessages.FindTimeChooseLengthHeader,
		&tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: calendarInlineKeyboards.FindTimeLengthButtons(),
			},
			ReplyTo: c.Message.ReplyTo,
		})
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}
}
func (ch *CalendarHandlers) HandleFindTimeLength(c *tb.Callback) {
	if c.Sender.ID != c.Message.ReplyTo.Sender.ID {
		err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
			Text:       calendarMessages.GetUserNotAllow(),
			ShowAlert:  true,
		})
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
		return
	}

	err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
		CallbackID: c.ID,
	})
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	session, err := ch.getSession(c.Sender, c.Message.Chat)
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	if c.Data == calendarMessages.GetCreateCancelText() {
		session = &types.BotRedisSession{}

		if session.InfoMsg.ChatID != 0 {
			err := ch.handler.bot.Delete(&session.InfoMsg)
			if err != nil {
				customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
			}
		}

		session.InfoMsg = utils.InitCustomEditable("", 0)

		err = ch.setSession(session, c.Sender, c.Message.Chat)
		if err != nil {
			return
		}

		_, err = ch.handler.bot.Send(c.Message.Chat, calendarMessages.GetCreateCanceledText(), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
			},
		})

		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}

		err = ch.handler.bot.Delete(c.Message)
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
		return
	}

	d, _ := time.ParseDuration(c.Data)

	session.FindTimeDuration = d

	err = ch.setSession(session, c.Sender, c.Message.Chat)
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	err = ch.handler.bot.Delete(c.Message)
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	ch.sendOrUpdateVote(session, c.Message.Chat, c.Sender, c.Sender, c.Message.ReplyTo, false)
}
func (ch *CalendarHandlers) FindTimeAdd(c *tb.Callback) {

	userId, err := strconv.Atoi(c.Data)
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		return
	}

	session, err := ch.getSession(&tb.User{ID: userId}, c.Message.Chat)
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		return
	}

	for _, uId := range session.Users {
		if uId == int64(c.Sender.ID) {
			err = ch.handler.bot.Respond(c, &tb.CallbackResponse{
				CallbackID: c.ID,
				Text:       calendarMessages.FindTimeExist,
			})
			if err != nil {
				customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
			}
			return
		}
	}

	ch.sendOrUpdateVote(session, c.Message.Chat, c.Sender, c.Message.ReplyTo.Sender, c.Message.ReplyTo, false)
}
func (ch *CalendarHandlers) HandleFindTimeFind(c *tb.Callback) {
	if c.Sender.ID != c.Message.ReplyTo.Sender.ID {
		err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
			Text:       calendarMessages.GetUserNotAllow(),
			ShowAlert:  true,
		})
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
		return
	}

	userId, err := strconv.Atoi(c.Data)
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		return
	}

	session, err := ch.getSession(&tb.User{ID: userId}, c.Message.Chat)
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		return
	}
	ch.sendOrUpdateVote(session, c.Message.Chat, c.Sender, c.Message.ReplyTo.Sender, c.Message.ReplyTo, true)
}
func (ch *CalendarHandlers) HandleFindTimeBack(c *tb.Callback) {
	if c.Sender.ID != c.Message.ReplyTo.Sender.ID {
		err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
			Text:       calendarMessages.GetUserNotAllow(),
			ShowAlert:  true,
		})
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
		return
	}

	session, err := ch.getSession(c.Sender, c.Message.Chat)
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		return
	}

	if session.PollMsg.ChatID != 0 {
		err = ch.handler.bot.Delete(&session.PollMsg)
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
	}

	session.PollMsg.ChatID = 0
	session.PollMsg.MessageID = ""
	session.FreeBusy = types.FreeBusy{}

	err = ch.setSession(session, c.Sender, c.Message.Chat)
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		return
	}

	ch.HandleGroupFindTimeYes(c)
}
func (ch *CalendarHandlers) FindTimeCreate(c *tb.Callback) {
	if c.Sender.ID != c.Message.ReplyTo.Sender.ID {
		err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
			Text:       calendarMessages.GetUserNotAllow(),
			ShowAlert:  true,
		})
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
		return
	}

	session, err := ch.getSession(c.Sender, c.Message.Chat)
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	err = ch.handler.bot.Respond(c, &tb.CallbackResponse{
		CallbackID: c.ID,
	})
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	if c.Data == calendarMessages.GetCreateCancelText() {
		if session.PollMsg.ChatID != 0 {
			err = ch.handler.bot.Delete(&session.PollMsg)
			if err != nil {
				customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
			}
		}

		if session.InlineMsg.ChatID != 0 {
			_, err = ch.handler.bot.EditReplyMarkup(&session.InlineMsg, nil)
			if err != nil {
				customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
			}
		}

		if session.InfoMsg.ChatID != 0 {
			err = ch.handler.bot.Delete(&session.InfoMsg)
			if err != nil {
				customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
			}
		}

		session = &types.BotRedisSession{}
		err = ch.setSession(session, c.Sender, c.Message.Chat)
		if err != nil {
			ch.handler.SendError(c.Message.Chat, err)
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}

		_, err = ch.handler.bot.Send(c.Message.Chat, calendarMessages.GetCreateCanceledText())
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}

		return
	}

	text := ""
	vc := 0
	for _, options := range c.Message.Poll.Options {
		if options.VoterCount > vc {
			text = options.Text
			vc = options.VoterCount
		}
	}

	resp := ch.ParseEvent(&tb.Message{Text: text, Chat: c.Message.Chat})
	if resp == nil {
		return
	}

	if resp.EventStart.IsZero() || resp.EventEnd.IsZero() {
		return
	}

	err = ch.handler.bot.Delete(&session.PollMsg)
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	token, err := ch.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(int64(c.Sender.ID))
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	userInfo, err := ch.userUseCase.GetMailruUserInfo(token)
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	session.FindTimeDone = true
	session.IsCreate = true
	session.Event.From = resp.EventStart
	session.Event.To = resp.EventEnd
	session.Event.Organizer = types.AttendeeEvent{
		Email:  userInfo.Email,
		Name:   userInfo.Name,
		Role:   telegram.RoleRequired,
		Status: telegram.StatusAccepted,
	}
	users, err := ch.userUseCase.TryGetUsersEmailsByTelegramUserIDs(session.Users)
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	} else {
		for _, user := range users {
			session.Event.Attendees = append(session.Event.Attendees, types.AttendeeEvent{
				Email:  user,
				Role:   telegram.RoleRequired,
				Status: telegram.StatusAccepted,
			})
		}
	}

	newMsg, err := ch.handler.bot.Send(c.Message.Chat,
		calendarMessages.GetCreateEventHeader()+calendarMessages.SingleEventFullText(&session.Event),
		&tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyTo:   c.Message.ReplyTo,
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: calendarInlineKeyboards.CreateEventButtons(session.Event),
			},
		})

	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	session.InfoMsg = utils.InitCustomEditable(newMsg.MessageSig())

	session.Step = telegram.StepCreateTitle

	msg, err := ch.handler.bot.Send(c.Message.Chat, calendarMessages.GetCreateEventTitle(), &tb.SendOptions{
		ParseMode: tb.ModeHTML,
		ReplyMarkup: &tb.ReplyMarkup{
			InlineKeyboard: calendarInlineKeyboards.GetCreateOptionButtons(session),
		},
	})

	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	session.InlineMsg = utils.InitCustomEditable(msg.MessageSig())

	err = ch.setSession(session, c.Sender, c.Message.Chat)
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}
}
func (ch *CalendarHandlers) HandleGroupText(c *tb.Callback) {
	if c.Sender.ID != c.Message.ReplyTo.Sender.ID {
		err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
			Text:       calendarMessages.GetUserNotAllow(),
			ShowAlert:  true,
		})
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
		return
	}

	err := ch.handler.bot.Respond(c, &tb.CallbackResponse{CallbackID: c.ID})
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	_, err = ch.handler.bot.EditReplyMarkup(c.Message, nil)
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	c.Message.ReplyTo.Text = c.Data

	switch c.Data {
	case calendarMessages.CreateEventAddDescButton, calendarMessages.CreateEventChangeDescButton:
		ch.HandleDescChange(c.Message.ReplyTo)
	case calendarMessages.CreateEventAddTitleButton, calendarMessages.CreateEventChangeTitleButton:
		ch.HandleTitleChange(c.Message.ReplyTo)
	case calendarMessages.CreateEventAddLocationButton, calendarMessages.CreateEventChangeLocationButton:
		ch.HandleLocationChange(c.Message.ReplyTo)
	case calendarMessages.CreateEventAddUser:
		ch.HandleUserChange(c.Message.ReplyTo)
	case calendarMessages.CreateEventChangeStartTimeButton:
		ch.HandleStartTimeChange(c.Message.ReplyTo)
	case calendarMessages.CreateEventChangeStopTimeButton:
		ch.HandleStopTimeChange(c.Message.ReplyTo)
	case calendarMessages.GetCreateFullDay():
		ch.HandleFullDayChange(c.Message.ReplyTo)
	default:
		ch.HandleText(c.Message.ReplyTo)
	}
}

func (ch *CalendarHandlers) getSession(user *tb.User, chat *tb.Chat) (*types.BotRedisSession, error) {
	resp, err := ch.redisDB.Get(context.TODO(), strconv.Itoa(int(chat.ID))+strconv.Itoa(user.ID)).Result()
	if err != nil {
		newSession := &types.BotRedisSession{}
		err = ch.setSession(newSession, user, chat)

		if err != nil {
			customerrors.HandlerError(err, &chat.ID, nil)
			_, err = ch.handler.bot.Send(user, calendarMessages.RedisSessionNotFound(), &tb.SendOptions{
				ParseMode: tb.ModeHTML,
				ReplyMarkup: &tb.ReplyMarkup{
					ReplyKeyboardRemove: true,
				},
			})
			if err != nil {
				customerrors.HandlerError(err, &chat.ID, nil)
			}
			return nil, err
		}

		return newSession, nil
	}

	session := &types.BotRedisSession{}
	err = json.Unmarshal([]byte(resp), session)
	if err != nil {
		customerrors.HandlerError(err, &chat.ID, nil)
		_, err := ch.handler.bot.Send(user, calendarMessages.RedisSessionNotFound(), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
			},
		})
		if err != nil {
			customerrors.HandlerError(err, &chat.ID, nil)
		}
		return nil, err
	}

	return session, nil
}
func (ch *CalendarHandlers) setSession(session *types.BotRedisSession, user *tb.User, chat *tb.Chat) error {
	b, err := json.Marshal(session)
	if err != nil {
		customerrors.HandlerError(err, &chat.ID, nil)
		ch.handler.SendError(user, err)
		return err
	}
	err = ch.redisDB.Set(context.TODO(), strconv.Itoa(int(chat.ID))+strconv.Itoa(user.ID), b, 0).Err()
	if err != nil {
		customerrors.HandlerError(err, &chat.ID, nil)
		ch.handler.SendError(user, err)
		return err
	}

	return nil
}
func (ch *CalendarHandlers) sendShortEvents(events *types.Events, chat *tb.Chat) {
	for _, event := range *events {
		var err error
		var keyboard [][]tb.InlineButton = nil
		if chat.Type == tb.ChatPrivate {
			keyboard, err = calendarInlineKeyboards.EventShowMoreInlineKeyboard(&event, ch.redisDB)
			if err != nil {
				zap.S().Errorf("Can't set calendarId=%v for eventId=%v. Err: %v",
					event.Calendar.UID, event.Uid, err)
			}
		}
		_, err = ch.handler.bot.Send(chat, calendarMessages.SingleEventShortText(&event), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: keyboard,
			},
		})
		if err != nil {
			customerrors.HandlerError(err, &chat.ID, nil)
		}
	}
}
func (ch *CalendarHandlers) getEventByIdForCallback(c *tb.Callback, senderID int) *types.Event {
	calUid, err := ch.redisDB.Get(context.TODO(), c.Data).Result()
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		err = ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
			Text:       calendarMessages.RedisNotFoundMessage(),
			ShowAlert:  true,
		})
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
		return nil
	}

	token, err := ch.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(int64(senderID))

	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		ch.handler.SendAuthError(c.Message.Chat, err)

		err = ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
		})
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
		return nil
	}
	resp, err := ch.eventUseCase.GetEventByEventID(token, calUid, c.Data)
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		ch.handler.SendError(c.Message.Chat, err)

		err = ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
		})
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
		return nil
	}

	return &resp.Data.Event
}
func (ch *CalendarHandlers) handleCreateText(m *tb.Message, session *types.BotRedisSession) {
	if calendarMessages.GetCreateCancelText() == m.Text {
		session = &types.BotRedisSession{}

		if session.InfoMsg.ChatID != 0 {
			err := ch.handler.bot.Delete(&session.InfoMsg)
			if err != nil {
				customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
			}
		}

		if session.InlineMsg.ChatID != 0 {
			_, err := ch.handler.bot.EditReplyMarkup(&session.InlineMsg, nil)
			if err != nil {
				customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
			}
		}

		session.InfoMsg = utils.InitCustomEditable("", 0)
		session.InlineMsg = utils.InitCustomEditable("", 0)

		err := ch.setSession(session, m.Sender, m.Chat)
		if err != nil {
			return
		}

		_, err = ch.handler.bot.Send(m.Chat, calendarMessages.GetCreateCanceledText(), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
			},
		})

		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}

		return
	}

Step:
	switch session.Step {
	case telegram.StepCreateFrom:
		parsedDate := ch.ParseDate(m)
		if parsedDate == nil {
			return
		}

		if parsedDate.Date.IsZero() {
			_, err := ch.handler.bot.Send(m.Chat, calendarMessages.GetDateNotParsed())
			if err != nil {
				customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
			}
			return
		}

		if session.Event.To.IsZero() {
			session.Step = telegram.StepCreateTo
		} else {
			session.Event.To = parsedDate.Date.Add(session.Event.To.Sub(session.Event.From))
		}
		session.Event.From = parsedDate.Date
		break Step
	case telegram.StepCreateTo:
		session.Event.FullDay = false
		switch m.Text {
		case calendarMessages.GetCreateEventHalfHour():
			session.Event.To = session.Event.From.Add(30 * time.Minute)
			if session.Event.Title == "" {
				session.Step = telegram.StepCreateTitle
			}
			break Step
		case calendarMessages.GetCreateEventHour():
			session.Event.To = session.Event.From.Add(1 * time.Hour)
			if session.Event.Title == "" {
				session.Step = telegram.StepCreateTitle
			}
			break Step
		case calendarMessages.GetCreateEventHourAndHalf():
			session.Event.To = session.Event.From.Add(1 * time.Hour).Add(30 * time.Minute)
			if session.Event.Title == "" {
				session.Step = telegram.StepCreateTitle
			}
			break Step
		case calendarMessages.GetCreateEventTwoHours():
			session.Event.To = session.Event.From.Add(2 * time.Hour)
			if session.Event.Title == "" {
				session.Step = telegram.StepCreateTitle
			}
			break Step
		case calendarMessages.GetCreateEventFourHours():
			session.Event.To = session.Event.From.Add(4 * time.Hour)
			if session.Event.Title == "" {
				session.Step = telegram.StepCreateTitle
			}
			break Step
		case calendarMessages.GetCreateEventSixHours():
			session.Event.To = session.Event.From.Add(6 * time.Hour)
			if session.Event.Title == "" {
				session.Step = telegram.StepCreateTitle
			}
			break Step
		case calendarMessages.GetCreateFullDay():
			session.Event.FullDay = true
			session.Event.To = session.Event.From.Add(24 * time.Hour)
			if session.Event.Title == "" {
				session.Step = telegram.StepCreateTitle
			}
			break Step
		}

		parsedDate := ch.ParseDate(m)
		if parsedDate == nil {
			return
		}

		if parsedDate.Date.IsZero() {
			_, err := ch.handler.bot.Send(m.Chat, calendarMessages.GetDateNotParsed())
			if err != nil {
				customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
			}
			return
		}

		if parsedDate.Date.Before(session.Event.From) {
			_, err := ch.handler.bot.Send(m.Chat, calendarMessages.EventDateToIsBeforeFrom, &tb.SendOptions{
				ParseMode: tb.ModeHTML,
			})
			if err != nil {
				customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
			}
			return
		}

		if session.Event.Title == "" {
			session.Step = telegram.StepCreateTitle
		}

		session.Event.To = parsedDate.Date
		break Step
	case telegram.StepCreateTitle:
		session.Event.Title = m.Text
		break Step
	case telegram.StepCreateDesc:
		session.Event.Description = m.Text
		break Step
	case telegram.StepCreateUser:
		attendeesEmails := strings.Split(m.Text, ",")
		for _, email := range attendeesEmails {
			session.Event.Attendees = append(session.Event.Attendees, types.AttendeeEvent{
				Email:  strings.Trim(email, " "),
				Role:   telegram.RoleRequired,
				Status: telegram.StatusNeedsAction,
			})
		}
	case telegram.StepCreateLocation:
		session.Event.Location.Description = m.Text
		break Step
	}

	if session.InfoMsg.ChatID != 0 {
		err := ch.handler.bot.Delete(&session.InfoMsg)
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
	}

	e := session.Event

	newMsg, err := ch.handler.bot.Send(m.Chat,
		calendarMessages.GetCreateEventHeader()+calendarMessages.SingleEventFullText(&session.Event),
		&tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyTo:   m,
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: calendarInlineKeyboards.CreateEventButtons(session.Event),
			},
		})

	if err != nil {
		customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
	}

	session.InfoMsg = utils.InitCustomEditable(newMsg.MessageSig())

	err = ch.setSession(session, m.Sender, m.Chat)
	if err != nil {
		return
	}

	if session.Event.To.IsZero() {

		replyMarkup := tb.ReplyMarkup{}
		var replyTo *tb.Message = nil
		if m.Chat.Type != tb.ChatPrivate {
			replyMarkup = tb.ReplyMarkup{
				InlineKeyboard: calendarInlineKeyboards.GetCreateDuration(),
			}
			replyTo = m
		} else {
			replyMarkup = tb.ReplyMarkup{
				ReplyKeyboard:       calendarKeyboards.GetCreateDuration(),
				ResizeReplyKeyboard: true,
			}
		}

		msg, err := ch.handler.bot.Send(m.Chat, calendarMessages.GetCreateEventToText(), &tb.SendOptions{
			ParseMode:   tb.ModeHTML,
			ReplyMarkup: &replyMarkup,
			ReplyTo:     replyTo,
		})

		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}

		if m.Chat.Type != tb.ChatPrivate {
			session.InlineMsg = utils.InitCustomEditable(msg.MessageSig())
			err = ch.setSession(session, m.Sender, m.Chat)
			if err != nil {
				return
			}
		}

		return
	}

	if e.Title == "" && e.Description == "" && e.Location.Description == "" && len(e.Attendees) < 2 && session.Step == telegram.StepCreateTitle {
		session.FromTextCreate = false

		replyMarkup := tb.ReplyMarkup{}
		var replyTo *tb.Message = nil
		if m.Chat.Type != tb.ChatPrivate {
			replyMarkup = tb.ReplyMarkup{
				InlineKeyboard: calendarInlineKeyboards.GetCreateOptionButtons(session),
			}
			replyTo = m
		} else {
			replyMarkup = tb.ReplyMarkup{
				ReplyKeyboard:   calendarKeyboards.GetCreateOptionButtons(session),
				OneTimeKeyboard: true,
			}
		}

		msg, err := ch.handler.bot.Send(m.Chat, calendarMessages.GetCreateEventTitle(), &tb.SendOptions{
			ParseMode:   tb.ModeHTML,
			ReplyMarkup: &replyMarkup,
			ReplyTo:     replyTo,
		})

		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}

		if m.Chat.Type != tb.ChatPrivate {
			session.InlineMsg = utils.InitCustomEditable(msg.MessageSig())
		}

		err = ch.setSession(session, m.Sender, m.Chat)
		if err != nil {
			return
		}

		return
	}

	if e.Title != "" && !e.To.IsZero() && !e.From.IsZero() && session.FromTextCreate {
		session.FromTextCreate = false
		session.Step = telegram.StepCreateDesc
		replyMarkup := tb.ReplyMarkup{}
		var replyTo *tb.Message = nil
		if m.Chat.Type != tb.ChatPrivate {
			replyMarkup = tb.ReplyMarkup{
				InlineKeyboard: calendarInlineKeyboards.GetCreateOptionButtons(session),
			}
			replyTo = m
		} else {
			replyMarkup = tb.ReplyMarkup{
				ReplyKeyboard:   calendarKeyboards.GetCreateOptionButtons(session),
				OneTimeKeyboard: true,
			}
		}
		msg, err := ch.handler.bot.Send(m.Chat, calendarMessages.CreateEventDescText, &tb.SendOptions{
			ParseMode:   tb.ModeHTML,
			ReplyMarkup: &replyMarkup,
			ReplyTo:     replyTo,
		})

		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}

		if m.Chat.Type != tb.ChatPrivate {
			session.InlineMsg = utils.InitCustomEditable(msg.MessageSig())
		}

		err = ch.setSession(session, m.Sender, m.Chat)
		if err != nil {
			return
		}

		return
	}

}
func (ch *CalendarHandlers) handleFindTimeText(m *tb.Message, session *types.BotRedisSession) {
	if calendarMessages.GetCreateCancelText() == m.Text {
		session = &types.BotRedisSession{}

		if session.InfoMsg.ChatID != 0 {
			err := ch.handler.bot.Delete(&session.InfoMsg)
			if err != nil {
				customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
			}
		}

		if session.InlineMsg.ChatID != 0 {
			_, err := ch.handler.bot.EditReplyMarkup(&session.InlineMsg, nil)
			if err != nil {
				customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
			}
		}

		session.InfoMsg = utils.InitCustomEditable("", 0)
		session.InlineMsg = utils.InitCustomEditable("", 0)

		err := ch.setSession(session, m.Sender, m.Chat)
		if err != nil {
			return
		}

		_, err = ch.handler.bot.Send(m.Chat, calendarMessages.GetCreateCanceledText(), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
			},
		})

		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}

		return
	}

	if !session.FreeBusy.From.IsZero() && !session.FreeBusy.To.IsZero() {
		return
	}

	resp := ch.ParseDate(m)
	if resp == nil {
		return
	}

	if resp.Date.IsZero() {
		_, err := ch.handler.bot.Send(m.Chat, calendarMessages.GetDateNotParsed())
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
		return
	}

	if session.FreeBusy.From.IsZero() {
		t := resp.Date
		session.FreeBusy.From = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

		err := ch.setSession(session, m.Sender, m.Chat)
		if err != nil {
			ch.handler.SendError(m.Chat, err)
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
			return
		}

		_, err = ch.handler.bot.Send(m.Chat, calendarMessages.GetFindTimeStopText(session.FreeBusy.From),
			&tb.SendOptions{
				ParseMode: tb.ModeHTML,
				ReplyMarkup: &tb.ReplyMarkup{
					InlineKeyboard: calendarInlineKeyboards.GetDateFastCommand(true),
				},
				ReplyTo: m,
			})
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
		return
	}

	if session.FreeBusy.To.IsZero() {
		if resp.Date.Before(session.FreeBusy.From) {
			_, err := ch.handler.bot.Send(m.Chat, calendarMessages.EventDateToIsBeforeFrom, &tb.SendOptions{
				ParseMode: tb.ModeHTML,
			})
			if err != nil {
				customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
			}
		} else {
			t := resp.Date
			if t.Sub(session.FreeBusy.From).Hours() > 346 {
				t = session.FreeBusy.From
				session.FreeBusy.To = time.Date(t.Year(), t.Month(), t.Day()+14, 23, 59, 59, 0, t.Location())
				_, err := ch.handler.bot.Send(m.Chat, calendarMessages.FindTimePeriodIsTooLong)
				if err != nil {
					customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
				}
			} else {
				session.FreeBusy.To = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
			}

			err := ch.setSession(session, m.Sender, m.Chat)
			if err != nil {
				ch.handler.SendError(m.Chat, err)
				customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
				return
			}

			_, err = ch.handler.bot.Send(m.Chat, calendarMessages.GetFindTimeInfoText(session.FreeBusy.From,
				session.FreeBusy.To),
				&tb.SendOptions{
					ParseMode: tb.ModeHTML,
					ReplyMarkup: &tb.ReplyMarkup{
						ReplyKeyboardRemove: true,
					},
				})
			if err != nil {
				customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
			}

			_, err = ch.handler.bot.Send(m.Chat, calendarMessages.FindTimeChooseDayPartHeader,
				&tb.SendOptions{
					ParseMode: tb.ModeHTML,
					ReplyMarkup: &tb.ReplyMarkup{
						InlineKeyboard: calendarInlineKeyboards.FindTimeDayPartButtons(session.FreeBusy.From),
					},
					ReplyTo: m,
				})
			if err != nil {
				customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
			}
		}
	}

}
func (ch *CalendarHandlers) handleDateText(m *tb.Message, session *types.BotRedisSession) {
	if calendarMessages.GetCancelDateReplyButton() == m.Text {
		session = &types.BotRedisSession{}
		err := ch.setSession(session, m.Sender, m.Chat)
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
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}

		return
	}

	parseDate := ch.ParseDate(m)
	if parseDate == nil {
		return
	}

	if !parseDate.Date.IsZero() {
		session.IsDate = false
		err := ch.setSession(session, m.Sender, m.Chat)
		if err != nil {
			return
		}

		token, err := ch.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(int64(m.Sender.ID))
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
			ch.handler.SendError(m.Chat, err)
			return
		}
		events, err := ch.eventUseCase.GetEventsByDate(token, parseDate.Date)
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
			ch.handler.SendError(m.Chat, err)
			return
		}

		if events != nil {
			i := 0
			for _, event := range events.Data.Events {
				if !event.FullDay || event.From.Day() != time.Now().Day()-1 || event.To.Day() != time.Now().Day() {
					events.Data.Events[i] = event
					i++
				}
			}

			events.Data.Events = events.Data.Events[:i]
		}

		if events != nil || len(events.Data.Events) > 0 {
			title := calendarMessages.GetDateTitle(parseDate.Date)
			if m.Chat.Type != tb.ChatPrivate {
				title += calendarMessages.AddNameBold(m.Sender.FirstName + " " + m.Sender.LastName)
			}

			_, err := ch.handler.bot.Send(m.Chat, title, &tb.SendOptions{
				ParseMode: tb.ModeHTML,
				ReplyMarkup: &tb.ReplyMarkup{
					ReplyKeyboardRemove: true,
				},
			})
			if err != nil {
				customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
			}
			ch.sendShortEvents(&events.Data.Events, m.Chat)
		} else {
			title := calendarMessages.GetDateEventsNotFound()
			if m.Chat.Type != tb.ChatPrivate {
				title += calendarMessages.AddName(m.Sender.FirstName + " " + m.Sender.LastName)
			}
			_, err := ch.handler.bot.Send(m.Chat, title, &tb.SendOptions{
				ParseMode: tb.ModeHTML,
				ReplyMarkup: &tb.ReplyMarkup{
					ReplyKeyboardRemove: true,
				},
			})

			if err != nil {
				customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
			}
		}

	} else {
		_, err := ch.handler.bot.Send(m.Chat, calendarMessages.GetDateNotParsed(), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
		})
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
	}
}
func (ch *CalendarHandlers) handleGroup(c *tb.Callback, status string) {
	if !ch.AuthMiddleware(c.Sender, c.Message.Chat) {
		err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
		})
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
		return
	}
	data := strings.Split(c.Data, "|")
	userId, err := strconv.Atoi(data[1])
	c.Data = data[0]
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}
	event := ch.getEventByIdForCallback(c, userId)
	if event == nil {
		return
	}

	eventToken, err := ch.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(int64(userId))
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		return
	}

	token, err := ch.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(int64(c.Sender.ID))
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		return
	}

	userInfo, err := ch.userUseCase.GetMailruUserInfo(token)
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		return
	}

	if event.Organizer.Email == userInfo.Email {
		err = ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
			Text:       calendarMessages.CreateEventAlreadyOrganize,
		})
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}
		return
	}

	for idx, attendee := range event.Attendees {
		if attendee.Email == userInfo.Email {
			if attendee.Status == status {
				text := ""
				if status == telegram.StatusAccepted {
					text = calendarMessages.CreateEventAlreadyGo
				} else {
					text = calendarMessages.CreateEventAlreadyNotGo
				}
				err = ch.handler.bot.Respond(c, &tb.CallbackResponse{
					CallbackID: c.ID,
					Text:       text,
				})
				if err != nil {
					customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
				}
			} else {
				if attendee.Status == telegram.StatusDeclined {
					_, err = ch.eventUseCase.AddAttendee(eventToken, types.AddAttendee{
						EventID:    event.Uid,
						CalendarID: event.Calendar.UID,
						Email:      userInfo.Email,
						Role:       telegram.RoleRequired,
					})

					if err != nil {
						ch.handler.SendError(c.Message.Chat, err)
						customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
					}
				}

				err = ch.ChangeStatusCallback(c, token, event, status)
				if err != nil {
					err = ch.handler.bot.Respond(c, &tb.CallbackResponse{
						CallbackID: c.ID,
					})
					if err != nil {
						customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
					}
					return
				}

				err = ch.handler.bot.Respond(c, &tb.CallbackResponse{
					CallbackID: c.ID,
				})

				if err != nil {
					customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
				}

				event.Attendees[idx].Status = status

				inlineKeyboard, err := calendarInlineKeyboards.GroupChatButtons(event, ch.redisDB, userId)

				if err != nil {
					ch.handler.SendError(c.Message.Chat, err)
					customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
					return
				}

				_, err = ch.handler.bot.Edit(c.Message, calendarMessages.SingleEventFullText(event), &tb.SendOptions{
					ParseMode: tb.ModeHTML,
					ReplyMarkup: &tb.ReplyMarkup{
						InlineKeyboard: inlineKeyboard,
					},
				})

				if err != nil {
					customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
				}

				return
			}

			return
		}
	}

	_, err = ch.eventUseCase.AddAttendee(eventToken, types.AddAttendee{
		EventID:    event.Uid,
		CalendarID: event.Calendar.UID,
		Email:      userInfo.Email,
		Role:       telegram.RoleRequired,
	})

	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	err = ch.ChangeStatusCallback(c, token, event, status)
	if err != nil {
		return
	}

	err = ch.handler.bot.Respond(c, &tb.CallbackResponse{
		CallbackID: c.ID,
	})
	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	event.Attendees = append(event.Attendees, types.AttendeeEvent{
		Email:  userInfo.Email,
		Name:   userInfo.Name,
		Role:   telegram.RoleRequired,
		Status: status,
	})

	inlineKeyboard, err := calendarInlineKeyboards.GroupChatButtons(event, ch.redisDB, userId)

	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		return
	}

	_, err = ch.handler.bot.Edit(c.Message, calendarMessages.SingleEventFullText(event), &tb.SendOptions{
		ParseMode: tb.ModeHTML,
		ReplyMarkup: &tb.ReplyMarkup{
			InlineKeyboard: inlineKeyboard,
		},
	})

	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}
}
func (ch *CalendarHandlers) ParseDate(m *tb.Message) *types.ParseDateResp {
	reqData := &types.ParseDateReq{Timezone: "Europe/Moscow", Text: m.Text}
	b, err := json.Marshal(reqData)
	if err != nil {
		customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		ch.handler.SendError(m.Chat, err)
		return nil
	}

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, ch.handler.parseAddress+"/parse/date", bytes.NewBuffer(b))
	if err != nil {
		customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		ch.handler.SendError(m.Chat, err)
		return nil
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		ch.handler.SendError(m.Chat, err)
		return nil
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
	}(resp.Body)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ch.handler.SendError(m.Chat, err)
		customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
	}

	parseDate := &types.ParseDateResp{}
	err = json.Unmarshal(body, parseDate)
	if err != nil {
		customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		ch.handler.SendError(m.Chat, err)
		return nil
	}

	return parseDate
}
func (ch *CalendarHandlers) ParseEvent(m *tb.Message) *types.ParseEventResp {
	reqData := &types.ParseDateReq{Timezone: "Europe/Moscow", Text: m.Text}
	b, err := json.Marshal(reqData)
	if err != nil {
		customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		ch.handler.SendError(m.Chat, err)
		return nil
	}

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, ch.handler.parseAddress+"/parse/event", bytes.NewBuffer(b))
	if err != nil {
		customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		ch.handler.SendError(m.Chat, err)
		return nil
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		ch.handler.SendError(m.Chat, err)
		return nil
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
	}(resp.Body)

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		ch.handler.SendError(m.Chat, err)
		customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
	}

	parseDate := &types.ParseEventResp{}
	err = json.Unmarshal(body, parseDate)
	if err != nil {
		customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		ch.handler.SendError(m.Chat, err)
		return nil
	}

	return parseDate
}
func (ch *CalendarHandlers) sendOrUpdateVote(session *types.BotRedisSession, c *tb.Chat, userAdd *tb.User, userInit *tb.User, msgToReply *tb.Message, sendPoll bool) {

	session.Users = append(session.Users, int64(userAdd.ID))

	emails, err := ch.userUseCase.TryGetUsersEmailsByTelegramUserIDs(session.Users)
	if err != nil {
		ch.handler.SendError(c, err)
		customerrors.HandlerError(err, &c.ID, &msgToReply.ID)
		return
	}

	session.FreeBusy.Users = emails

	if !sendPoll {

		if session.PollMsg.ChatID == 0 {
			msg, err := ch.handler.bot.Send(c, calendarMessages.GenFindTimePollHeader(emails), &tb.SendOptions{
				ParseMode: tb.ModeHTML,
				ReplyTo:   msgToReply,
				ReplyMarkup: &tb.ReplyMarkup{
					InlineKeyboard: calendarInlineKeyboards.FindTimeAddUser(userInit.ID),
				},
			})

			if err != nil {
				customerrors.HandlerError(err, &c.ID, &msgToReply.ID)
				ch.handler.SendError(c, err)
				return
			}

			session.PollMsg = utils.InitCustomEditable(msg.MessageSig())

			err = ch.setSession(session, userInit, c)
			if err != nil {
				ch.handler.SendError(c, err)
				customerrors.HandlerError(err, &c.ID, &msgToReply.ID)
			}

			return
		}

		_, err = ch.handler.bot.Edit(&session.PollMsg, calendarMessages.GenFindTimePollHeader(emails), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyTo:   msgToReply,
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: calendarInlineKeyboards.FindTimeAddUser(userInit.ID),
			},
		})

		if err != nil {
			customerrors.HandlerError(err, &c.ID, &msgToReply.ID)
			return
		}

		return
	}

	token, err := ch.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(int64(userInit.ID))
	if err != nil {
		ch.handler.SendError(c, err)
		customerrors.HandlerError(err, &c.ID, &msgToReply.ID)
		return
	}

	stretchBusyIntervalsBy := 15 * time.Minute

	spans, err := ch.eventUseCase.GetUsersFreeIntervals(token, session.FreeBusy, eUseCase.FreeBusyConfig{
		DayPart:                session.FindTimeDayPart,
		StretchBusyIntervalsBy: &stretchBusyIntervalsBy,
		SplitFreeIntervalsBy:   &session.FindTimeDuration,
	})

	if err != nil {
		ch.handler.SendError(c, err)
		customerrors.HandlerError(err, &c.ID, &msgToReply.ID)
		return
	}

	if len(spans) < 1 {
		_, err = ch.handler.bot.Send(c, calendarMessages.FindTimeNotFound)
		if err != nil {
			customerrors.HandlerError(err, &c.ID, &msgToReply.ID)
		}
		return
	}

	if session.PollMsg.ChatID != 0 {
		err = ch.handler.bot.Delete(&session.PollMsg)
		if err != nil {
			customerrors.HandlerError(err, &c.ID, &msgToReply.ID)
		}
	}

	poll := tb.Poll{
		Type:            tb.PollRegular,
		Question:        calendarMessages.GenFindTimePollHeader(emails),
		MultipleAnswers: true,
		ParseMode:       tb.ModeHTML,
	}

	poll.AddOptions(calendarMessages.GenOptionsForPoll(spans)...)

	pollMsg, err := poll.Send(ch.handler.bot, c, &tb.SendOptions{
		ParseMode: tb.ModeHTML,
		ReplyMarkup: &tb.ReplyMarkup{
			InlineKeyboard: calendarInlineKeyboards.FindTimePollButtons(),
		},
		ReplyTo: msgToReply,
	})

	if err != nil {
		customerrors.HandlerError(err, &c.ID, &msgToReply.ID)
	}

	session.PollMsg = utils.InitCustomEditable(pollMsg.MessageSig())

	err = ch.setSession(session, userInit, c)
	if err != nil {
		ch.handler.SendError(c, err)
		customerrors.HandlerError(err, &c.ID, &msgToReply.ID)
	}
}

func (ch *CalendarHandlers) AuthMiddleware(u *tb.User, c *tb.Chat) bool {
	isAuth, err := ch.userUseCase.IsUserAuthenticatedByTelegramUserID(int64(u.ID))
	if err != nil {
		customerrors.HandlerError(err, &c.ID, nil)
		return false
	}

	if !isAuth {
		msg := ""
		if c.Type != tb.ChatPrivate {
			msg += calendarMessages.AddNameStartBold(u.FirstName + " " + u.LastName)
		}
		msg += calendarMessages.GetUserNotAuth()
		_, err = ch.handler.bot.Send(c, msg, &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
			},
		})
		if err != nil {
			customerrors.HandlerError(err, &c.ID, nil)
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
			customerrors.HandlerError(err, &m.Chat.ID, &m.ID)
		}
		return true
	}

	return false
}
func (ch *CalendarHandlers) ChangeStatusCallback(c *tb.Callback, token string, event *types.Event, status string) error {
	userInfo, err := ch.userUseCase.GetMailruUserInfo(token)
	if err != nil {
		return err
	}
	userCalId, err := ch.redisDB.Get(context.TODO(), userInfo.Email+event.Uid).Result()
	if err != nil {
		events, err := ch.eventUseCase.GetEventsByDate(token, event.From)
		if err != nil {
			ch.handler.SendError(c.Message.Chat, err)
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
			return err
		}

		if events != nil {
			for _, userEvent := range events.Data.Events {
				if userEvent.Uid == event.Uid {
					userCalId = userEvent.Calendar.UID
				}
			}
		}
	}

	if userCalId == "" {
		err = ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
			Text:       calendarMessages.CreateEventCannotAdd,
			ShowAlert:  true,
		})
		if err != nil {
			customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		}

		return errors.New("Calendar UID not found")
	}

	err = ch.redisDB.Set(context.TODO(), userInfo.Email+event.Uid, userCalId, 0).Err()

	if err != nil {
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
	}

	_, err = ch.eventUseCase.ChangeStatus(token, types.ChangeStatus{
		EventID:    event.Uid,
		CalendarID: userCalId,
		Status:     status,
	})
	if err != nil {
		ch.handler.SendError(c.Message.Chat, err)
		customerrors.HandlerError(err, &c.Message.Chat.ID, &c.Message.ID)
		return err
	}

	return nil
}

func EventToEventInput(event types.Event) types.EventInput {
	ret := types.EventInput{}

	id := event.Uid
	location, _ := time.LoadLocation("Europe/Moscow")
	from := event.From.In(location).Format(time.RFC3339)
	to := event.To.In(location).Format(time.RFC3339)

	ret.Uid = &id
	ret.From = &from
	ret.To = &to
	ret.FullDay = &event.FullDay

	if event.Title != "" {
		ret.Title = &event.Title
	} else {
		title := "Без названия"
		ret.Title = &title
	}

	if event.Description != "" {
		ret.Description = &event.Description
	} else {
		desc := ""
		ret.Description = &desc
	}

	if event.Location.Description != "" {
		loc := &types.Location{}
		loc.Description = event.Location.Description
		ret.Location = loc
	}

	if len(event.Attendees) > 0 {
		attendees := types.Attendees{}
		for _, attendee := range event.Attendees {
			if event.Organizer.Email == attendee.Email {
				continue
			}
			attendees = append(attendees, types.Attendee{
				Email: attendee.Email,
				Role:  attendee.Role,
			})
		}
		ret.Attendees = &attendees
	}

	return ret
}
