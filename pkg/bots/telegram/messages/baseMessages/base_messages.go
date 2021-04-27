package baseMessages

import (
	"fmt"
	"github.com/calendar-bot/pkg/types"
)

const (
	startNoRegText     = "Добрый день, пожалуйста авторизуйтесь для начала работы с календарем"
	startRegButtonText = "Войти в аккаунт mail.ru"
	startRegText       = "Здравствуйте, %s! Вы успешно авторизовались в телеграм боте ассистент календаря " +
		"Mail.ru.Теперь вы можете начать пользоваться ботом. Чтобы узнать какие функции доступны в боте - " +
		"воспользуйтесь командой /help"
)

const (
	helpInfoText = "Это бот для работы с календарем mail.ru. Сейчас доступны следующие команды"
)

const (
	aboutInfoText = "Данный бот - это бот ассистент для калнедаря mail.ru. Для просмотра возможностей бота " +
		"воспользуйтесь командой /help. \n\n" +
		"<b>Проект разработан командой Технопарка \"Трое в лодке не считая дебага\"</b>"
)

func StartNoRegText() string {
	return startNoRegText
}

func StartRegButtonText() string {
	return startRegButtonText
}

func StartRegText(info types.MailruUserInfo) string {
	return fmt.Sprintf(startRegText, info.Name)
}

func HelpInfoText() string {
	return helpInfoText
}

func AboutText() string {
	return aboutInfoText
}
