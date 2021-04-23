package text

import "fmt"

const (
	errorText = "Что-то пошло не так ... \nОшибка: %s"
)

func Error(err string) string  {
	return fmt.Sprintf(errorText, err)
}
