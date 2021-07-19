package baseMessages

import (
	"fmt"
	"github.com/calendar-bot/pkg/services/oauth"
)

const (
	startNoRegText = "Добрый день, пожалуйста авторизуйтесь для начала работы с календарем. <b>Обратите внимание, что ваши " +
		"данные будут отправлены во внешний сервис!</b>\n\nПосле того, как " +
		"вы вернетесь сюда после авторизации - нажмите на кнопку, которая появится вместо поля для ввода сообщения," +
		" или же воспользуйтесь командой /start"
	startRegButtonText = "Войти в аккаунт mail.ru"
	startRegText       = "Здравствуйте, %s! Вы успешно авторизовались в телеграм боте ассистент календаря " +
		"Mail.ru.Теперь вы можете начать пользоваться ботом. Чтобы узнать какие функции доступны в боте - " +
		"воспользуйтесь командой /help"
)

const (
	helpInfoText = "Это бот для работы с календарем mail.ru. Сейчас доступны следующие команды:\n\n" +
		"/today - просмотр информации в календаре за <b>сегодня</b>\n/next - получение информации о " +
		"<b>ближайшем событии</b>\n" +
		"/date - просмотр информации в календаре за <b>выбранную дату</b>\n/create - <b>создание</b> события в " +
		"календаре в одиночном " +
		"и групповом чате. Поиск удобного времени для всех участников в групповом чате <b> - для работы каждом участнику" +
		" необходимо авторизоваться в боте в личном чате с ним</b>\n/about - информация о команде разработке"
)

const (
	aboutInfoText = "Данный бот - это бот ассистент для калнедаря mail.ru. Для просмотра возможностей бота " +
		"воспользуйтесь командой /help \n\n" +
		"<b>Проект разработан командой Технопарка \"Трое в лодке не считая дебага\"</b>"
)

func StartNoRegText() string {
	return startNoRegText
}

func StartRegButtonText() string {
	return startRegButtonText
}

func StartRegText(info oauth.UserInfoResponse) string {
	return fmt.Sprintf(startRegText, info.Name)
}

func HelpInfoText() string {
	return helpInfoText
}

func AboutText() string {
	return aboutInfoText
}
