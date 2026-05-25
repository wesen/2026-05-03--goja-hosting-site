package observability

import "testing"

func TestInitTracingDisabledReturnsNoopShutdown(t *testing.T) {
	tracing, err := InitTracing(t.Context(), TracingConfig{})
	if err != nil {
		t.Fatalf("InitTracing disabled error = %v", err)
	}
	if tracing == nil || tracing.Tracer == nil || tracing.Shutdown == nil {
		t.Fatalf("disabled tracing did not return complete no-op tracing handle")
	}
	if err := tracing.Shutdown(t.Context()); err != nil {
		t.Fatalf("disabled tracing shutdown error = %v", err)
	}
}
