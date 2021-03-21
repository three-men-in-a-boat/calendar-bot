package main

import (
	"database/sql"
	_ "database/sql"
	"github.com/asaskevich/govalidator"
	"github.com/calendar-bot/cmd/config"
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
)

type RequestHandlers struct {
	eventHandlers eHandlers.EventHandlers
	userHandlers  uHandlers.UserHandlers
}

func newRequestHandler(db *sql.DB, conf *config.App) RequestHandlers {

	eventStorage := eRepo.NewEventStorage(db)
	eventUseCase := eUsecase.NewEventUseCase(eventStorage)
	eventHandlers := eHandlers.NewEventHandlers(eventUseCase)

	states := types.NewStatesDictionary()
	userStorage := uRepo.NewUserRepository(db)
	userUseCase := uUsecase.NewUserUseCase(userStorage, conf)
	userHandlers := uHandlers.NewUserHandlers(userUseCase, states, conf)

	return RequestHandlers{
		eventHandlers: eventHandlers,
		userHandlers:  userHandlers,
	}
}

func init() {
	// nickeskov: error != nil if no .env file
	dotenvErr := godotenv.Load()

	if err := config.InitLog(); err != nil {
		// nickeskov: in this case this we can do only one thing - start panicking. Or maybe use log.Fatal(...)
		panic(err)
	}

	if dotenvErr != nil {
		zap.S().Info("No .env file found: %v", dotenvErr)
	}

	govalidator.SetFieldsRequiredByDefault(true)
}

func main() {
	appConf, err := config.LoadAppConfig()
	if err != nil {
		zap.S().Fatalf("cannot load APP config: %v", err)
	}
	server := echo.New()

	db, err := config.ConnectToDB(&appConf)
	if err != nil {
		zap.S().Fatalf("failed to connect to db, %v", err)
	}

	defer func() {
		err := db.Close()
		if err != nil {
			zap.S().Errorf("failed to close db connection, %v", err)
		}
	}()

	allHandler := newRequestHandler(db, &appConf)

	allHandler.eventHandlers.InitHandlers(server)
	allHandler.userHandlers.InitHandlers(server)

	server.Logger.Fatal(server.Start(appConf.Address))
}
