package oauth

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"github.com/calendar-bot/pkg/customerrors"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"time"
)

const (
	nonceBitsLength = 256
	nonceBase       = 16
)

type Service struct {
	config  *Config
	redisDB *redis.Client
}

func NewService(config *Config, redisDB *redis.Client) Service {
	return Service{
		config:  config,
		redisDB: redisDB,
	}
}

func (s *Service) Config() *Config {
	return s.config
}

func (s *Service) GenerateOAuthLink(stateValue interface{}, expire time.Duration) (loginLink LoginLink, err error) {
	timer := prometheus.NewTimer(metricGenerateLoginLinkDuration)
	defer func() {
		metricGeneratedLoginLinksTotalCount.WithLabelValues(metricStatusFromErr(err)).Inc()
		timer.ObserveDuration()
	}()

	stateKey, err := generateStateKey()
	if err != nil {
		return LoginLink{}, errors.WithStack(err)
	}

	if setErr := s.setStateValueByStateKey(stateKey, stateValue, expire); setErr != nil {
		return LoginLink{}, errors.Wrapf(setErr, "failed to set stateValue '%v' in redis storage", stateValue)
	}

	loginLink, err = newLoginLink(s.config, stateKey)
	if err != nil {
		return LoginLink{}, errors.WithStack(err)
	}
	return loginLink, nil
}

func (s *Service) GenerateOAuthLinkWithDefaultExpire(stateValue interface{}) (LoginLink, error) {
	return s.GenerateOAuthLink(stateValue, s.config.LinkExpireIn)
}

func (s *Service) setStateValueByStateKey(stateKey string, stateValue interface{}, expire time.Duration) (err error) {
	defer func() {
		metricStateSetTotalCount.WithLabelValues(metricStatusFromErr(err)).Inc()
	}()

	if setErr := s.redisDB.Set(context.TODO(), stateKey, stateValue, expire).Err(); setErr != nil {
		return errors.WithStack(setErr)
	}
	return nil
}

func (s *Service) ScanStateValueByStateKey(stateKey string, outValue interface{}) (err error) {
	defer func() {
		metricStateScanTotalCount.WithLabelValues(metricStatusFromErr(err)).Inc()
	}()

	res := s.redisDB.Get(context.TODO(), stateKey)

	if err := res.Err(); err != nil {
		switch err {
		case redis.Nil:
			return StateKeyDoesNotExist
		default:
			return errors.Wrapf(err, "failed to get StateValue by StateKey='%s'", stateKey)
		}
	}

	if err := res.Scan(outValue); err != nil {
		return errors.Wrapf(err, "cannot scan StateValue into %T by StateKey='%s'", outValue, stateKey)
	}

	return nil
}

func (s *Service) SetAccessTokenByKey(key, accessToken string, expire time.Duration) (err error) {
	defer func() {
		metricAccessTokenSetTotalCount.WithLabelValues(metricStatusFromErr(err)).Inc()
	}()

	return s.redisDB.Set(context.TODO(), key, accessToken, expire).Err()
}

func (s *Service) GetAccessTokenByKey(key string) (token string, err error) {
	defer func() {
		metricAccessTokenGetTotalCount.WithLabelValues(metricStatusFromErr(err)).Inc()
	}()

	res := s.redisDB.Get(context.TODO(), key)

	if err := res.Err(); err != nil {
		switch err {
		case redis.Nil:
			return "", AccessTokenDoesNotExist
		default:
			return "", errors.Wrapf(err, "failed to get AccessToken by key='%s'", key)
		}
	}

	return res.Val(), nil
}

func (s *Service) DelOAuthAccessTokenByKey(key string) (err error) {
	defer func() {
		metricAccessTokenDeleteTotalCount.WithLabelValues(metricStatusFromErr(err)).Inc()
	}()

	if err := s.redisDB.Del(context.TODO(), key).Err(); err != nil {
		switch err {
		case redis.Nil:
			return nil
		default:
			return errors.Wrapf(err, "failed to del AccessToken by key='%s'", key)
		}
	}
	return nil
}

func (s *Service) ObtainTokensFromOAuthHost(authCode string) (response ObtainTokensResponse, err error) {
	timer := prometheus.NewTimer(metricTokensObtainFromOAuthHostDuration)
	defer func() {
		metricTokensObtainFromOAuthHostTotalCount.WithLabelValues(metricStatusFromErr(err)).Inc()
		timer.ObserveDuration()
	}()

	form, err := newObtainTokensPostForm(s.config, authCode)
	if err != nil {
		return ObtainTokensResponse{}, errors.WithStack(err)
	}

	apiResponse, err := http.PostForm(form.tokensUrl.String(), form.values)
	if err != nil {
		return ObtainTokensResponse{}, errors.Wrap(err, "failed to obtain tokens from api")
	}
	defer func() {
		err = customerrors.HandleCloser(err, apiResponse.Body)
	}()

	if err := json.NewDecoder(apiResponse.Body).Decode(&response); err != nil {
		return ObtainTokensResponse{},
			errors.Wrapf(err, "cannot decode json body into %T struct", response)
	}

	if response.APIResponseErr != nil {
		return ObtainTokensResponse{}, response.APIResponseErr
	}

	// nickeskov: in this case http status must be ok
	if apiResponse.StatusCode != http.StatusOK {
		return ObtainTokensResponse{},
			errors.Errorf(
				"something wrong with token obtaining and oauth API: http_status='%s'",
				apiResponse.Status,
			)
	}

	return response, nil
}

func (s *Service) RenewAccessTokenByRefreshToken(refreshToken string) (response RenewAccessTokenResponse, err error) {
	timer := prometheus.NewTimer(metricRenewAccessTokenDuration)
	defer func() {
		metricRenewAccessTokenTotalCount.WithLabelValues(metricStatusFromErr(err)).Inc()
		timer.ObserveDuration()
	}()

	form, err := newRenewAccessTokenPostForm(s.config, refreshToken)
	if err != nil {
		return RenewAccessTokenResponse{}, errors.WithStack(err)
	}

	apiResponse, err := http.PostForm(form.tokensUrl.String(), form.values)
	if err != nil {
		return RenewAccessTokenResponse{}, errors.Wrap(err, "failed to renew new access token from api")
	}
	defer func() {
		err = customerrors.HandleCloser(err, apiResponse.Body)
	}()

	if err := json.NewDecoder(apiResponse.Body).Decode(&response); err != nil {
		return RenewAccessTokenResponse{},
			errors.Wrapf(err, "cannot decode json body into %T struct", response)
	}

	if response.APIResponseErr != nil {
		return RenewAccessTokenResponse{}, response.APIResponseErr
	}

	// nickeskov: in this case http status must be ok
	if apiResponse.StatusCode != http.StatusOK {
		return RenewAccessTokenResponse{},
			errors.Errorf(
				"something wrong with access token renewal and oauth API: http_status='%s'",
				apiResponse.Status,
			)
	}

	return response, nil
}

func (s *Service) GetUserInfo(accessToken string) (response UserInfoResponse, err error) {
	timer := prometheus.NewTimer(metricUserInfoRequestDuration)
	defer func() {
		metricUserInfoRequestsTotalCount.WithLabelValues(metricStatusFromErr(err)).Inc()
		timer.ObserveDuration()
	}()

	userInfoLink, err := newUserInfoLink(s.config, accessToken)
	if err != nil {
		return UserInfoResponse{}, errors.WithStack(err)
	}

	apiResponse, err := http.Get(userInfoLink.String())
	if err != nil {
		return UserInfoResponse{}, errors.Wrap(err, "failed to send userinfo request to oauth API")
	}
	defer func() {
		err = customerrors.HandleCloser(err, apiResponse.Body)
	}()

	if err := json.NewDecoder(apiResponse.Body).Decode(&response); err != nil {
		return UserInfoResponse{}, errors.Wrapf(err, "cannot decode json body into %T struct", response)
	}

	if response.APIResponseErr != nil {
		return UserInfoResponse{}, response.APIResponseErr
	}

	// nickeskov: in this case http status must be ok
	if apiResponse.StatusCode != http.StatusOK {
		return UserInfoResponse{},
			errors.Errorf(
				"something wrong with userinfo request and oauth API: http_status='%s'",
				apiResponse.Status,
			)
	}

	return response, nil
}

func generateStateKey() (string, error) {
	bigInt, err := rand.Prime(rand.Reader, nonceBitsLength)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate StateKey")
	}
	return bigInt.Text(nonceBase), nil
}
