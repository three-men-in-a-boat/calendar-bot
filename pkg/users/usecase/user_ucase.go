package usecase

import (
	"encoding/json"
	"errors"
	"github.com/calendar-bot/pkg/models"
	"github.com/calendar-bot/pkg/users/storage"
	"net/http"
	"net/url"
	"time"
)

const RedirectURI = "https://calendarbot.xyz/api/v1/oaut"

type UserUseCase struct {
	userStorage storage.UserStorage
}

func NewUserUseCase(userStor storage.UserStorage) UserUseCase {
	return UserUseCase{
		userStorage: userStor,
	}
}

type tokenGetResp struct {
	expires_in    int64
	access_token  string
	refresh_token string
}

func (uuc UserUseCase) TelegramCreateUser(tgUserID int64, mailAuthCode string) (err error) {

	response, err := http.PostForm(
		"https://oauth.mail.ru/token",
		url.Values{
			"code":         []string{mailAuthCode},
			"grant_type":   []string{"authorization_code"},
			"redirect_uri": []string{RedirectURI},
		},
	)

	if err != nil {
		return err
	}

	defer func() {
		if newErr := response.Body.Close(); newErr != nil {
			err = newErr
		}
	}()

	if response.StatusCode != http.StatusOK {
		return errors.New("something wrong with token recv: " + response.Status)
	}

	var tokenResp tokenGetResp

	if newErr := json.NewDecoder(response.Body).Decode(&tokenResp); err != nil {
		return newErr
	}

	accessToken := tokenResp.access_token

	response, err = http.Get("https://oauth.mail.ru/userinfo?access_token=" + accessToken)

	if err != nil {
		return err
	}

	defer func() {
		if newErr := response.Body.Close(); newErr != nil {
			err = newErr
		}
	}()

	if response.StatusCode != http.StatusOK {
		return errors.New("something wrong with user info recv: " + response.Status)
	}


	var userInfo map[string]string
	if newErr := json.NewDecoder(response.Body).Decode(&userInfo); err != nil {
		return newErr
	}

	email := userInfo["email"]
	userID := userInfo["id"]

	user := models.User{
		ID:                 0,
		UserID:             userID,
		MailUserEmail:      email,
		MailAccessToken:    tokenResp.access_token,
		MailRefreshToken:   tokenResp.refresh_token,
		MailTokenExpiresIn: time.Now().Add(time.Second * time.Duration(tokenResp.expires_in)),
		TelegramUserId:     tgUserID,
		CreatedAt:          time.Time{},
	}

	return uuc.userStorage.CreateUser(user)
}
