package petstore

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestApp() *Application {
	return &Application{infoLog: log.New(io.Discard, "", 0), errorLog: log.New(io.Discard, "", 0)}
}

func TestIndexHandler(t *testing.T) {
	app := newTestApp()
	req := httptest.NewRequest("GET", "/petstore/v2/", nil)
	rr := httptest.NewRecorder()

	app.Index(rr, req)

	if rr.Code != http.StatusOK && rr.Code != 0 {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

func TestUploadFileAndUpdateForm_NoDB(t *testing.T) {
	app := newTestApp()
	req := httptest.NewRequest("POST", "/petstore/v2/pet/1/uploadImage", nil)
	rr := httptest.NewRecorder()
	app.UploadFile(rr, req)
	if rr.Code != http.StatusOK && rr.Code != 0 {
		t.Fatalf("UploadFile expected 200, got %d", rr.Code)
	}

	req2 := httptest.NewRequest("POST", "/petstore/v2/pet/1", nil)
	rr2 := httptest.NewRecorder()
	app.UpdatePetWithForm(rr2, req2)
	if rr2.Code != http.StatusOK && rr2.Code != 0 {
		t.Fatalf("UpdatePetWithForm expected 200, got %d", rr2.Code)
	}
}

func TestUserSimpleHandlers(t *testing.T) {
	app := newTestApp()

	// LoginUser
	req := httptest.NewRequest("GET", "/petstore/v2/user/login", nil)
	rr := httptest.NewRecorder()
	app.LoginUser(rr, req)
	if rr.Code != http.StatusOK && rr.Code != 0 {
		t.Fatalf("LoginUser expected 200, got %d", rr.Code)
	}

	// LogoutUser
	req2 := httptest.NewRequest("GET", "/petstore/v2/user/logout", nil)
	rr2 := httptest.NewRecorder()
	app.LogoutUser(rr2, req2)
	if rr2.Code != http.StatusOK && rr2.Code != 0 {
		t.Fatalf("LogoutUser expected 200, got %d", rr2.Code)
	}

	// CreateUsersWithArrayInput and CreateUsersWithListInput should return 200
	req3 := httptest.NewRequest("POST", "/petstore/v2/user/createWithArray", bytes.NewReader([]byte("[]")))
	rr3 := httptest.NewRecorder()
	app.CreateUsersWithArrayInput(rr3, req3)
	if rr3.Code != http.StatusOK && rr3.Code != 0 {
		t.Fatalf("CreateUsersWithArrayInput expected 200, got %d", rr3.Code)
	}

	req4 := httptest.NewRequest("POST", "/petstore/v2/user/createWithList", bytes.NewReader([]byte("[]")))
	rr4 := httptest.NewRecorder()
	app.CreateUsersWithListInput(rr4, req4)
	if rr4.Code != http.StatusOK && rr4.Code != 0 {
		t.Fatalf("CreateUsersWithListInput expected 200, got %d", rr4.Code)
	}
}
