package repository

import "database/sql"

type EventRepository struct {
	storage *sql.DB
}

func NewEventStorage(db *sql.DB) EventRepository {
	return EventRepository{storage: db}

}
