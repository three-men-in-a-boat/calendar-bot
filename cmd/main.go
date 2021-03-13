package main

import (
	"database/sql"
	_ "database/sql"
	"github.com/calendar-bot/pkg/events/handlers"
	"github.com/calendar-bot/pkg/events/storage"
	"github.com/calendar-bot/pkg/events/usecase"
	"github.com/labstack/echo"
	"go.uber.org/zap"
	"os"
)

type RequestHandlers struct {
	eventHandlers handlers.EventHandlers
}

func newRequestHandler(db *sql.DB) *RequestHandlers {

	eventStorage := storage.NewEventStorage(db)
	eventUseCase := usecase.NewEventUseCase(eventStorage)
	eventHandlers := handlers.NewEventHandlers(eventUseCase)

	return &(RequestHandlers{
		eventHandlers: eventHandlers,
	})
}

func connectToDB(server *echo.Echo) (*sql.DB, error) {
	usernameDB := "main"
	passwordDB := "main"
	nameDB := "mainnet"
	connectString := "user=" + usernameDB + " password=" + passwordDB + " dbname=" + nameDB + " sslmode=disable"

	db, err := sql.Open("postgres", connectString)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func main() {
	server := echo.New()

	db, err := connectToDB(server)
	if err != nil {
		zap.S().Fatalf("failed to connect to db, %v", err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			zap.S().Errorf("failed to close db connection, %v", err)
		}
	}()

	allHandler := newRequestHandler(db)

	allHandler.eventHandlers.InitHandlers(server)

	server.Logger.Fatal(server.Start(":8080"))
}