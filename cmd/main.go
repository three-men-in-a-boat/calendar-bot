package main

import (
	"database/sql"
	_ "database/sql"
	"github.com/asaskevich/govalidator"
	teleHandlers "github.com/calendar-bot/pkg/bots/telegram/handlers"
	"github.com/calendar-bot/pkg/config"
	eRepo "github.com/calendar-bot/pkg/events/repository"
	eUsecase "github.com/calendar-bot/pkg/events/usecase"
	"github.com/calendar-bot/pkg/log"
	"github.com/calendar-bot/pkg/middlewares"
	"github.com/calendar-bot/pkg/services/db"
	"github.com/calendar-bot/pkg/services/oauth"
	redisService "github.com/calendar-bot/pkg/services/redis"
	uHandlers "github.com/calendar-bot/pkg/users/handlers"
	uRepo "github.com/calendar-bot/pkg/users/repository"
	uUsecase "github.com/calendar-bot/pkg/users/usecase"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	echoProm "github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	tb "gopkg.in/tucnak/telebot.v2"
)

type RequestHandlers struct {
	userHandlers             uHandlers.UserHandlers
	telegramBaseHandlers     teleHandlers.BaseHandlers
	telegramCalendarHandlers teleHandlers.CalendarHandlers
}

func newRequestHandler(db *sql.DB, client *redis.Client, botClient *redis.Client, conf *config.AppConfig) RequestHandlers {
	oauthService := oauth.NewService(&conf.OAuth, client)

	userStorage := uRepo.NewUserRepository(db)
	userUseCase := uUsecase.NewUserUseCase(userStorage, &oauthService)
	userHandlers := uHandlers.NewUserHandlers(userUseCase)

	eventStorage := eRepo.NewEventStorage(db)
	eventUseCase := eUsecase.NewEventUseCase(eventStorage)

	teleBaseHandlers := teleHandlers.NewBaseHandlers(eventUseCase, userUseCase, conf.ParseAddress)
	teleCalendarHandler := teleHandlers.NewCalendarHandlers(eventUseCase, userUseCase, botClient, conf.ParseAddress)

	return RequestHandlers{
		userHandlers:             userHandlers,
		telegramBaseHandlers:     teleBaseHandlers,
		telegramCalendarHandlers: teleCalendarHandler,
	}
}

func init() {
	// nickeskov: error != nil if no .env file
	dotenvErr := godotenv.Load()

	if err := log.InitLog(); err != nil {
		// nickeskov: in this case this we can do only one thing - start panicking. Or maybe use log.Fatal(...)
		panic(err)
	}

	if dotenvErr != nil {
		zap.S().Infof("No .env file found: %v", dotenvErr)
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

	if appConf.Environment == config.AppEnvironmentDev {
		botSettings.Verbose = true
	}

	bot, err := tb.NewBot(botSettings)
	if err != nil {
		zap.S().Fatalf("failed to start bot, %v", err)
	}

	server := echo.New()

	dbConnection, err := db.ConnectToPostgresDB(&appConf.DB)
	if err != nil {
		zap.S().Fatalf("failed to connect to db, %v", err)
	}
	defer func() {
		err := dbConnection.Close()
		if err != nil {
			zap.S().Errorf("failed to close db connection, %v", err)
		}
	}()

	redisClient, err := redisService.ConnectToRedis(&appConf.Redis)
	if err != nil {
		zap.S().Fatalf("failed to connect to redis, %v", err)
	}

	botRedisClient, err := redisService.ConnectToRedis(&appConf.BotRedis)
	if err != nil {
		zap.S().Fatalf("failed to connect to bot redis, %v", err)
	}

	allHandler := newRequestHandler(dbConnection, redisClient, botRedisClient, &appConf)

	server.Use(middlewares.LogErrorMiddleware)

	allHandler.userHandlers.InitHandlers(server)
	allHandler.telegramBaseHandlers.InitHandlers(bot)
	allHandler.telegramCalendarHandlers.InitHandlers(bot)

	echoProm.NewPrometheus("http", nil).Use(server)

	go func() { zap.S().Fatal(server.Start(appConf.Address)) }()

	bot.Start()
}
