package storage


import "database/sql"

type UserStorage struct {
	storage *sql.DB
}

func NewUserStorage(db *sql.DB) UserStorage {
	return UserStorage{storage: db}

}
