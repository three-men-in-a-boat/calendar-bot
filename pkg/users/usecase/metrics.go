package usecase

import (
	"github.com/calendar-bot/pkg/users/repository"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

const usersMetricsNamespace = "users"

const statusMetricLabel = "status"

// nickeskov: counters
var (
	metricGetOrRefreshOAuthAccessTokenByTelegramUserIDTotalCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: usersMetricsNamespace,
			Name:      "get_or_refresh_oauth_access_token_by_telegram_user_id_count",
			Help:      "Total count of 'get or refresh OAuth access token by telegram user ID' requests",
		},
		[]string{statusMetricLabel},
	)
	metricTelegramCreateAuthenticatedUserTotalCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: usersMetricsNamespace,
			Name:      "telegram_create_authenticated_user_count",
			Help:      "Total count of 'create authenticated user for telegram user' requests",
		},
		[]string{statusMetricLabel},
	)
	metricGetUserEmailByTelegramUserIDCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: usersMetricsNamespace,
			Name:      "get_user_email_by_telegram_user_id_count",
			Help:      "Total count of 'get user email by telegram user id' requests",
		},
		[]string{statusMetricLabel},
	)
	metricTryGetUsersEmailsByTelegramUserIDsCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: usersMetricsNamespace,
			Name:      "try_get_users_emails_by_telegram_user_ids_count",
			Help:      "Total count of 'try get users emails by telegram user ids' requests",
		},
		[]string{statusMetricLabel},
	)
	metricIsUserAuthenticatedByTelegramUserIDCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: usersMetricsNamespace,
			Name:      "is_user_authenticated_by_telegram_user_id_count",
			Help:      "Total count of 'is user authenticated by telegram user id' requests",
		},
		[]string{statusMetricLabel},
	)
	metricDeleteLocalAuthenticatedUserByTelegramUserIDCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: usersMetricsNamespace,
			Name:      "delete_local_authenticated_user_by_telegram_user_id_count",
			Help:      "Total count of 'delete local authenticated user by telegram user id' requests",
		},
		[]string{statusMetricLabel},
	)
)

// nickeskov: histograms
var (
	// TODO(nickeskov): specify custom buckets
	metricGetOrRefreshOAuthAccessTokenByTelegramUserIDDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: usersMetricsNamespace,
			Name:      "get_or_refresh_oauth_access_token_by_telegram_user_id_duration",
			Help:      "'get or refresh OAuth access token by telegram user ID' request duration",
		},
	)
	metricTelegramCreateAuthenticatedUserDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: usersMetricsNamespace,
			Name:      "telegram_create_authenticated_user_duration",
			Help:      "'telegram create authenticated user' request duration",
		},
	)
	metricGetUserEmailByTelegramUserIDDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: usersMetricsNamespace,
			Name:      "get_user_email_by_telegram_user_id_duration",
			Help:      "'get user email by telegram user id' request duration",
		},
	)
	metricTryGetUsersEmailsByTelegramUserIDsDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: usersMetricsNamespace,
			Name:      "try_get_users_emails_by_telegram_user_ids_duration",
			Help:      "'get users emails by telegram user ids' duration",
		},
	)
	metricIsUserAuthenticatedByTelegramUserIDDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: usersMetricsNamespace,
			Name:      "is_user_authenticated_by_telegram_user_id_duration",
			Help:      "'is user authenticated by telegram user id' request duration",
		},
	)
	metricDeleteLocalAuthenticatedUserByTelegramUserIDDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: usersMetricsNamespace,
			Name:      "delete_local_authenticated_user_by_telegram_user_id_count",
			Help:      "'delete local authenticated user by telegram user id' request duration",
		},
	)
)

func init() {
	// nickeskov: counters
	prometheus.MustRegister(
		metricGetOrRefreshOAuthAccessTokenByTelegramUserIDTotalCount,
		metricTelegramCreateAuthenticatedUserTotalCount,
		metricGetUserEmailByTelegramUserIDCount,
		metricTryGetUsersEmailsByTelegramUserIDsCount,
		metricIsUserAuthenticatedByTelegramUserIDCount,
		metricDeleteLocalAuthenticatedUserByTelegramUserIDCount,
	)
	// nickeskov: histograms
	prometheus.MustRegister(
		metricGetOrRefreshOAuthAccessTokenByTelegramUserIDDuration,
		metricTelegramCreateAuthenticatedUserDuration,
		metricGetUserEmailByTelegramUserIDDuration,
		metricTryGetUsersEmailsByTelegramUserIDsDuration,
		metricIsUserAuthenticatedByTelegramUserIDDuration,
		metricDeleteLocalAuthenticatedUserByTelegramUserIDDuration,
	)
}

func metricStatusFromErr(err error) string {
	if err == nil {
		return "ok"
	}
	switch err := errors.Cause(err).(type) {
	case repository.UserEntityError:
		switch err {
		case repository.UserDoesNotExist:
			return "user_does_not_exist"
		default:
			return "unknown_user_entity_error"
		}
	//case oauth.Error:
	//	switch err {
	//	case oauth.AccessTokenDoesNotExist:
	//		return "access_token_does_not_exist"
	//	default:
	//		return "unknown_oauth_service_err"
	//	}
	//case *oauth.APIResponseErr:
	//	return fmt.Sprintf("oauth_api_err_%d", err.ErrorCode)
	default:
		return "unknown_err"
	}
}
