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

const stateExpire = 15 * time.Minute

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

type tokenGetResp struct {
	ExpiresInSeconds int64  `json:"expires_in"`
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	*types.MailruAPIResponseErr
}

type tokenRenewalResp struct {
	ExpiresInSeconds int64  `json:"expires_in"`
	AccessToken      string `json:"access_token"`
	*types.MailruAPIResponseErr
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

func (uuc *UserUseCase) GetOrRefreshOAuthAccessTokenByTelegramUserID(telegramID int64) (string, error) {
	oAuthToken, err := uuc.GetOAuthAccessTokenByTelegramUserID(telegramID)
	switch {
	case err == repository.OAuthAccessTokenDoesNotExist:
		oAuthToken, err = uuc.RefreshOAuthTokenByTelegramUserID(telegramID)
		if err != nil {
			switch concreteErr := err.(type) {
			case repository.OAuthError, repository.UserEntityError:
				return "", concreteErr
			default:
				return "", errors.WithStack(concreteErr)
			}
		}
	case err != nil:
		return "", errors.WithStack(err)
	}

	return oAuthToken, nil
}

func (uuc *UserUseCase) GetTelegramIDByState(state string) (int64, error) {
	return uuc.userRepository.GetTelegramUserIDByState(state)
}

func (uuc *UserUseCase) TelegramCreateAuthenticatedUser(tgUserID int64, mailAuthCode string) (err error) {
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
	if tokenResp.MailruAPIResponseErr != nil {
		return tokenResp.MailruAPIResponseErr
	}
	// nickeskov: in this case http status must be ok
	if response.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("something wrong with token recv and mailru API: http_status='%s'", response.Status))
	}

	userInfo, err := uuc.GetMailruUserInfo(tokenResp.AccessToken)
	if err != nil {
		return errors.WithStack(err)
	}

	if !userInfo.IsValid() {
		return errors.Wrapf(userInfo.GetError(),
			"mailru oauth userinfo API, response error, telegramUserID=%d", tgUserID)
	}

	user := types.TelegramDBUser{
		MailUserID:       userInfo.ID,
		MailUserEmail:    userInfo.Email,
		MailRefreshToken: tokenResp.RefreshToken,
		TelegramUserId:   tgUserID,
	}

	if err := uuc.userRepository.CreateUser(user); err != nil {
		return errors.Wrapf(err, "db error, failed to authenticated user, telegramUserID=%d", tgUserID)
	}

	expire := time.Duration(tokenResp.ExpiresInSeconds) * time.Second
	if err := uuc.userRepository.SetOAuthAccessTokenByTelegramUserID(tgUserID, tokenResp.AccessToken, expire); err != nil {
		return errors.Wrapf(err, "redis error, failed to set accessToken for telegramUserID=%d", tgUserID)
	}

	return nil
}

func (uuc *UserUseCase) GetMailruUserInfo(accessToken string) (userInfo types.MailruUserInfo, err error) {
	response, err := http.Get(fmt.Sprintf("https://oauth.mail.ru/userinfo?access_token=%s", accessToken))
	if err != nil {
		return types.MailruUserInfo{}, errors.Wrap(err, "failed to send userinfo oauth request")
	}
	defer func() {
		err = customerrors.HandleCloser(err, response.Body)
	}()

	if err := json.NewDecoder(response.Body).Decode(&userInfo); err != nil {
		return types.MailruUserInfo{}, errors.Wrap(err, "cannot decode json into MailruUserInfo struct")
	}

	if userInfo.MailruAPIResponseErr == nil && response.StatusCode != http.StatusOK {
		return types.MailruUserInfo{},
			errors.New(
				fmt.Sprintf("something wrong with token recv and mailru API: http_status='%s'", response.Status),
			)
	}

	if userInfo.MailruAPIResponseErr != nil {
		return types.MailruUserInfo{}, errors.Wrap(err, "error in userinfo oauth response from Mail.ru API")
	}

	return userInfo, nil
}

func (uuc *UserUseCase) IsUserAuthenticatedByTelegramUserID(telegramID int64) (bool, error) {
	_, err := uuc.userRepository.GetOAuthAccessTokenByTelegramUserID(telegramID)
	if err == nil {
		return true, nil
	}
	if err != repository.OAuthAccessTokenDoesNotExist {
		return false, errors.WithStack(err)
	}

	_, err = uuc.userRepository.GetOAuthRefreshTokenByTelegramUserID(telegramID)
	if err == nil {
		// TODO(nickeskov): maybe need RefreshOAuthTokenByTelegramUserID?
		return true, nil
	}
	// nickeskov: if user does not exist in db or exists, but not authorized
	if err == repository.UserDoesNotExist || err == repository.UserUnauthorized {
		return false, nil
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
	case tokenResp.MailruAPIResponseErr != nil:
		return "", tokenResp.MailruAPIResponseErr
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
	if tokenResp.MailruAPIResponseErr == nil && response.StatusCode != http.StatusOK {
		return tokenRenewalResp{},
			errors.New(
				fmt.Sprintf("something wrong with token recv and mailru API: http_status='%s'", response.Status),
			)
	}

	return tokenResp, nil
}

func (uuc *UserUseCase) DeleteLocalAuthenticatedUserByTelegramUserID(telegramID int64) error {
	if err := uuc.userRepository.DeleteOAuthAccessTokenByTelegramUserID(telegramID); err != nil {
		return errors.Wrapf(err, "failed to delete acces token in redis by telegramUserID=%d", telegramID)
	}
	switch err := uuc.userRepository.DeleteUserByTelegramUserID(telegramID); err {
	case nil:
		return nil
	case repository.UserDoesNotExist:
		return err
	default:
		return errors.WithStack(err)
	}
}

func (uuc *UserUseCase) GetOAuthAccessTokenByTelegramUserID(telegramID int64) (string, error) {
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
