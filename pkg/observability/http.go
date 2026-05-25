package observability

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "goja_site"

type HTTPMetrics struct {
	Requests      *prometheus.CounterVec
	Duration      *prometheus.HistogramVec
	ResponseBytes *prometheus.HistogramVec
	InFlight      *prometheus.GaugeVec
}

func NewHTTPMetrics(registry *prometheus.Registry) *HTTPMetrics {
	m := &HTTPMetrics{
		Requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total HTTP requests handled by goja-site.",
		}, []string{"site", "method", "route", "status_class"}),
		Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration by site, method, and coarse route.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"site", "method", "route"}),
		ResponseBytes: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_response_bytes",
			Help:      "HTTP response body bytes by site, method, and coarse route.",
			Buckets:   []float64{64, 256, 1024, 4096, 16384, 65536, 262144, 1048576},
		}, []string{"site", "method", "route"}),
		InFlight: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "http_in_flight_requests",
			Help:      "Current in-flight HTTP requests by site.",
		}, []string{"site"}),
	}
	registry.MustRegister(m.Requests, m.Duration, m.ResponseBytes, m.InFlight)
	return m
}

func (m *HTTPMetrics) Wrap(site string, next http.Handler) http.Handler {
	if m == nil || next == nil {
		return next
	}
	site = SiteLabel(site)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := MethodLabel(r.Method)
		route := CoarseRoute(r.URL.Path)
		m.InFlight.WithLabelValues(site).Inc()
		start := time.Now()
		rec := &StatusRecorder{ResponseWriter: w}
		defer func() {
			m.InFlight.WithLabelValues(site).Dec()
			status := rec.Status()
			m.Requests.WithLabelValues(site, method, route, StatusClass(status)).Inc()
			m.Duration.WithLabelValues(site, method, route).Observe(time.Since(start).Seconds())
			m.ResponseBytes.WithLabelValues(site, method, route).Observe(float64(rec.Bytes()))
		}()
		next.ServeHTTP(rec, r)
	})
}

type StatusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *StatusRecorder) WriteHeader(status int) {
	if r.status != 0 {
		return
	}
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *StatusRecorder) Write(p []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(p)
	r.bytes += n
	return n, err
}

func (r *StatusRecorder) Status() int {
	if r.status == 0 {
		return http.StatusOK
	}
	return r.status
}

func (r *StatusRecorder) Bytes() int { return r.bytes }
