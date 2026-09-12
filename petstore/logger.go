package petstore

import (
	"log"
	"net/http"
	"strings"
	"time"
)

func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		// Sanitize variables to prevent CRLF log injection (G706)
		safeURI := strings.ReplaceAll(strings.ReplaceAll(r.RequestURI, "\n", ""), "\r", "")
		safeMethod := strings.ReplaceAll(strings.ReplaceAll(r.Method, "\n", ""), "\r", "")

		log.Printf(
			"%s%s%s %s%s%s %s %s%s%s",
			Green, safeMethod, Reset,
			Cyan, safeURI, Reset,
			name,
			Yellow, time.Since(start), Reset,
		)
	})
}
