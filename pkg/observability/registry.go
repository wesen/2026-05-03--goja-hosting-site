package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Observability owns the Prometheus registry and the metric groups used by the
// application. A nil *Observability means instrumentation is disabled.
type Observability struct {
	Registry       *prometheus.Registry
	HTTP           *HTTPMetrics
	Multi          *MultiMetrics
	DB             *DBMetrics
	Guard          *GuardMetrics
	Kanban         *KanbanMetrics
	Tracer         trace.Tracer
	TracingEnabled bool
}

// New creates an isolated registry so tests and embedded servers do not contend
// with the Prometheus default global registry.
func New() *Observability {
	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	httpMetrics := NewHTTPMetrics(registry)
	multiMetrics := NewMultiMetrics(registry)
	dbMetrics := NewDBMetrics(registry)
	guardMetrics := NewGuardMetrics(registry)
	kanbanMetrics := NewKanbanMetrics(registry)

	return &Observability{Registry: registry, HTTP: httpMetrics, Multi: multiMetrics, DB: dbMetrics, Guard: guardMetrics, Kanban: kanbanMetrics, Tracer: otel.Tracer(tracerName)}
}
