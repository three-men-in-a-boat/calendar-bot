package main

import (
	"database/sql"
	_ "database/sql"
	"errors"
	eHandlers "github.com/calendar-bot/pkg/events/handlers"
	eRepo "github.com/calendar-bot/pkg/events/repository"
	eUsecase "github.com/calendar-bot/pkg/events/usecase"
	"github.com/calendar-bot/pkg/types"
	uHandlers "github.com/calendar-bot/pkg/users/handlers"
	uRepo "github.com/calendar-bot/pkg/users/repository"
	uUsecase "github.com/calendar-bot/pkg/users/usecase"
	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"os"
)

type RequestHandlers struct {
	eventHandlers eHandlers.EventHandlers
	userHandlers  uHandlers.UserHandlers
}

func newRequestHandler(db *sql.DB) *RequestHandlers {

	eventStorage := eRepo.NewEventStorage(db)
	eventUseCase := eUsecase.NewEventUseCase(eventStorage)
	eventHandlers := eHandlers.NewEventHandlers(eventUseCase)

	states := types.NewStatesDictionary()
	userStorage := uRepo.NewUserRepository(db)
	userUseCase := uUsecase.NewUserUseCase(userStorage)
	userHandlers := uHandlers.NewUserHandlers(userUseCase, states)

	return &(RequestHandlers{
		eventHandlers: eventHandlers,
		userHandlers:  userHandlers,
	})
}

func connectToDB() (*sql.DB, error) {
	err := godotenv.Load("project.env")
	if err != nil {
		return nil, errors.New("failed to load .env file")
	}
	usernameDB := os.Getenv("POSTGRES_USERNAME")
	passwordDB := os.Getenv("POSTGRES_PASSWORD")
	nameDB := os.Getenv("POSTGRES_DB_NAME")

	connectString := "user=" + usernameDB + " password=" + passwordDB + " dbname=" + nameDB + " sslmode=disable"

	db, err := sql.Open("postgres", connectString)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func initLog() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
}

func main() {
	initLog()

	server := echo.New()

	db, err := connectToDB()
	if err != nil {
		zap.S().Errorf("failed to connect to db, %v", err)
		return
	}

	defer func() {
		err := db.Close()
		if err != nil {
			zap.S().Errorf("failed to close db connection, %v", err)
		}
	}()

	allHandler := newRequestHandler(db)

	allHandler.eventHandlers.InitHandlers(server)
	allHandler.userHandlers.InitHandlers(server)

	server.Logger.Fatal(server.Start(":8080"))
}
