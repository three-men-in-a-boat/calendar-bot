package text

const (
	startNoRegText     = "Добрый день, пожалуйста авторизуйтесь для начала работы с календарем"
	startRegButtonText = "Войти в аккаунт mail.ru"
	startRegText = "Вы уже авторизованы"
)

func StartNoRegText() string {
	return startNoRegText
}

func StartRegButtonText() string {
	return startRegButtonText
}

func StartRegText() string {
	return startRegText
}
