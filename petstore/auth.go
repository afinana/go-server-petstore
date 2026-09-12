package petstore

import (
	"context"
	"crypto/subtle"
	"net/http"
	"os"
	"strings"
)

// Auth context key type
type contextKey string

const (
	AuthUserKey   = contextKey("authUser")
	UserClaimsKey = contextKey("userClaims")
)

// AuthUser represents an authenticated caller whose identity has been verified
// by the API gateway (e.g. KrakenD) and forwarded downstream.
type AuthUser struct {
	ID    string   `json:"id"`
	Roles []string `json:"roles"`
	Email string   `json:"email"`
}

// HasRole checks if the user possesses a specific role (case-insensitive).
func (u *AuthUser) HasRole(role string) bool {
	if u == nil {
		return false
	}
	for _, r := range u.Roles {
		if strings.EqualFold(strings.TrimSpace(r), role) {
			return true
		}
	}
	return false
}

// IsAdmin checks if the user has the 'admin' role.
func (u *AuthUser) IsAdmin() bool {
	return u.HasRole("admin")
}

// CanManageResource checks resource-level ownership (ABAC).
// An admin can manage any resource; a standard user can only manage
// resources they own or unowned legacy resources.
func (u *AuthUser) CanManageResource(resourceOwnerID string) bool {
	if u == nil {
		return false
	}
	if u.IsAdmin() {
		return true
	}
	if resourceOwnerID == "" {
		// If resource has no owner, allow access
		return true
	}
	return u.ID == resourceOwnerID
}

// GetAuthUser retrieves the authenticated user from the request context.
func GetAuthUser(ctx context.Context) (*AuthUser, bool) {
	if ctx == nil {
		return nil, false
	}
	user, ok := ctx.Value(AuthUserKey).(*AuthUser)
	return user, ok && user != nil && user.ID != ""
}

// AuthMiddleware extracts authenticated user identity forwarded by the API Gateway (KrakenD).
// When validation is enabled (default), it strictly validates that the request contains
// required gateway identity headers (X-User-Id, X-User-Roles, X-User-Email) and passes
// the optional gateway secret check (X-Gateway-Secret).
//
// When validation is disabled via configuration (ENABLE_AUTH_VALIDATION=false or app.SetAuthValidationEnabled(false)),
// requests without gateway headers are not rejected; instead, a default permissive development user is used.
func (app *Application) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		validationEnabled := app.IsAuthValidationEnabled()

		if validationEnabled {
			// 1. Defense-in-depth: verify gateway secret if configured
			expectedSecret := os.Getenv("GATEWAY_SECRET")
			if expectedSecret != "" {
				providedSecret := r.Header.Get("X-Gateway-Secret")
				if providedSecret == "" || subtle.ConstantTimeCompare([]byte(providedSecret), []byte(expectedSecret)) != 1 {
					http.Error(w, "Unauthorized: missing or invalid gateway secret", http.StatusUnauthorized)
					return
				}
			}

			// 2. Extract identity forwarded by the API Gateway
			userID := strings.TrimSpace(r.Header.Get("X-User-Id"))
			if userID == "" {
				userID = strings.TrimSpace(r.Header.Get("X-Forwarded-User"))
			}

			// If no user identity header is present, reject with 401 Unauthorized
			if userID == "" {
				http.Error(w, "Unauthorized: missing gateway identity headers", http.StatusUnauthorized)
				return
			}

			// 3. Extract roles
			roleHeader := r.Header.Get("X-User-Roles")
			if roleHeader == "" {
				roleHeader = r.Header.Get("X-User-Role")
			}
			var roles []string
			if roleHeader != "" {
				for _, role := range strings.Split(roleHeader, ",") {
					if trimmed := strings.TrimSpace(role); trimmed != "" {
						roles = append(roles, trimmed)
					}
				}
			}

			// 4. Extract email
			email := strings.TrimSpace(r.Header.Get("X-User-Email"))

			authUser := &AuthUser{
				ID:    userID,
				Roles: roles,
				Email: email,
			}

			// Store AuthUser in context
			ctx := context.WithValue(r.Context(), AuthUserKey, authUser)

			// Populate UserClaimsKey map for backwards compatibility
			claimsMap := map[string]interface{}{
				"sub":   userID,
				"roles": roles,
				"email": email,
			}
			ctx = context.WithValue(ctx, UserClaimsKey, claimsMap)

			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// --- Validation is DISABLED (Bypass / Development mode) ---
		userID := strings.TrimSpace(r.Header.Get("X-User-Id"))
		if userID == "" {
			userID = strings.TrimSpace(r.Header.Get("X-Forwarded-User"))
		}
		if userID == "" {
			userID = "dev-user"
		}

		roleHeader := r.Header.Get("X-User-Roles")
		if roleHeader == "" {
			roleHeader = r.Header.Get("X-User-Role")
		}
		var roles []string
		if roleHeader != "" {
			for _, role := range strings.Split(roleHeader, ",") {
				if trimmed := strings.TrimSpace(role); trimmed != "" {
					roles = append(roles, trimmed)
				}
			}
		}
		if len(roles) == 0 {
			roles = []string{"admin"} // Grant admin privileges when auth validation is disabled
		}

		email := strings.TrimSpace(r.Header.Get("X-User-Email"))
		if email == "" {
			email = "dev@local"
		}

		authUser := &AuthUser{
			ID:    userID,
			Roles: roles,
			Email: email,
		}

		ctx := context.WithValue(r.Context(), AuthUserKey, authUser)
		claimsMap := map[string]interface{}{
			"sub":   userID,
			"roles": roles,
			"email": email,
		}
		ctx = context.WithValue(ctx, UserClaimsKey, claimsMap)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
