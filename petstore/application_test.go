package petstore

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMiddlewareSetsHeaders(t *testing.T) {
	app := &Application{infoLog: log.New(io.Discard, "", 0), errorLog: log.New(io.Discard, "", 0)}

	nextHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})
	handlerToTest := app.Middleware(nextHandler)

	req := httptest.NewRequest("GET", "http://example.local/", nil)
	req.Header.Set("Origin", "http://origin.test")
	rr := httptest.NewRecorder()

	handlerToTest.ServeHTTP(rr, req)

	got := rr.Header().Get("Access-Control-Allow-Origin")
	if got != "http://origin.test" {
		t.Fatalf("expected Access-Control-Allow-Origin header set, got %q", got)
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		t.Fatalf("expected Content-Type header set, got %q", contentType)
	}
}

func TestServerErrorWrites500(t *testing.T) {
	app := &Application{infoLog: log.New(io.Discard, "", 0), errorLog: log.New(io.Discard, "", 0)}
	rr := httptest.NewRecorder()
	// call serverError with a sample error
	app.serverError(rr, &simpleErr{"boom"})

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
}

type simpleErr struct{ s string }

func (e *simpleErr) Error() string { return e.s }
