package storage

import "database/sql"

type EventStorage struct {
	storage *sql.DB
}

func NewEventStorage(db *sql.DB) EventStorage {
	return EventStorage{storage: db}

}
