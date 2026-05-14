package observability

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/go-go-golems/goja-site"

type TracingConfig struct {
	Enabled     bool
	ServiceName string
	Endpoint    string
	SampleRatio float64
}

type Tracing struct {
	Tracer   trace.Tracer
	Shutdown func(context.Context) error
}

func InitTracing(ctx context.Context, cfg TracingConfig) (*Tracing, error) {
	if !cfg.Enabled {
		return &Tracing{Tracer: otel.Tracer(tracerName), Shutdown: func(context.Context) error { return nil }}, nil
	}
	serviceName := strings.TrimSpace(cfg.ServiceName)
	if serviceName == "" {
		serviceName = "goja-site"
	}
	sampleRatio := cfg.SampleRatio
	if sampleRatio <= 0 || sampleRatio > 1 {
		sampleRatio = 0.01
	}
	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		endpoint = "http://127.0.0.1:4318/v1/traces"
	}
	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(endpoint))
	if err != nil {
		return nil, fmt.Errorf("create OTLP trace exporter: %w", err)
	}
	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceName(serviceName)))
	if err != nil {
		return nil, fmt.Errorf("create trace resource: %w", err)
	}
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(sampleRatio))),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return &Tracing{Tracer: provider.Tracer(tracerName), Shutdown: provider.Shutdown}, nil
}

func (o *Observability) EnableTracing(tracing *Tracing) {
	if o == nil || tracing == nil {
		return
	}
	o.Tracer = tracing.Tracer
	o.TracingEnabled = true
}

func (o *Observability) WrapTrace(site string, next http.Handler) http.Handler {
	if o == nil || !o.TracingEnabled || next == nil {
		return next
	}
	return otelhttp.NewHandler(next, "goja-site.http.request",
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return r.Method + " " + CoarseRoute(r.URL.Path)
		}),
		otelhttp.WithSpanOptions(trace.WithAttributes(attribute.String("goja_site.site", SiteLabel(site)))),
	)
}
