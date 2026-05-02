package metrics

import (
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var (
	Registry = prom.NewRegistry()

	RepoTotal = prom.NewGauge(prom.GaugeOpts{
		Name: "gms_repo_total",
		Help: "Total number of registered repositories.",
	})

	SyncTaskTotal = prom.NewCounter(prom.CounterOpts{
		Name: "gms_sync_task_total",
		Help: "Total number of sync task executions.",
	})

	SyncTaskDuration = prom.NewHistogram(prom.HistogramOpts{
		Name:    "gms_sync_task_duration_seconds",
		Help:    "Duration of sync task execution in seconds.",
		Buckets: prom.DefBuckets,
	})

	SyncTaskFailures = prom.NewCounter(prom.CounterOpts{
		Name: "gms_sync_task_failures_total",
		Help: "Total number of failed sync task executions.",
	})

	GitOperationTotal = prom.NewCounterVec(
		prom.CounterOpts{
			Name: "gms_git_operation_total",
			Help: "Total number of git operations.",
		},
		[]string{"operation"},
	)

	ActiveWebhooks = prom.NewGauge(prom.GaugeOpts{
		Name: "gms_active_webhooks",
		Help: "Number of active webhooks.",
	})

	WebhookEventsTotal = prom.NewCounterVec(
		prom.CounterOpts{
			Name: "gms_webhook_events_total",
			Help: "Total number of webhook events received.",
		},
		[]string{"event_type"},
	)
)

func init() {
	Registry.MustRegister(collectors.NewGoCollector())
	Registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	Registry.MustRegister(RepoTotal)
	Registry.MustRegister(SyncTaskTotal)
	Registry.MustRegister(SyncTaskDuration)
	Registry.MustRegister(SyncTaskFailures)
	Registry.MustRegister(GitOperationTotal)
	Registry.MustRegister(ActiveWebhooks)
	Registry.MustRegister(WebhookEventsTotal)
}

func RecordGitOperation(op string) {
	GitOperationTotal.WithLabelValues(op).Inc()
}

func RecordSyncTask(durationSeconds float64, success bool) {
	SyncTaskTotal.Inc()
	SyncTaskDuration.Observe(durationSeconds)
	if !success {
		SyncTaskFailures.Inc()
	}
}

func RecordWebhookEvent(eventType string) {
	WebhookEventsTotal.WithLabelValues(eventType).Inc()
}
