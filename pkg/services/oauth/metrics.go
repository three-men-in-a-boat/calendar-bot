package oauth

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

const oauthMetricsNamespace = "oauth_service"

const statusMetricLabel = "status"

// nickeskov: counters
var (
	metricAccessTokenGetTotalCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: oauthMetricsNamespace,
			Name:      "access_token_get_count",
			Help:      "Total access tokens fetches count",
		},
		[]string{statusMetricLabel},
	)
	metricAccessTokenSetTotalCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: oauthMetricsNamespace,
			Name:      "access_token_set_count",
			Help:      "Total access tokens sets count",
		},
		[]string{statusMetricLabel},
	)
	metricAccessTokenDeleteTotalCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: oauthMetricsNamespace,
			Name:      "access_token_delete_count",
			Help:      "Total access tokens deletions count",
		},
		[]string{statusMetricLabel},
	)
	metricStateSetTotalCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: oauthMetricsNamespace,
			Name:      "state_set_count",
			Help:      "Total sets count in state for oauth login links",
		},
		[]string{statusMetricLabel},
	)
	metricStateScanTotalCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: oauthMetricsNamespace,
			Name:      "state_scan_count",
			Help:      "Total scans count from state for oauth login links",
		},
		[]string{statusMetricLabel},
	)
	metricGeneratedLoginLinksTotalCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: oauthMetricsNamespace,
			Name:      "generated_login_links_count",
			Help:      "Total generated login links count",
		},
		[]string{statusMetricLabel},
	)
	metricTokensObtainFromOAuthHostTotalCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: oauthMetricsNamespace,
			Name:      "tokens_obtain_from_oauth_host_count",
			Help:      "Total tokens obtain count from remote oauth host",
		},
		[]string{statusMetricLabel},
	)
	metricRenewAccessTokenTotalCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: oauthMetricsNamespace,
			Name:      "renew_access_token_count",
			Help:      "Total access tokens renewals count from oauth host",
		},
		[]string{statusMetricLabel},
	)
	metricUserInfoRequestsTotalCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: oauthMetricsNamespace,
			Name:      "user_info_requests_count",
			Help:      "Total user info requests count",
		},
		[]string{statusMetricLabel},
	)
)

// nickeskov: histograms
var (
	// TODO(nickeskov): specify custom buckets
	metricGenerateLoginLinkDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: oauthMetricsNamespace,
			Name:      "generate_login_link_duration",
			Help:      "Login link generation duration",
		},
	)
	metricTokensObtainFromOAuthHostDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: oauthMetricsNamespace,
			Name:      "tokens_obtain_from_oauth_host_request_duration",
			Help:      "Tokens obtain count from remote oauth host request duration",
		},
	)
	metricRenewAccessTokenDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: oauthMetricsNamespace,
			Name:      "renew_access_token_request_duration",
			Help:      "Renew access token request duration",
		},
	)
	metricUserInfoRequestDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: oauthMetricsNamespace,
			Name:      "user_info_request_duration",
			Help:      "User info request duration",
		},
	)
)

func init() {
	// nickeskov: counters
	prometheus.MustRegister(
		metricAccessTokenGetTotalCount,
		metricAccessTokenSetTotalCount,
		metricAccessTokenDeleteTotalCount,
		metricStateSetTotalCount,
		metricStateScanTotalCount,
		metricGeneratedLoginLinksTotalCount,
		metricTokensObtainFromOAuthHostTotalCount,
		metricRenewAccessTokenTotalCount,
		metricUserInfoRequestsTotalCount,
	)
	// nickeskov: histograms
	prometheus.MustRegister(
		metricGenerateLoginLinkDuration,
		metricTokensObtainFromOAuthHostDuration,
		metricRenewAccessTokenDuration,
		metricUserInfoRequestDuration,
	)
}

func metricStatusFromErr(err error) string {
	if err == nil {
		return "ok"
	}
	switch err := errors.Cause(err).(type) {
	case *APIResponseErr:
		return fmt.Sprintf("oauth_api_err_%d", err.ErrorCode)
	case redis.Error:
		switch err {
		case redis.Nil:
			return "redis_nil"
		default:
			return "redis_err"
		}
	case Error:
		switch err {
		case AccessTokenDoesNotExist:
			return "access_token_does_not_exist"
		case StateKeyDoesNotExist:
			return "state_key_does_not_exist"
		default:
			return "unknown_oauth_service_err"
		}
	default:
		return "unknown_err"
	}
}
