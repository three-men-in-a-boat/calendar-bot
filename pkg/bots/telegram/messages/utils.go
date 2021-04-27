package messages

import (
	"fmt"
	"github.com/calendar-bot/cmd/config"
	"os"
)

const (
	errorTextDev = "Что-то пошло не так ... \nОшибка: %s"
	errorTextProd = "Прошу прощения,но что-то пошло не так - меня обязательно скоро починят :)"
	errorReport = "Сообщить об ошибке"
)

func MessageUnexpectedError(err string) string  {
	if os.Getenv(config.EnvAppEnvironment) == config.AppEnvironmentDev {
		return fmt.Sprintf(errorTextDev, err)
	} else {
		return errorTextProd
	}
}

func GetMessageReportBug() string {
	return errorReport
}