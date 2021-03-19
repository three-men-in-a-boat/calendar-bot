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
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

func connectToDB(conf *config.App) (*sql.DB, error) {
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

func initLog() error {
	logConfig := config.LoadLogConfig()

	var zapLoglevel zapcore.Level
	switch logConfig.Level {
	case config.LogLevelDebug:
		zapLoglevel = zapcore.DebugLevel
	case config.LogLevelInfo:
		zapLoglevel = zapcore.InfoLevel
	case config.LogLevelWarn:
		zapLoglevel = zapcore.WarnLevel
	case config.LogLevelError:
		zapLoglevel = zapcore.ErrorLevel
	case config.LogLevelDevPanic:
		zapLoglevel = zapcore.DPanicLevel
	case config.LogLevelPanic:
		zapLoglevel = zapcore.PanicLevel
	case config.LogLevelFatal:
		zapLoglevel = zapcore.FatalLevel
	default:
		zapLoglevel = zapcore.InfoLevel
	}

	var zapLogConfig zap.Config
	switch logConfig.Type {
	case config.LogTypeDev:
		zapLogConfig = zap.NewDevelopmentConfig()
	case config.LogTypeProd:
		zapLogConfig = zap.NewProductionConfig()
	default:
		zapLogConfig = zap.NewProductionConfig()
	}

	zapLogConfig.Level.SetLevel(zapLoglevel)

	logger, err := zapLogConfig.Build()
	if err != nil {
		return errors.WithStack(err)
	}
	zap.ReplaceGlobals(logger)

	return nil
}

func init() {
	// nickeskov: error != nil if no .env file
	dotenvErr := godotenv.Load()

	if err := initLog(); err != nil {
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

	db, err := connectToDB(&appConf)
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
