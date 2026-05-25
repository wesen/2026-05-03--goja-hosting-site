package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCoarseRouteAvoidsRawDynamicPaths(t *testing.T) {
	tests := map[string]string{
		"/":                                "/",
		"/healthz":                         "/healthz",
		"/_kanban/client.js":               "/_kanban/client.js",
		"/_kanban/trail-notes/fragment":    "/_kanban/:board/fragment",
		"/_kanban/trail-notes/action/move": "/_kanban/:board/action/:action",
		"/assets/site.css":                 "/assets/*",
		"/customers/123?should-not-be-in-label=1": "other",
	}
	for path, want := range tests {
		if got := CoarseRoute(path); got != want {
			t.Fatalf("CoarseRoute(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestStatusClass(t *testing.T) {
	for status, want := range map[int]string{0: "unknown", 200: "2xx", 302: "3xx", 404: "4xx", 503: "5xx"} {
		if got := StatusClass(status); got != want {
			t.Fatalf("StatusClass(%d) = %q, want %q", status, got, want)
		}
	}
}

func TestStatusRecorderCapturesStatusAndBytes(t *testing.T) {
	rr := httptest.NewRecorder()
	rec := &StatusRecorder{ResponseWriter: rr}
	rec.WriteHeader(http.StatusCreated)
	if _, err := rec.Write([]byte("hello")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if got := rec.Status(); got != http.StatusCreated {
		t.Fatalf("status = %d, want %d", got, http.StatusCreated)
	}
	if got := rec.Bytes(); got != 5 {
		t.Fatalf("bytes = %d, want 5", got)
	}
}

func TestStartDiagnosticsRejectsPprofWithoutMetricsAddr(t *testing.T) {
	if _, err := StartDiagnostics(t.Context(), Config{EnablePprof: true}, nil); err == nil {
		t.Fatalf("expected pprof without metrics addr to fail")
	}
}
