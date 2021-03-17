package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/calendar-bot/pkg/types"
	"github.com/calendar-bot/pkg/users/repository"
	"net/http"
	"net/url"
	"time"
)

const RedirectURI = "https://calendarbot.xyz/api/v1/oauth"
const ClientID = "885a013d102b40c7a46a994bc49e68f1"
const ClientSecret = "e16021defba34d869f0e1cfe7461ef2d"

type UserUseCase struct {
	userRepository repository.UserStorage
}

func NewUserUseCase(userRepo repository.UserStorage) UserUseCase {
	return UserUseCase{
		userRepository: userRepo,
	}
}

type tokenGetResp struct {
	ExpiresIn    int64  `json:"expires_in"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

func (uuc UserUseCase) TelegramCreateUser(tgUserID int64, mailAuthCode string) (err error) {
	fmt.Println("tgUserID =", tgUserID)
	fmt.Println("mailAuthCode =", mailAuthCode)

	response, err := http.PostForm(
		"https://oauth.mail.ru/token",
		url.Values{
			"code":          []string{mailAuthCode},
			"grant_type":    []string{"authorization_code"},
			"redirect_uri":  []string{RedirectURI},
			"client_id":     []string{ClientID},
			"client_secret": []string{ClientSecret},
		},
	)

	fmt.Println("response from PosrFrom", response)

	if err != nil {
		println(err.Error())
		return err
	}

	defer func() {
		if response.Body != nil {
			if newErr := response.Body.Close(); newErr != nil {
				println(err.Error())
				err = newErr
			}
		}
	}()

	if response.StatusCode != http.StatusOK {
		return errors.New("something wrong with token recv: " + response.Status)
	}

	var tokenResp tokenGetResp

	if newErr := json.NewDecoder(response.Body).Decode(&tokenResp); err != nil {
		return newErr
	}

	println("access token =", tokenResp.AccessToken)
	println("refresh token =", tokenResp.RefreshToken)
	println("expired in =", tokenResp.ExpiresIn)

	accessToken := tokenResp.AccessToken

	response, err = http.Get("https://oauth.mail.ru/userinfo?access_token=" + accessToken)

	if err != nil {
		println(err.Error())
		return err
	}

	defer func() {
		if newErr := response.Body.Close(); newErr != nil {
			println(err)
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

	fmt.Println(userInfo)

	email := userInfo["email"]
	userID := userInfo["id"]

	user := types.User{
		ID:                 0,
		UserID:             userID,
		MailUserEmail:      email,
		MailAccessToken:    tokenResp.AccessToken,
		MailRefreshToken:   tokenResp.RefreshToken,
		MailTokenExpiresIn: time.Now().Add(time.Second * time.Duration(tokenResp.ExpiresIn)),
		TelegramUserId:     tgUserID,
		CreatedAt:          time.Time{},
	}

	fmt.Println(user)

	return uuc.userRepository.CreateUser(user)
}
