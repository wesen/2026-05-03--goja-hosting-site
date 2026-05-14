package observability

import (
	"database/sql"
	"strings"
	"time"

	databasemod "github.com/go-go-golems/go-go-goja/modules/database"
	"github.com/go-go-golems/goja-site/pkg/dbguard"
	"github.com/prometheus/client_golang/prometheus"
)

type DBMetrics struct {
	Operations *prometheus.CounterVec
	Duration   *prometheus.HistogramVec
	Errors     *prometheus.CounterVec
}

func NewDBMetrics(registry *prometheus.Registry) *DBMetrics {
	m := &DBMetrics{
		Operations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "db_operations_total",
			Help:      "Total database operations issued by JavaScript code.",
		}, []string{"site", "db_policy", "operation", "sql_kind"}),
		Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "db_operation_duration_seconds",
			Help:      "Database operation duration by site, policy, operation, and coarse SQL kind.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"site", "db_policy", "operation", "sql_kind"}),
		Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "db_errors_total",
			Help:      "Total database operation errors by site, policy, operation, and coarse SQL kind.",
		}, []string{"site", "db_policy", "operation", "sql_kind", "error_class"}),
	}
	registry.MustRegister(m.Operations, m.Duration, m.Errors)
	return m
}

type InstrumentedQueryExecer struct {
	inner   databasemod.QueryExecer
	site    string
	policy  string
	metrics *DBMetrics
}

func InstrumentQueryExecer(inner databasemod.QueryExecer, site, policy string, metrics *DBMetrics) databasemod.QueryExecer {
	if inner == nil || metrics == nil {
		return inner
	}
	return &InstrumentedQueryExecer{inner: inner, site: SiteLabel(site), policy: policy, metrics: metrics}
}

func (i *InstrumentedQueryExecer) Query(query string, args ...any) (*sql.Rows, error) {
	kind := SQLKindLabel(query)
	start := time.Now()
	rows, err := i.inner.Query(query, args...)
	i.observe("query", kind, start, err)
	return rows, err
}

func (i *InstrumentedQueryExecer) Exec(query string, args ...any) (sql.Result, error) {
	kind := SQLKindLabel(query)
	start := time.Now()
	result, err := i.inner.Exec(query, args...)
	i.observe("exec", kind, start, err)
	return result, err
}

func (i *InstrumentedQueryExecer) observe(operation, kind string, start time.Time, err error) {
	i.metrics.Operations.WithLabelValues(i.site, i.policy, operation, kind).Inc()
	i.metrics.Duration.WithLabelValues(i.site, i.policy, operation, kind).Observe(time.Since(start).Seconds())
	if err != nil {
		i.metrics.Errors.WithLabelValues(i.site, i.policy, operation, kind, ErrorClass(err)).Inc()
	}
}

func SQLKindLabel(query string) string {
	switch firstSQLToken(query) {
	case "SELECT":
		return "select"
	case "INSERT":
		return "insert"
	case "UPDATE":
		return "update"
	case "DELETE":
		return "delete"
	case "REPLACE":
		return "replace"
	case "CREATE":
		return "create"
	case "ALTER":
		return "alter"
	case "DROP":
		return "drop"
	case "PRAGMA":
		return "pragma"
	case "WITH":
		return "with"
	case "EXPLAIN":
		return "explain"
	case "VACUUM":
		return "vacuum"
	case "ANALYZE":
		return "analyze"
	case "REINDEX":
		return "reindex"
	case "":
		return "unknown"
	default:
		return "other"
	}
}

func ErrorClass(err error) string {
	if err == nil {
		return "none"
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "writes are disabled"):
		return "writes_disabled"
	case strings.Contains(msg, "hard limit"):
		return "hard_limit"
	case strings.Contains(msg, "no such table"):
		return "no_such_table"
	case strings.Contains(msg, "constraint"):
		return "constraint"
	case strings.Contains(msg, "locked"):
		return "locked"
	default:
		return "other"
	}
}

func firstSQLToken(query string) string {
	s := strings.TrimSpace(query)
	for {
		if strings.HasPrefix(s, "--") {
			idx := strings.IndexByte(s, '\n')
			if idx < 0 {
				return ""
			}
			s = strings.TrimSpace(s[idx+1:])
			continue
		}
		if strings.HasPrefix(s, "/*") {
			idx := strings.Index(s, "*/")
			if idx < 0 {
				return ""
			}
			s = strings.TrimSpace(s[idx+2:])
			continue
		}
		break
	}
	if s == "" {
		return ""
	}
	for i, r := range s {
		if !(r == '_' || r == '-' || r >= '0' && r <= '9' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z') {
			return strings.ToUpper(s[:i])
		}
	}
	return strings.ToUpper(s)
}

func NewGuardObserver(site string, metrics *GuardMetrics) dbguard.Observer {
	if metrics == nil {
		return nil
	}
	return &GuardObserver{site: SiteLabel(site), metrics: metrics}
}
