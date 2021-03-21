package usecase

import "github.com/calendar-bot/pkg/events/repository"

type EventUseCase struct {
	eventStorage repository.EventRepository
}

func NewEventUseCase(eventStor repository.EventRepository) EventUseCase {
	return EventUseCase{
		eventStorage: eventStor,
	}
}
