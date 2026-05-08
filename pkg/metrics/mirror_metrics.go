package metrics

import (
	prom "github.com/prometheus/client_golang/prometheus"
)

var (
	MirrorSyncTotal = prom.NewCounterVec(
		prom.CounterOpts{
			Name: "gms_mirror_sync_total",
			Help: "Total number of mirror sync operations.",
		},
		[]string{"mirror_type", "status"},
	)

	MirrorSyncDuration = prom.NewHistogramVec(
		prom.HistogramOpts{
			Name:    "gms_mirror_sync_duration_seconds",
			Help:    "Duration of mirror sync operations in seconds.",
			Buckets: prom.DefBuckets,
		},
		[]string{"mirror_type"},
	)

	MirrorSyncQueueSize = prom.NewGauge(prom.GaugeOpts{
		Name: "gms_mirror_sync_queue_size",
		Help: "Current number of items in the mirror sync queue.",
	})

	MirrorSyncActiveWorkers = prom.NewGauge(prom.GaugeOpts{
		Name: "gms_mirror_sync_active_workers",
		Help: "Number of active mirror sync workers.",
	})

	MirrorSyncRetryTotal = prom.NewCounterVec(
		prom.CounterOpts{
			Name: "gms_mirror_sync_retry_total",
			Help: "Total number of mirror sync retries.",
		},
		[]string{"mirror_id"},
	)

	MirrorSyncLastSuccess = prom.NewGaugeVec(
		prom.GaugeOpts{
			Name: "gms_mirror_sync_last_success_timestamp",
			Help: "Unix timestamp of last successful sync per mirror.",
		},
		[]string{"mirror_id"},
	)

	MirrorSyncBranches = prom.NewCounterVec(
		prom.CounterOpts{
			Name: "gms_mirror_sync_branches_total",
			Help: "Total number of branches synced.",
		},
		[]string{"mirror_type"},
	)
)

func init() {
	Registry.MustRegister(MirrorSyncTotal)
	Registry.MustRegister(MirrorSyncDuration)
	Registry.MustRegister(MirrorSyncQueueSize)
	Registry.MustRegister(MirrorSyncActiveWorkers)
	Registry.MustRegister(MirrorSyncRetryTotal)
	Registry.MustRegister(MirrorSyncLastSuccess)
	Registry.MustRegister(MirrorSyncBranches)
}

func RecordMirrorSync(mirrorType, status string, durationSeconds float64) {
	MirrorSyncTotal.WithLabelValues(mirrorType, status).Inc()
	MirrorSyncDuration.WithLabelValues(mirrorType).Observe(durationSeconds)
}

func RecordMirrorRetry(mirrorID string) {
	MirrorSyncRetryTotal.WithLabelValues(mirrorID).Inc()
}

func RecordMirrorSuccess(mirrorID string) {
	MirrorSyncLastSuccess.WithLabelValues(mirrorID).SetToCurrentTime()
}

func SetMirrorQueueSize(size int) {
	MirrorSyncQueueSize.Set(float64(size))
}

func SetMirrorActiveWorkers(count int32) {
	MirrorSyncActiveWorkers.Set(float64(count))
}

func RecordMirrorBranches(mirrorType string, count int) {
	MirrorSyncBranches.WithLabelValues(mirrorType).Add(float64(count))
}
