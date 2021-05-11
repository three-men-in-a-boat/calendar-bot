package handlers

import (
	"fmt"
	eUseCase "github.com/calendar-bot/pkg/events/usecase"
	"github.com/calendar-bot/pkg/middlewares"
	"github.com/calendar-bot/pkg/types"
	uUseCase "github.com/calendar-bot/pkg/users/usecase"
	"github.com/calendar-bot/pkg/utils/contextutils"
	"github.com/calendar-bot/pkg/utils/pathutils"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

type EventHandlers struct {
	eventUseCase eUseCase.EventUseCase
	userUseCase  uUseCase.UserUseCase
}

func NewEventHandlers(eventUseCase eUseCase.EventUseCase, userUseCase uUseCase.UserUseCase) EventHandlers {
	return EventHandlers{eventUseCase: eventUseCase, userUseCase: userUseCase}
}

func (eh *EventHandlers) InitHandlers(server *echo.Echo) {
	oauthMiddleware := middlewares.NewCheckOAuthTelegramMiddleware(&eh.userUseCase)

	eventRouter := server.Group("/api/v1/telegram/user/"+pathutils.TelegramUserIDRouteKey+"/events", oauthMiddleware.Handle)

	eventRouter.GET("/today", eh.getEventsToday)
	eventRouter.GET("/closest", eh.getClosestEvent)
	eventRouter.GET("/users/busy", eh.getUsersBusyIntervals)
	eventRouter.GET("/date/:date", eh.getEventsByDate)

	eventRouter.PUT("/calendar/event", eh.getEventByEventID)
	eventRouter.POST("/event/create", eh.createEvent)
	eventRouter.PUT("/calendar/add/attendee", eh.addAttendee)
	eventRouter.PUT("/calendar/change/attendee/status", eh.changeStatus)
}

func (eh *EventHandlers) getEventsToday(ctx echo.Context) error {
	telegramID, err := contextutils.GetTelegramUserIDFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	accessToken, err := contextutils.GetOAuthAccessTokenFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	todayEvent, err := eh.eventUseCase.GetEventsToday(accessToken)
	if err != nil {
		return errors.Wrapf(err, "failed to get today's events for telegramUserID=%d", telegramID)
	}
	if todayEvent == nil {
		return ctx.String(http.StatusNotFound, "There are no events for today")
	}
	ctx.Response().Header().Set("Content-Type", "application/json")

	return ctx.JSON(http.StatusOK, *todayEvent)
}

func (eh *EventHandlers) getClosestEvent(ctx echo.Context) error {
	telegramID, err := contextutils.GetTelegramUserIDFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	accessToken, err := contextutils.GetOAuthAccessTokenFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	closesEvent, err := eh.eventUseCase.GetClosestEvent(accessToken)
	if err != nil {
		return errors.Wrapf(err, "failed to get the closest event for telegramUserID=%d", telegramID)
	}
	if closesEvent == nil {
		return ctx.String(http.StatusNotFound, "There are no events for today")
	}
	ctx.Response().Header().Set("Content-Type", "application/json")

	return ctx.JSON(http.StatusOK, *closesEvent)
}

func (eh *EventHandlers) getEventsByDate(ctx echo.Context) error {
	telegramID, err := contextutils.GetTelegramUserIDFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	accessToken, err := contextutils.GetOAuthAccessTokenFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	dateFromCtx := ctx.Param("date")
	date, err := time.Parse(time.RFC3339, dateFromCtx)
	if err != nil {
		return err
	}

	eventsByDate, err := eh.eventUseCase.GetEventsByDate(accessToken, date)
	if err != nil {
		return errors.Wrapf(err, "failed to get the closest event for telegramUserID=%d", telegramID)
	}
	if eventsByDate == nil {
		return ctx.String(http.StatusNotFound, "There are no events for today")
	}
	ctx.Response().Header().Set("Content-Type", "application/json")

	return ctx.JSON(http.StatusOK, *eventsByDate)
}

type EventCalendarIDs struct {
	CalendarID string `json:"calendar_id,omitempty"`
	EventID    string `json:"event_id,omitempty"`
}

func (eh *EventHandlers) getUsersBusyIntervals(ctx echo.Context) error {
	telegramID, err := contextutils.GetTelegramUserIDFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	accessToken, err := contextutils.GetOAuthAccessTokenFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	freeBusyUsers := types.FreeBusy{}

	if err := ctx.Bind(&freeBusyUsers); err != nil {
		return errors.Wrapf(err, "failed to unmarshal content from body")
	}

	freeBusyResponse, err := eh.eventUseCase.GetUsersBusyIntervals(accessToken, freeBusyUsers)
	if err != nil {
		return errors.Wrapf(err, "failed to get event by event_id and calendar_id for telegramUserID=%d", telegramID)
	}
	if freeBusyResponse == nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to get busy intervals for user with telegram id %d", telegramID))
	}
	ctx.Response().Header().Set("Content-Type", "application/json")

	return ctx.JSON(http.StatusOK, *freeBusyResponse)
}

func (eh *EventHandlers) getEventByEventID(ctx echo.Context) error {
	telegramID, err := contextutils.GetTelegramUserIDFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	accessToken, err := contextutils.GetOAuthAccessTokenFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	eventCalendarIDs := EventCalendarIDs{}

	if err := ctx.Bind(&eventCalendarIDs); err != nil {
		return errors.Wrapf(err, "failed to unmarshal content from body")
	}

	event, err := eh.eventUseCase.GetEventByEventID(accessToken, eventCalendarIDs.CalendarID, eventCalendarIDs.EventID)
	if err != nil {
		return errors.Wrapf(err, "failed to get event by event_id and calendar_id for telegramUserID=%d", telegramID)
	}
	if event == nil {
		return ctx.String(http.StatusNotFound, fmt.Sprintf("event by event_id %s and calendar_id %s is not found for telegram id %d", eventCalendarIDs.EventID, eventCalendarIDs.CalendarID, telegramID))
	}
	ctx.Response().Header().Set("Content-Type", "application/json")

	return ctx.JSON(http.StatusOK, *event)
}

func (eh *EventHandlers) createEvent(ctx echo.Context) error {
	telegramID, err := contextutils.GetTelegramUserIDFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	accessToken, err := contextutils.GetOAuthAccessTokenFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	eventInput := types.EventInput{}

	if err := ctx.Bind(&eventInput); err != nil {
		return errors.Wrapf(err, "failed to unmarshal content from body")
	}

	event, err := eh.eventUseCase.CreateEvent(accessToken, eventInput)
	if err != nil {
		return errors.Wrapf(err, "failed to create event for telegramUserID=%d", telegramID)
	}
	if event == nil {
		return errors.Wrapf(err, "failed to get event after creation for telegramUserID=%d", telegramID)
	}
	ctx.Response().Header().Set("Content-Type", "application/json")
	return ctx.JSON(http.StatusOK, string(event))
}

func (eh *EventHandlers) addAttendee(ctx echo.Context) error {
	telegramID, err := contextutils.GetTelegramUserIDFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	accessToken, err := contextutils.GetOAuthAccessTokenFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	eventInput := types.AddAttendee{}

	if err := ctx.Bind(&eventInput); err != nil {
		return errors.Wrapf(err, "failed to unmarshal content from body")
	}

	attendeeResponse, err := eh.eventUseCase.AddAttendee(accessToken, eventInput)
	if err != nil {
		return errors.Wrapf(err, "failed to add attendee for event of telegramUserID=%d", telegramID)
	}
	if attendeeResponse == nil {
		return errors.Wrapf(err, "failed to add attendee for event of telegramUserID=%d, response is nil", telegramID)
	}

	ctx.Response().Header().Set("Content-Type", "application/json")
	return ctx.JSON(http.StatusOK, string(attendeeResponse))
}

func (eh *EventHandlers) changeStatus(ctx echo.Context) error {
	telegramID, err := contextutils.GetTelegramUserIDFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	accessToken, err := contextutils.GetOAuthAccessTokenFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	reactEvent := types.ChangeStatus{}

	if err := ctx.Bind(&reactEvent); err != nil {
		return errors.Wrapf(err, "failed to unmarshal content from body")
	}
	response, err := eh.eventUseCase.ChangeStatus(accessToken, reactEvent)
	if err != nil {
		return errors.Wrapf(err, "failed to add attendee for event of telegramUserID=%d", telegramID)
	}
	if response == nil {
		return errors.Wrapf(err, "failed to add attendee for event of telegramUserID=%d, response is nil", telegramID)
	}

	ctx.Response().Header().Set("Content-Type", "application/json")
	return ctx.JSON(http.StatusOK, string(response))
}
