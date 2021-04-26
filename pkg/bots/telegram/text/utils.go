package text

import "fmt"

const (
	errorTextDev = "Что-то пошло не так ... \nОшибка: %s"
	errorTextProd = "Извини, что-то пошло не так - меня обязательно скоро починят :)"
)

func Error(err string) string  {
	return fmt.Sprintf(errorTextDev, err)
}
