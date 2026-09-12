package petstore

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestAuthMiddleware_NoGatewayHeaders(t *testing.T) {
	app := &Application{}

	handler := app.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req, err := http.NewRequest("GET", "/v2/pet", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_ValidUserHeader(t *testing.T) {
	app := &Application{}

	var capturedUser *AuthUser
	handler := app.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser, _ = GetAuthUser(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req, err := http.NewRequest("GET", "/v2/pet", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-User-Id", "alice123")
	req.Header.Set("X-User-Roles", "user,pet-owner")
	req.Header.Set("X-User-Email", "alice@example.com")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	if capturedUser == nil || capturedUser.ID != "alice123" {
		t.Errorf("expected user alice123, got %+v", capturedUser)
	}
	if !capturedUser.HasRole("user") || !capturedUser.HasRole("pet-owner") {
		t.Errorf("expected roles user and pet-owner, got %v", capturedUser.Roles)
	}
	if capturedUser.IsAdmin() {
		t.Errorf("expected IsAdmin to be false")
	}
}

func TestAuthMiddleware_GatewaySecretCheck(t *testing.T) {
	t.Setenv("GATEWAY_SECRET", "super-secret-token")

	app := &Application{}

	handler := app.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Missing gateway secret should fail even with X-User-Id
	req1 := httptest.NewRequest("GET", "/v2/pet", nil)
	req1.Header.Set("X-User-Id", "alice123")
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 on missing gateway secret, got %d", rr1.Code)
	}

	// Valid gateway secret should succeed
	req2 := httptest.NewRequest("GET", "/v2/pet", nil)
	req2.Header.Set("X-User-Id", "alice123")
	req2.Header.Set("X-Gateway-Secret", "super-secret-token")
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Errorf("expected 200 on valid gateway secret, got %d", rr2.Code)
	}
}

func TestAuthMiddleware_DisabledValidation(t *testing.T) {
	app := &Application{}
	app.SetAuthValidationEnabled(false)

	if app.IsAuthValidationEnabled() {
		t.Fatalf("expected IsAuthValidationEnabled to return false")
	}

	var capturedUser *AuthUser
	handler := app.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser, _ = GetAuthUser(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	// Request with NO headers should succeed when validation is disabled
	req := httptest.NewRequest("GET", "/v2/pet", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK when auth validation is disabled, got %d", rr.Code)
	}
	if capturedUser == nil || capturedUser.ID != "dev-user" {
		t.Fatalf("expected default dev-user, got %+v", capturedUser)
	}
	if !capturedUser.IsAdmin() {
		t.Fatalf("expected dev-user to have admin role when validation is disabled")
	}
}

func TestAuthMiddleware_EnvDisabledValidation(t *testing.T) {
	t.Setenv("ENABLE_AUTH_VALIDATION", "false")

	app := &Application{} // Does not explicitly set enableAuthValidation, relies on env var

	if app.IsAuthValidationEnabled() {
		t.Fatalf("expected IsAuthValidationEnabled to return false when ENABLE_AUTH_VALIDATION=false")
	}

	handler := app.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/v2/pet", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK when ENABLE_AUTH_VALIDATION=false, got %d", rr.Code)
	}
}

func TestAuthMiddleware_EnabledByDefault(t *testing.T) {
	_ = os.Unsetenv("ENABLE_AUTH_VALIDATION")
	_ = os.Unsetenv("ENABLE_HEADER_VALIDATION")

	app := &Application{}
	if !app.IsAuthValidationEnabled() {
		t.Fatalf("expected IsAuthValidationEnabled to default to true")
	}

	cfg := LoadConfig()
	if !cfg.EnableAuthValidation {
		t.Fatalf("expected LoadConfig().EnableAuthValidation to default to true")
	}
}

func TestAuthUser_CanManageResource(t *testing.T) {
	user := &AuthUser{ID: "alice", Roles: []string{"user"}}
	admin := &AuthUser{ID: "admin-1", Roles: []string{"admin"}}

	if !user.CanManageResource("alice") {
		t.Errorf("expected alice to manage resource owned by alice")
	}
	if user.CanManageResource("bob") {
		t.Errorf("expected alice NOT to manage resource owned by bob")
	}
	if !admin.CanManageResource("bob") {
		t.Errorf("expected admin to manage resource owned by bob")
	}
	if !user.CanManageResource("") {
		t.Errorf("expected user to manage unowned legacy resource")
	}
}
