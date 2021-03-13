package usecase

import "github.com/calendar-bot/pkg/users/storage"

type UserUseCase struct {
	userStorage storage.UserStorage
}



func NewUserUseCase(userStor storage.UserStorage) UserUseCase {
	return UserUseCase{
		userStorage: userStor,
	}
}

