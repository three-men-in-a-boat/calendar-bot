package usecase

import "github.com/calendar-bot/pkg/events/storage"

type EventUseCase struct {
	eventStorage storage.EventStorage
}



func NewEventUseCase(eventStor storage.EventStorage) EventUseCase {
	return EventUseCase{
		eventStorage: eventStor,
	}
}
