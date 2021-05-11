package usecase

import (
	"github.com/prometheus/client_golang/prometheus"
)

const eventsMetricsNamespace = "events"

const statusMetricLabel = "status"

// nickeskov: counters
var (
	metricGetEventsBySpecificDayTotalCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: eventsMetricsNamespace,
			Name:      "get_events_by_specific_day_count",
			Help:      "Total count of 'get events by specific day' requests",
		},
		[]string{statusMetricLabel},
	)
	metricGetClosestEventTotalCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: eventsMetricsNamespace,
			Name:      "get_closest_event_count",
			Help:      "Total count of 'get closest event' requests",
		},
		[]string{statusMetricLabel},
	)
	metricGetEventByEventIDTotalCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: eventsMetricsNamespace,
			Name:      "get_event_by_event_id_count",
			Help:      "Total count of 'get event by event id' requests",
		},
		[]string{statusMetricLabel},
	)
	metricGetUsersBusyIntervalsTotalCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: eventsMetricsNamespace,
			Name:      "get_user_busy_intervals_count",
			Help:      "Total count of 'get user busy intervals' requests",
		},
		[]string{statusMetricLabel},
	)
	metricGetUsersFreeIntervalsTotalCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: eventsMetricsNamespace,
			Name:      "get_user_free_intervals_count",
			Help:      "Total count of 'get user busy intervals' requests",
		},
		[]string{statusMetricLabel},
	)
	metricCreateEventTotalCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: eventsMetricsNamespace,
			Name:      "create_event_count",
			Help:      "Total count of 'create event' requests",
		},
		[]string{statusMetricLabel},
	)
	metricAddAttendeeTotalCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: eventsMetricsNamespace,
			Name:      "add_attendee_count",
			Help:      "Total count of 'create event' requests",
		},
		[]string{statusMetricLabel},
	)
	metricChangeStatusTotalCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: eventsMetricsNamespace,
			Name:      "change_status_count",
			Help:      "total count of 'change status' requests",
		},
		[]string{statusMetricLabel},
	)
)

// nickeskov: histograms
var (
	// TODO(nickeskov): specify custom buckets
	metricGetEventsBySpecificDayDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: eventsMetricsNamespace,
			Name:      "get_events_by_specific_day_duration",
			Help:      "'get events by specific day' request duration",
		},
	)
	metricGetClosestEventDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: eventsMetricsNamespace,
			Name:      "get_closest_event_duration",
			Help:      "'get closes event' request duration",
		},
	)
	metricGetEventByEventIDDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: eventsMetricsNamespace,
			Name:      "get_event_by_event_id_duration",
			Help:      "'get event by event id' request duration",
		},
	)
	metricGetUsersBusyIntervalsDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: eventsMetricsNamespace,
			Name:      "get_users_busy_intervals_duration",
			Help:      "'get users busy intervals' request duration",
		},
	)
	metricGetUsersFreeIntervalsDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: eventsMetricsNamespace,
			Name:      "get_users_free_intervals_duration",
			Help:      "'get users free intervals' request duration",
		},
	)
	metricCreateEventDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: eventsMetricsNamespace,
			Name:      "create_event_duration",
			Help:      "'create event' request duration",
		},
	)
	metricAddAttendeeDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: eventsMetricsNamespace,
			Name:      "add_attendee_duration",
			Help:      "'add attendee' request duration",
		},
	)
	metricChangeStatusDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: eventsMetricsNamespace,
			Name:      "change_status_duration",
			Help:      "'change status' request duration",
		},
	)
)

func init() {
	// nickeskov: counters
	prometheus.MustRegister(
		metricGetEventsBySpecificDayTotalCount,
		metricGetClosestEventTotalCount,
		metricGetEventByEventIDTotalCount,
		metricGetUsersBusyIntervalsTotalCount,
		metricGetUsersFreeIntervalsTotalCount,
		metricCreateEventTotalCount,
		metricAddAttendeeTotalCount,
		metricChangeStatusTotalCount,
	)
	// nickeskov: histograms
	prometheus.MustRegister(
		metricGetEventsBySpecificDayDuration,
		metricGetClosestEventDuration,
		metricGetEventByEventIDDuration,
		metricGetUsersBusyIntervalsDuration,
		metricGetUsersFreeIntervalsDuration,
		metricCreateEventDuration,
		metricAddAttendeeDuration,
		metricChangeStatusDuration,
	)
}

func metricStatusFromErr(err error) string {
	if err == nil {
		return "ok"
	}
	return "unknown_err"
}
