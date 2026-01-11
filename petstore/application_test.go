package petstore

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEnableCorsSetsHeaders(t *testing.T) {
	app := &Application{infoLog: log.New(io.Discard, "", 0), errorLog: log.New(io.Discard, "", 0)}

	req := httptest.NewRequest("GET", "http://example.local/", nil)
	req.Header.Set("Origin", "http://origin.test")
	rr := httptest.NewRecorder()

	// enableCors expects an http.ResponseWriter
	app.enableCors(rr, req)

	got := rr.Header().Get("Access-Control-Allow-Origin")
	if got != "http://origin.test" {
		t.Fatalf("expected Access-Control-Allow-Origin header set, got %q", got)
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
