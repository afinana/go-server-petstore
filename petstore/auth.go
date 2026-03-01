package petstore

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

// Auth context key type
type contextKey string

const (
	UserClaimsKey = contextKey("userClaims")
)

// jwks caches the Google public keys
var jwks keyfunc.Keyfunc

func init() {
	var err error
	googleCertsURL := "https://www.googleapis.com/oauth2/v3/certs"
	jwks, err = keyfunc.NewDefault([]string{googleCertsURL})
	if err != nil {
		log.Fatalf("Failed to create JWKS from resource at the given URL.\nError: %s", err.Error())
	}
}

// AuthMiddleware validates a Google JWT in the Authorization header.
func (app *Application) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, jwks.Keyfunc)

		if err != nil {
			log.Printf("Failed to parse the JWT.\nError: %s", err.Error())
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		// Validate issuer
		validIssuers := []string{"https://accounts.google.com", "accounts.google.com"}
		issuer, ok := claims["iss"].(string)
		if !ok {
			http.Error(w, "Missing issuer claim", http.StatusUnauthorized)
			return
		}

		isValidIssuer := false
		for _, v := range validIssuers {
			if issuer == v {
				isValidIssuer = true
				break
			}
		}

		if !isValidIssuer {
			http.Error(w, "Invalid token issuer", http.StatusUnauthorized)
			return
		}

		// Authentication is successful, store claims in context
		ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
