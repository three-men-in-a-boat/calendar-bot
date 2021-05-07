package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

func ConnectToPostgresDB(conf *Config) (*sql.DB, error) {
	if conf == nil {
		return nil, errors.New("nil db config")
	}

	connectionString := fmt.Sprintf(
		"user=%s password=%s dbname=%s sslmode=disable",
		conf.Username,
		conf.Password,
		conf.Name,
	)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if pingErr := db.Ping(); pingErr != nil {
		return nil, errors.WithStack(pingErr)
	}
	return db, nil
}
