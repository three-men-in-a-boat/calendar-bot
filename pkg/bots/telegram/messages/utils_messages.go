package messages

import (
	"fmt"
	config2 "github.com/calendar-bot/pkg/config"
	"os"
)

const (
	errorTextDev = "Что-то пошло не так ... \nОшибка: %s"
	errorTextProd = "Прошу прощения,но что-то пошло не так - меня обязательно скоро починят :)"
	errorReport = "Сообщить об ошибке"

	errorAuthDev = "Ошибка авторизации: %s"
	errorAuthProd = "Похоже, что вы не авторизованы - войдите в аккаунт mail.ru с помощью команды /start"
)

func MessageUnexpectedError(err string) string  {
	if os.Getenv(config2.EnvAppEnvironment) == config2.AppEnvironmentDev {
		return fmt.Sprintf(errorTextDev, err)
	} else {
		return errorTextProd
	}
}

func MessageAuthError(err string) string {
	if os.Getenv(config2.EnvAppEnvironment) == config2.AppEnvironmentDev {
		return fmt.Sprintf(errorAuthDev, err)
	} else {
		return errorAuthProd
	}
}

func GetMessageReportBug() string {
	return errorReport
}
