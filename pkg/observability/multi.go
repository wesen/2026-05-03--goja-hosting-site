package observability

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type MultiMetrics struct {
	HostsConfigured *prometheus.GaugeVec
	SiteUp          *prometheus.GaugeVec
	UnknownHosts    *prometheus.CounterVec
	Dispatch        *prometheus.HistogramVec
}

func NewMultiMetrics(registry *prometheus.Registry) *MultiMetrics {
	m := &MultiMetrics{
		HostsConfigured: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "hosts_configured",
			Help:      "Number of configured goja-site hosts in this process.",
		}, []string{"mode"}),
		SiteUp: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "site_up",
			Help:      "Whether a configured site was successfully created.",
		}, []string{"site"}),
		UnknownHosts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "unknown_host_requests_total",
			Help:      "Requests for hosts that do not match a configured site. The host label is intentionally classified to avoid cardinality explosions.",
		}, []string{"host_class"}),
		Dispatch: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "multi_dispatch_duration_seconds",
			Help:      "Time spent in multi-site host dispatch, including downstream site handling for successful dispatches.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"result"}),
	}
	registry.MustRegister(m.HostsConfigured, m.SiteUp, m.UnknownHosts, m.Dispatch)
	return m
}

func (m *MultiMetrics) SetHostsConfigured(mode string, count int) {
	if m == nil {
		return
	}
	m.HostsConfigured.WithLabelValues(mode).Set(float64(count))
}

func (m *MultiMetrics) SetSiteUp(site string, up bool) {
	if m == nil {
		return
	}
	value := 0.0
	if up {
		value = 1
	}
	m.SiteUp.WithLabelValues(SiteLabel(site)).Set(value)
}

func (m *MultiMetrics) ObserveUnknownHost() {
	if m == nil {
		return
	}
	m.UnknownHosts.WithLabelValues("unknown").Inc()
}

func (m *MultiMetrics) ObserveDispatch(result string, started time.Time) {
	if m == nil {
		return
	}
	m.Dispatch.WithLabelValues(result).Observe(time.Since(started).Seconds())
}
