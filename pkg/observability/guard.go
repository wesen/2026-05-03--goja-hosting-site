package observability

import (
	"time"

	"github.com/go-go-golems/goja-site/pkg/dbguard"
	"github.com/prometheus/client_golang/prometheus"
)

type GuardMetrics struct {
	Checks          *prometheus.CounterVec
	CheckDuration   *prometheus.HistogramVec
	LimitExceeded   *prometheus.CounterVec
	CleanupAttempts *prometheus.CounterVec
	DBSize          *prometheus.GaugeVec
	DBLimit         *prometheus.GaugeVec
	WritesSince     *prometheus.GaugeVec
}

func NewGuardMetrics(registry *prometheus.Registry) *GuardMetrics {
	m := &GuardMetrics{
		Checks: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "db_guard_checks_total",
			Help:      "Total db.guard checks by phase and result.",
		}, []string{"site", "phase", "result"}),
		CheckDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "db_guard_check_duration_seconds",
			Help:      "db.guard check duration by phase and result.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"site", "phase", "result"}),
		LimitExceeded: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "db_guard_limit_exceeded_total",
			Help:      "db.guard limit exceeded events.",
		}, []string{"site", "kind", "hard"}),
		CleanupAttempts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "db_guard_cleanup_attempts_total",
			Help:      "db.guard cleanup callback attempts by result.",
		}, []string{"site", "result"}),
		DBSize: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "db_size_bytes",
			Help:      "SQLite database size measured by db.guard.",
		}, []string{"site", "component"}),
		DBLimit: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "db_limit_bytes",
			Help:      "SQLite database guard configured byte limits.",
		}, []string{"site", "limit_type"}),
		WritesSince: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "db_guard_writes_since_check",
			Help:      "Number of writes seen by db.guard since process start.",
		}, []string{"site"}),
	}
	registry.MustRegister(m.Checks, m.CheckDuration, m.LimitExceeded, m.CleanupAttempts, m.DBSize, m.DBLimit, m.WritesSince)
	return m
}

type GuardObserver struct {
	site    string
	metrics *GuardMetrics
}

func (o *GuardObserver) ObserveCheck(phase, result string, duration time.Duration) {
	if o == nil || o.metrics == nil {
		return
	}
	o.metrics.Checks.WithLabelValues(o.site, phase, result).Inc()
	o.metrics.CheckDuration.WithLabelValues(o.site, phase, result).Observe(duration.Seconds())
}

func (o *GuardObserver) ObserveLimitExceeded(kind dbguard.SQLKind, hard bool) {
	if o == nil || o.metrics == nil {
		return
	}
	hardLabel := "false"
	if hard {
		hardLabel = "true"
	}
	o.metrics.LimitExceeded.WithLabelValues(o.site, string(kind), hardLabel).Inc()
}

func (o *GuardObserver) ObserveCleanup(result string) {
	if o == nil || o.metrics == nil {
		return
	}
	o.metrics.CleanupAttempts.WithLabelValues(o.site, result).Inc()
}

func (o *GuardObserver) SetStats(stats dbguard.Stats, writes int64) {
	if o == nil || o.metrics == nil {
		return
	}
	o.metrics.DBSize.WithLabelValues(o.site, "db").Set(float64(stats.FileBytes))
	o.metrics.DBSize.WithLabelValues(o.site, "wal").Set(float64(stats.WALBytes))
	o.metrics.DBSize.WithLabelValues(o.site, "shm").Set(float64(stats.SHMBytes))
	o.metrics.DBSize.WithLabelValues(o.site, "total").Set(float64(stats.TotalBytes))
	o.metrics.DBSize.WithLabelValues(o.site, "live").Set(float64(stats.EstimatedLiveBytes))
	o.metrics.DBLimit.WithLabelValues(o.site, "max").Set(float64(stats.MaxBytes))
	o.metrics.DBLimit.WithLabelValues(o.site, "soft").Set(float64(stats.SoftMaxBytes))
	o.metrics.DBLimit.WithLabelValues(o.site, "hard").Set(float64(stats.HardMaxBytes))
	o.metrics.WritesSince.WithLabelValues(o.site).Set(float64(writes))
}
