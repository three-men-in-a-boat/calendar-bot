package config

import (
	"database/sql"
	_ "database/sql"
	"github.com/asaskevich/govalidator"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"os"
	"strconv"
)

const (
	EnvDBName               = "DB_NAME"
	EnvDBUsername           = "DB_USERNAME"
	EnvDBPassword           = "DB_PASSWORD"
	EnvDBMaxOpenConnections = "DB_MAX_OPEN_CONNECTIONS"
)

type DB struct {
	Name               string `valid:"-"`
	Username           string `valid:"-"`
	Password           string `valid:"-"`
	MaxOpenConnections int    `valid:"-"`
}

func LoadDBConfig() (DB, error) {
	name := os.Getenv(EnvDBName)
	username := os.Getenv(EnvDBUsername)
	password := os.Getenv(EnvDBPassword)

	maxOpenConnections := 10
	if maxOpenConnectionsStr := os.Getenv(EnvDBMaxOpenConnections); maxOpenConnectionsStr != "" {
		value, err := strconv.Atoi(maxOpenConnectionsStr)
		if err != nil {
			return DB{},
				errors.WithMessagef(err, "failed to parse %s environment variable as int", EnvDBMaxOpenConnections)
		}
		maxOpenConnections = value
	}

	conf := DB{
		Name:               name,
		Username:           username,
		Password:           password,
		MaxOpenConnections: maxOpenConnections,
	}

	if _, err := govalidator.ValidateStruct(&conf); err != nil {
		return DB{}, errors.WithMessagef(err, "govalidator validation error")
	}

	return conf, nil
}

func (db *DB) ToEnv() map[string]string {
	return map[string]string{
		EnvDBName:               db.Name,
		EnvDBUsername:           db.Username,
		EnvDBPassword:           db.Password,
		EnvDBMaxOpenConnections: strconv.Itoa(db.MaxOpenConnections),
	}
}


func ConnectToDB(conf *App) (*sql.DB, error) {
	nameDB := conf.DB.Name
	usernameDB := conf.DB.Username
	passwordDB := conf.DB.Password

	connectString := "user=" + usernameDB + " password=" + passwordDB + " dbname=" + nameDB + " sslmode=disable"

	db, err := sql.Open("postgres", connectString)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if err := db.Ping(); err != nil {
		return nil, errors.WithStack(err)
	}
	return db, nil
}