package usecase

import (
	"fmt"
	"github.com/calendar-bot/pkg/services/oauth"
	"github.com/calendar-bot/pkg/types"
	"github.com/calendar-bot/pkg/users/repository"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type UserUseCase struct {
	userRepository repository.UserRepository
	oauthService   *oauth.Service
}

func NewUserUseCase(userRepo repository.UserRepository, service *oauth.Service) UserUseCase {
	return UserUseCase{
		userRepository: userRepo,
		oauthService:   service,
	}
}

func (uuc *UserUseCase) GetTelegramBotRedirectURI() string {
	return uuc.oauthService.Config().TelegramBotRedirectURI
}

func (uuc *UserUseCase) GenOauthLinkForTelegramID(telegramUserID int64) (string, error) {
	link, err := uuc.oauthService.GenerateOAuthLinkWithDefaultExpire(telegramUserID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to generate oauth link for telegramUserID=%d", telegramUserID)
	}
	return link.String(), nil

}

func (uuc *UserUseCase) GetTelegramUserIDByState(state string) (int64, error) {
	var telegramUserID int64
	if err := uuc.oauthService.ScanStateValueByStateKey(state, &telegramUserID); err != nil {
		switch err {
		case oauth.StateKeyDoesNotExist:
			return 0, err
		default:
			return 0, errors.Wrap(err, "GetTelegramUserIDByState")
		}
	}
	return telegramUserID, nil
}

func (uuc *UserUseCase) GetOrRefreshOAuthAccessTokenByTelegramUserID(telegramID int64) (token string, err error) {
	timer := prometheus.NewTimer(metricGetOrRefreshOAuthAccessTokenByTelegramUserIDDuration)
	defer func() {
		metricGetOrRefreshOAuthAccessTokenByTelegramUserIDTotalCount.WithLabelValues(metricStatusFromErr(err))
		timer.ObserveDuration()
	}()

	oAuthToken, err := uuc.getOAuthAccessTokenByTelegramUserID(telegramID)
	switch {
	case errors.Is(err, oauth.AccessTokenDoesNotExist):
		oAuthToken, err = uuc.refreshOAuthTokenByTelegramUserID(telegramID)
		if err != nil {
			switch concreteErr := err.(type) {
			case oauth.Error, repository.UserEntityError:
				return "", concreteErr
			default:
				return "", errors.WithStack(concreteErr)
			}
		}
	case err != nil:
		return "", errors.Wrap(err, "GetOrRefreshOAuthAccessTokenByTelegramUserID")
	}
	return oAuthToken, nil
}

func (uuc *UserUseCase) TelegramCreateAuthenticatedUser(tgUserID int64, mailAuthCode string) (err error) {
	timer := prometheus.NewTimer(metricTelegramCreateAuthenticatedUserDuration)
	defer func() {
		metricTelegramCreateAuthenticatedUserTotalCount.WithLabelValues(metricStatusFromErr(err))
		timer.ObserveDuration()
	}()

	tokenResp, err := uuc.oauthService.ObtainTokensFromOAuthHost(mailAuthCode)
	if err != nil {
		switch err.(type) {
		case *oauth.APIResponseErr:
			return err
		default:
			return errors.Wrap(err, "TelegramCreateAuthenticatedUser")
		}
	}

	userInfo, err := uuc.oauthService.GetUserInfo(tokenResp.AccessToken)
	if err != nil {
		switch err.(type) {
		case *oauth.APIResponseErr:
			return err
		default:
			return errors.Wrap(err, "TelegramCreateAuthenticatedUser")
		}
	}

	user := types.TelegramDBUser{
		MailUserID:           userInfo.ID,
		MailUserEmail:        userInfo.Email,
		MailRefreshToken:     tokenResp.RefreshToken,
		TelegramUserId:       tgUserID,
		TelegramUserTimezone: nil,
	}

	if err := uuc.userRepository.CreateUser(user); err != nil {
		return errors.Wrapf(err, "db error, failed to authenticated user, telegramUserID=%d", tgUserID)
	}

	err = uuc.setOAuthAccessTokenByTelegramUserID(
		tgUserID,
		tokenResp.AccessToken,
		time.Duration(tokenResp.ExpiresInSeconds)*time.Second,
	)
	if err != nil {
		return errors.Wrapf(err, "redis error, failed to set accessToken for telegramUserID=%d", tgUserID)
	}

	return nil
}

func (uuc *UserUseCase) GetMailruUserInfo(accessToken string) (oauth.UserInfoResponse, error) {
	response, err := uuc.oauthService.GetUserInfo(accessToken)
	if err != nil {
		return oauth.UserInfoResponse{}, errors.Wrap(err, "GetMailruUserInfo")
	}
	return response, nil
}

func (uuc *UserUseCase) GetTelegramUserTimezoneByTelegramUserID(telegramID int64) (*string, error) {
	tz, err := uuc.userRepository.GetTelegramUserTimezoneByTelegramUserID(telegramID)
	if err != nil {
		switch err.(type) {
		case repository.UserEntityError:
			return nil, err
		default:
			return nil, errors.Wrap(err, "GetTelegramUserTimezoneByTelegramUserID")
		}
	}
	return tz, nil
}

func (uuc *UserUseCase) UpdateTelegramUserTimezoneByTelegramUserID(telegramID int64, timezone *string) error {
	err := uuc.userRepository.UpdateTelegramUserTimezoneByTelegramUserID(telegramID, timezone)
	if err != nil {
		switch err.(type) {
		case repository.UserEntityError:
			return err
		default:
			return errors.Wrap(err, "UpdateTelegramUserTimezoneByTelegramUserID")
		}
	}
	return nil
}

func (uuc *UserUseCase) GetUserEmailByTelegramUserID(telegramID int64) (email string, err error) {
	timer := prometheus.NewTimer(metricGetUserEmailByTelegramUserIDDuration)
	defer func() {
		metricGetUserEmailByTelegramUserIDCount.WithLabelValues(metricStatusFromErr(err))
		timer.ObserveDuration()
	}()

	email, err = uuc.userRepository.GetUserEmailByTelegramUserID(telegramID)
	if err != nil {
		switch err.(type) {
		case repository.UserEntityError:
			return "", err
		default:
			return "", errors.Wrap(err, "GetUserEmailByTelegramUserID")
		}
	}
	return email, nil
}

func (uuc *UserUseCase) TryGetUsersEmailsByTelegramUserIDs(telegramIDs []int64) (emails []string, err error) {
	timer := prometheus.NewTimer(metricTryGetUsersEmailsByTelegramUserIDsDuration)
	defer func() {
		metricTryGetUsersEmailsByTelegramUserIDsCount.WithLabelValues(metricStatusFromErr(err))
		timer.ObserveDuration()
	}()

	emails, err = uuc.userRepository.TryGetUsersEmailsByTelegramUserIDs(telegramIDs)
	if err != nil {
		switch err.(type) {
		case repository.UserEntityError:
			return nil, err
		default:
			return nil, errors.Wrap(err, "TryGetUsersEmailsByTelegramUserIDs")
		}
	}
	return emails, nil
}

func (uuc *UserUseCase) IsUserAuthenticatedByTelegramUserID(telegramID int64) (isAuth bool, err error) {
	timer := prometheus.NewTimer(metricIsUserAuthenticatedByTelegramUserIDDuration)
	defer func() {
		metricIsUserAuthenticatedByTelegramUserIDCount.WithLabelValues(metricStatusFromErr(err))
		timer.ObserveDuration()
	}()

	// nickeskov: validate user authentication if access token stored in local tokens storage
	switch accessToken, err := uuc.getOAuthAccessTokenByTelegramUserID(telegramID); {
	case err == nil:
		isValid, err := uuc.isValidOAuthAccessToken(accessToken)
		if err != nil {
			return false, errors.Wrapf(err,
				"failed to check is user authenticated by telegram user ID %d", telegramID)
		}
		if !isValid {
			// nickeskov: user revoked his token
			if err := uuc.DeleteLocalAuthenticatedUserByTelegramUserID(telegramID); err != nil {
				return false, errors.Wrapf(err,
					"failed to delete user who revoked his token, telegram user ID %d", telegramID)
			}
			return false, nil
		}
		// nickeskov: token is valid and user is authenticated
		return true, nil
	case errors.Is(err, oauth.AccessTokenDoesNotExist):
		// nickeskov: going to the next validation step
	default:
		// nickeskov: unknown or unexpected error
		return false, errors.Wrapf(err,
			"failed to get saved oauth access token for telegram user ID %d", telegramID)
	}

	// nickeskov: validate user authentication if access token was outdated
	switch _, err := uuc.refreshOAuthTokenByTelegramUserID(telegramID); {
	case err == nil:
		// nickeskov: we've just refreshed access token and user is definitely authenticated
		return true, nil
	case errors.Is(err, repository.UserDoesNotExist):
		// nickeskov: user isn't authenticated because he doesn't exist
		return false, nil
	case isBadTokenError(err):
		// nickeskov: user revoked his token
		if err := uuc.DeleteLocalAuthenticatedUserByTelegramUserID(telegramID); err != nil {
			return false, errors.Wrapf(err,
				"failed to delete user who revoked his token, telegram user ID %d", telegramID)
		}
		return false, nil
	default:
		// nickeskov: unknown or unexpected error
		return false, errors.Wrapf(err,
			"failed to refresh oauth token by telegram user ID %d", telegramID)
	}
}

func (uuc *UserUseCase) isValidOAuthAccessToken(accessToken string) (bool, error) {
	// nickeskov: trying to fetch UserInfo by token to check access token validity and revoke status
	_, err := uuc.oauthService.GetUserInfo(accessToken)
	switch {
	case err == nil:
		return true, nil
	case isBadTokenError(err):
		return false, nil
	default:
		return false, errors.Wrap(err, "failed to validate access token")
	}
}

func isBadTokenError(err error) bool {
	if err == nil {
		return false
	}
	if err, ok := err.(*oauth.APIResponseErr); ok {
		switch err.ErrorCode {
		case oauth.APIResponseErrTokenNotFound:
			return true
		case oauth.APIResponseErrInvalidRequest:
			// TODO(nickeskov): I'm not sure in this case.
			return true
		}
	}
	return false
}

func (uuc *UserUseCase) refreshOAuthTokenByTelegramUserID(telegramID int64) (string, error) {
	refreshToken, err := uuc.userRepository.GetOAuthRefreshTokenByTelegramUserID(telegramID)
	switch {
	case errors.Is(err, repository.UserDoesNotExist):
		return "", err
	case err != nil:
		return "", errors.WithStack(err)
	}

	tokenResp, err := uuc.oauthService.RenewAccessTokenByRefreshToken(refreshToken)
	if err != nil {
		switch err.(type) {
		case *oauth.APIResponseErr:
			return "", err
		default:
			return "", errors.Wrapf(err,
				"failed to RenewAccessTokenByRefreshToken for telegram user ID %d", telegramID)
		}
	}

	err = uuc.setOAuthAccessTokenByTelegramUserID(
		telegramID,
		tokenResp.AccessToken,
		time.Duration(tokenResp.ExpiresInSeconds)*time.Second,
	)
	if err != nil {
		return "", errors.Wrapf(err, "cannot set new oauth access token for telegramID")
	}

	return tokenResp.AccessToken, nil
}

func (uuc *UserUseCase) DeleteLocalAuthenticatedUserByTelegramUserID(telegramID int64) (err error) {
	timer := prometheus.NewTimer(metricDeleteLocalAuthenticatedUserByTelegramUserIDDuration)
	defer func() {
		metricDeleteLocalAuthenticatedUserByTelegramUserIDCount.WithLabelValues(metricStatusFromErr(err))
		timer.ObserveDuration()
	}()

	if err := uuc.delOAuthAccessTokenByTelegramUserID(telegramID); err != nil {
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

func (uuc *UserUseCase) getOAuthAccessTokenByTelegramUserID(telegramUserID int64) (string, error) {
	key := createRedisKeyForTelegramUserID(telegramUserID)
	return uuc.oauthService.GetAccessTokenByKey(key)
}

func (uuc *UserUseCase) setOAuthAccessTokenByTelegramUserID(
	telegramUserID int64,
	accessToken string,
	expire time.Duration,
) error {
	key := createRedisKeyForTelegramUserID(telegramUserID)
	return uuc.oauthService.SetAccessTokenByKey(key, accessToken, expire)
}

func (uuc *UserUseCase) delOAuthAccessTokenByTelegramUserID(telegramUserID int64) error {
	key := createRedisKeyForTelegramUserID(telegramUserID)
	return uuc.oauthService.DelOAuthAccessTokenByKey(key)
}

func createRedisKeyForTelegramUserID(telegramUserID int64) string {
	return fmt.Sprintf("telegram_user_id_%d", telegramUserID)
}
