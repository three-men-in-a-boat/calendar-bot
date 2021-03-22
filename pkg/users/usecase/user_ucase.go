package usecase

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/calendar-bot/cmd/config"
	"github.com/calendar-bot/pkg/customerrors"
	"github.com/calendar-bot/pkg/types"
	"github.com/calendar-bot/pkg/users/repository"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"time"
)

const (
	nonceBitsLength = 256
	nonceBase       = 16
)

const stateExpire = 5 * time.Minute

type UserUseCase struct {
	userRepository repository.UserRepository
	conf           *config.App
}

func NewUserUseCase(userRepo repository.UserRepository, conf *config.App) UserUseCase {
	return UserUseCase{
		userRepository: userRepo,
		conf:           conf,
	}
}

type MailruOauthResponseErr struct {
	ErrorName        string `json:"error"`
	ErrorCode        int    `json:"error_code"`
	ErrorDescription string `json:"error_description"`
}

func (o *MailruOauthResponseErr) Error() string {
	return fmt.Sprintf(
		"MailruOauthResponseErr: error='%s', error_code=%d, error_description='%s'",
		o.ErrorName, o.ErrorCode, o.ErrorDescription,
	)
}

type tokenGetResp struct {
	ExpiresInSeconds int64  `json:"expires_in"`
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	*MailruOauthResponseErr
}

type tokenRenewalResp struct {
	ExpiresInSeconds int64  `json:"expires_in"`
	AccessToken      string `json:"access_token"`
	*MailruOauthResponseErr
}

func (uuc *UserUseCase) GenOauthLinkForTelegramID(telegramID int64) (string, error) {
	bigInt, err := rand.Prime(rand.Reader, nonceBitsLength)
	if err != nil {
		return "", errors.WithStack(err)
	}

	state := bigInt.Text(nonceBase)

	if err := uuc.userRepository.SetTelegramUserIDByState(state, telegramID, stateExpire); err != nil {
		return "", errors.Wrapf(err, "cannot insert telegramID=%d in redis, state=%s", telegramID, state)
	}

	link := uuc.generateOAuthLink(state)

	return link, nil

}

func (uuc *UserUseCase) GetTelegramIDByState(state string) (int64, error) {
	return uuc.userRepository.GetTelegramUserIDByState(state)
}

func (uuc *UserUseCase) TelegramCreateAuthentificatedUser(tgUserID int64, mailAuthCode string) (err error) {
	response, err := http.PostForm(
		"https://oauth.mail.ru/token",
		url.Values{
			"code":          []string{mailAuthCode},
			"grant_type":    []string{"authorization_code"},
			"redirect_uri":  []string{uuc.conf.OAuth.RedirectURI},
			"client_id":     []string{uuc.conf.OAuth.ClientID},
			"client_secret": []string{uuc.conf.OAuth.ClientSecret},
		},
	)

	if err != nil {
		return errors.Wrapf(err, "cannot get token for telegramID=%d", tgUserID)
	}

	defer func() {
		err = customerrors.HandleCloser(err, response.Body)
	}()

	var tokenResp tokenGetResp
	if err := json.NewDecoder(response.Body).Decode(&tokenResp); err != nil {
		return errors.Wrap(err, "cannot decode json into tokenGetResp struct")
	}
	if tokenResp.MailruOauthResponseErr != nil {
		return tokenResp.MailruOauthResponseErr
	}
	// nickeskov: in this case http status must be ok
	if response.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("something wrong with token recv and mailru API: http_status='%s'", response.Status))
	}

	accessToken := tokenResp.AccessToken

	response, err = http.Get(fmt.Sprintf("https://oauth.mail.ru/userinfo?access_token=%s", accessToken))

	if err != nil {
		return errors.Wrapf(err, "cannot get accessToken for tgUserID=%d", tgUserID)
	}

	defer func() {
		err = customerrors.HandleCloser(err, response.Body)
	}()

	if response.StatusCode != http.StatusOK {
		return errors.New("something wrong with user info recv: " + response.Status)
	}

	var userInfo map[string]string
	if newErr := json.NewDecoder(response.Body).Decode(&userInfo); err != nil {
		return newErr
	}

	if err, ok := userInfo["error"]; ok {
		return errors.New(err)
	}

	email := userInfo["email"]
	userID := userInfo["id"]

	user := types.User{
		ID:                 0,
		UserID:             userID,
		MailUserEmail:      email,
		MailAccessToken:    tokenResp.AccessToken,
		MailRefreshToken:   tokenResp.RefreshToken,
		MailTokenExpiresIn: time.Now().Add(time.Second * time.Duration(tokenResp.ExpiresInSeconds)),
		TelegramUserId:     tgUserID,
		CreatedAt:          time.Time{},
	}

	return uuc.userRepository.CreateUser(user)
}

func (uuc *UserUseCase) IsUserAuthenticatedByTelegramUserID(telegramID int64) (bool, error) {
	_, err := uuc.userRepository.GetOAuthAccessTokenByTelegramUserID(telegramID)
	if err == nil {
		return true, nil
	}

	if _, ok := err.(repository.OAuthError); ok {
		_, err = uuc.userRepository.GetOAuthRefreshTokenByTelegramUserID(telegramID)
		if err == nil {
			return true, nil
		}
		if _, ok := err.(repository.OAuthError); ok {
			return false, nil
		}
	}

	return false, errors.WithStack(err)
}

func (uuc *UserUseCase) RefreshOAuthTokenByTelegramUserID(telegramID int64) (string, error) {
	refreshToken, err := uuc.userRepository.GetOAuthRefreshTokenByTelegramUserID(telegramID)
	switch {
	case err == repository.UserDoesNotExist || err == repository.UserUnauthorized:
		return "", err
	case err != nil:
		return "", errors.WithStack(err)
	}

	tokenResp, err := uuc.obtainNewOAuthTokenByRefreshToken(refreshToken)
	switch {
	case err != nil:
		return "", errors.WithStack(err)
	case tokenResp.MailruOauthResponseErr != nil:
		return "", tokenResp.MailruOauthResponseErr
	}

	err = uuc.userRepository.SetOAuthAccessTokenByTelegramUserID(
		telegramID,
		tokenResp.AccessToken,
		time.Duration(tokenResp.ExpiresInSeconds)*time.Second,
	)
	if err != nil {
		return "", errors.Wrapf(err, "cannot set new oauth access token for telegramID")
	}

	return tokenResp.AccessToken, nil
}

func (uuc *UserUseCase) obtainNewOAuthTokenByRefreshToken(refreshToken string) (tokenResp tokenRenewalResp, err error) {
	response, err := http.PostForm(
		"https://oauth.mail.ru/token",
		url.Values{
			"grant_type":    []string{"refresh_token"},
			"client_id":     []string{uuc.conf.OAuth.ClientID},
			"refresh_token": []string{refreshToken},
		},
	)
	if err != nil {
		return tokenRenewalResp{}, errors.Wrap(err, "cannot obtain new access token by refresh token")
	}
	defer func() {
		err = customerrors.HandleCloser(err, response.Body)
	}()

	if err := json.NewDecoder(response.Body).Decode(&tokenResp); err != nil {
		return tokenResp, errors.Wrap(err, "cannot decode json into tokenGetResp struct")
	}
	// nickeskov: in this case http status must be ok
	if tokenResp.MailruOauthResponseErr == nil && response.StatusCode != http.StatusOK {
		return tokenRenewalResp{},
			errors.New(
				fmt.Sprintf("something wrong with token recv and mailru API: http_status='%s'", response.Status),
			)
	}

	return tokenResp, nil
}

func (uuc *UserUseCase) SetTelegramIDByState(state string, telegramID int64, expire time.Duration) error {
	return uuc.userRepository.SetTelegramUserIDByState(state, telegramID, expire)
}

func (uuc *UserUseCase) GetOAuthTokenByTelegramID(telegramID int64) (string, error) {
	return uuc.userRepository.GetOAuthAccessTokenByTelegramUserID(telegramID)
}

func (uuc *UserUseCase) generateOAuthLink(state string) string {
	return fmt.Sprintf(
		"https://oauth.mail.ru/login?client_id=%s&response_type=code&scope=%s&redirect_uri=%s&state=%s",
		uuc.conf.OAuth.ClientID,
		uuc.conf.OAuth.Scope,
		uuc.conf.OAuth.RedirectURI,
		state,
	)
}
