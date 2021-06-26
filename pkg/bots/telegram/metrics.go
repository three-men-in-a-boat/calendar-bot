package telegram

import (
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/tucnak/telebot.v2"
)

const telegramBotMetricsNamespace = "telegram_bot"

func NewPollerWithTotalHitsMetric(poller telebot.Poller) telebot.Poller {
	return telebot.NewMiddlewarePoller(poller, func(upd *telebot.Update) bool {
		MetricTotalHitsCount.Inc()
		return true
	})
}

var (
	MetricTotalHitsCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: telegramBotMetricsNamespace,
			Name:      "hits_count",
			Help:      "Total telegram bot hits count",
		},
	)
	MetricTotalErrorsCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: telegramBotMetricsNamespace,
			Name:      "errors_count",
			Help:      "Total telegram bot errors count",
		},
	)
)

func init() {
	prometheus.MustRegister(
		MetricTotalHitsCount,
		MetricTotalErrorsCount,
	)
}
