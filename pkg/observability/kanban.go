package observability

import (
	"time"

	"github.com/go-go-golems/goja-site/pkg/kanbanddsl"
	"github.com/prometheus/client_golang/prometheus"
)

type KanbanMetrics struct {
	Fragments     *prometheus.HistogramVec
	Actions       *prometheus.HistogramVec
	Dispatch      *prometheus.HistogramVec
	Render        *prometheus.HistogramVec
	RenderedBytes *prometheus.HistogramVec
	Errors        *prometheus.CounterVec
}

func NewKanbanMetrics(registry *prometheus.Registry) *KanbanMetrics {
	m := &KanbanMetrics{
		Fragments: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "kanban_fragment_duration_seconds",
			Help:      "Kanban fragment route duration by site and board.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"site", "board"}),
		Actions: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "kanban_action_duration_seconds",
			Help:      "Kanban action route duration by site, board, action, and refresh behavior.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"site", "board", "action", "refresh"}),
		Dispatch: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "kanban_dispatch_duration_seconds",
			Help:      "Kanban action callback dispatch duration by site, board, and action.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"site", "board", "action"}),
		Render: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "kanban_render_duration_seconds",
			Help:      "Kanban render duration by site, board, and reason.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"site", "board", "reason"}),
		RenderedBytes: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "kanban_rendered_html_bytes",
			Help:      "Kanban rendered HTML bytes by site, board, and reason.",
			Buckets:   []float64{64, 256, 1024, 4096, 16384, 65536, 262144, 1048576},
		}, []string{"site", "board", "reason"}),
		Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "kanban_errors_total",
			Help:      "Kanban errors by site, board, action, phase, and bounded error class.",
		}, []string{"site", "board", "action", "phase", "error_class"}),
	}
	registry.MustRegister(m.Fragments, m.Actions, m.Dispatch, m.Render, m.RenderedBytes, m.Errors)
	return m
}

type KanbanObserver struct {
	site    string
	metrics *KanbanMetrics
}

func NewKanbanObserver(site string, metrics *KanbanMetrics) kanbanddsl.Observer {
	if metrics == nil {
		return nil
	}
	return &KanbanObserver{site: SiteLabel(site), metrics: metrics}
}

func (o *KanbanObserver) ObserveFragment(board string, duration time.Duration, err error) {
	if o == nil || o.metrics == nil {
		return
	}
	board = kanbanLabel(board)
	o.metrics.Fragments.WithLabelValues(o.site, board).Observe(duration.Seconds())
	if err != nil {
		o.metrics.Errors.WithLabelValues(o.site, board, "none", "fragment", ErrorClass(err)).Inc()
	}
}

func (o *KanbanObserver) ObserveAction(board, action string, refresh bool, duration time.Duration, err error) {
	if o == nil || o.metrics == nil {
		return
	}
	board = kanbanLabel(board)
	action = kanbanLabel(action)
	refreshLabel := "false"
	if refresh {
		refreshLabel = "true"
	}
	o.metrics.Actions.WithLabelValues(o.site, board, action, refreshLabel).Observe(duration.Seconds())
	if err != nil {
		o.metrics.Errors.WithLabelValues(o.site, board, action, "action", ErrorClass(err)).Inc()
	}
}

func (o *KanbanObserver) ObserveDispatch(board, action string, duration time.Duration, err error) {
	if o == nil || o.metrics == nil {
		return
	}
	board = kanbanLabel(board)
	action = kanbanLabel(action)
	o.metrics.Dispatch.WithLabelValues(o.site, board, action).Observe(duration.Seconds())
	if err != nil {
		o.metrics.Errors.WithLabelValues(o.site, board, action, "dispatch", ErrorClass(err)).Inc()
	}
}

func (o *KanbanObserver) ObserveRender(board, reason string, duration time.Duration, htmlBytes int, err error) {
	if o == nil || o.metrics == nil {
		return
	}
	board = kanbanLabel(board)
	reason = kanbanLabel(reason)
	o.metrics.Render.WithLabelValues(o.site, board, reason).Observe(duration.Seconds())
	if htmlBytes >= 0 {
		o.metrics.RenderedBytes.WithLabelValues(o.site, board, reason).Observe(float64(htmlBytes))
	}
	if err != nil {
		o.metrics.Errors.WithLabelValues(o.site, board, "none", "render", ErrorClass(err)).Inc()
	}
}

func kanbanLabel(value string) string {
	if value == "" {
		return "unknown"
	}
	return value
}
