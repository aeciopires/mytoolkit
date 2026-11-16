package health

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func get(t *testing.T, mux http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestEndpoints(t *testing.T) {
	t.Cleanup(func() { shuttingDown.Store(false) })
	mux := http.NewServeMux()
	Register(mux)

	for path, want := range map[string]string{"/healthz": `{"status":"ok"}`, "/readyz": `{"status":"ready"}`} {
		rec := get(t, mux, path)
		if rec.Code != http.StatusOK || rec.Body.String() != want+"\n" {
			t.Errorf("%s = %d %q, want 200 %s", path, rec.Code, rec.Body.String(), want)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("%s Content-Type = %q", path, ct)
		}
	}

	MarkShuttingDown()
	if rec := get(t, mux, "/readyz"); rec.Code != http.StatusServiceUnavailable || rec.Body.String() != `{"status":"shutting_down"}`+"\n" {
		t.Errorf("/readyz while shutting down = %d %q, want 503", rec.Code, rec.Body.String())
	}
	if rec := get(t, mux, "/healthz"); rec.Code != http.StatusOK {
		t.Errorf("/healthz while shutting down = %d, want 200 (liveness must not restart a draining pod)", rec.Code)
	}
}
