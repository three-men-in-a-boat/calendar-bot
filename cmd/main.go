package main

import (
	"database/sql"
	_ "database/sql"
	"github.com/asaskevich/govalidator"
	"github.com/calendar-bot/cmd/config"
	teleHandlers "github.com/calendar-bot/pkg/bots/telegram/handlers"
	eRepo "github.com/calendar-bot/pkg/events/repository"
	eUsecase "github.com/calendar-bot/pkg/events/usecase"
	"github.com/calendar-bot/pkg/middlewares"
	"github.com/calendar-bot/pkg/types"
	uHandlers "github.com/calendar-bot/pkg/users/handlers"
	uRepo "github.com/calendar-bot/pkg/users/repository"
	uUsecase "github.com/calendar-bot/pkg/users/usecase"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	tb "gopkg.in/tucnak/telebot.v2"
)

type RequestHandlers struct {
	userHandlers  uHandlers.UserHandlers
	baseHandlers teleHandlers.BaseHandlers
}

func newRequestHandler(db *sql.DB, client *redis.Client, conf *config.App) RequestHandlers {

	states := types.NewStatesDictionary()
	userStorage := uRepo.NewUserRepository(db, client)
	userUseCase := uUsecase.NewUserUseCase(userStorage, conf)
	userHandlers := uHandlers.NewUserHandlers(userUseCase, states, conf)

	eventStorage := eRepo.NewEventStorage(db)
	eventUseCase := eUsecase.NewEventUseCase(eventStorage)

	baseHandlers := teleHandlers.NewBaseHandlers(eventUseCase, userUseCase)

	return RequestHandlers{
		userHandlers:  userHandlers,
		baseHandlers: baseHandlers,
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

	webhook := &tb.Webhook{
		Listen:   appConf.BotAddress,
		Endpoint: &tb.WebhookEndpoint{PublicURL: appConf.BotWebhookUrl},
	}

	botSettings := tb.Settings{
		Token:  appConf.BotToken,
		Poller: webhook,
	}

	if appConf.Environment == "dev" {
		botSettings.Verbose = true
	}

	bot, err := tb.NewBot(botSettings)

	server := echo.New()

	db, err := config.ConnectToDB(&appConf.DB)
	if err != nil {
		zap.S().Fatalf("failed to connect to db, %v", err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			zap.S().Errorf("failed to close db connection, %v", err)
		}
	}()

	redisClient, err := config.ConnectToRedis(&appConf.Redis)
	if err != nil {
		zap.S().Fatalf("failed to connect to redis, %v", err)
	}

	allHandler := newRequestHandler(db, redisClient, &appConf)

	server.Use(middlewares.LogErrorMiddleware)

	allHandler.userHandlers.InitHandlers(server)
	allHandler.baseHandlers.InitHandlers(bot)

	go func() { server.Logger.Fatal(server.Start(appConf.Address)) }()

	bot.Start()
}
