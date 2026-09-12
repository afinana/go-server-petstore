package petstore

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestInputValidation_Pet(t *testing.T) {
	// 1. Valid Pet
	validPet := &Pet{
		Name:      "Fluffy",
		Status:    "available",
		PhotoUrls: []string{"https://example.com/photos/fluffy.jpg"},
		Tags:      []Tag{{Name: "friendly"}},
	}
	if err := ValidatePet(validPet); err != nil {
		t.Fatalf("expected valid pet to pass validation, got: %v", err)
	}

	// 2. Empty Name
	emptyNamePet := &Pet{Name: "   ", Status: "available"}
	if err := ValidatePet(emptyNamePet); err == nil {
		t.Fatalf("expected error for empty pet name, got nil")
	}

	// 3. Excessive Name Length (> 100 characters)
	longNamePet := &Pet{Name: strings.Repeat("A", 101), Status: "available"}
	if err := ValidatePet(longNamePet); err == nil {
		t.Fatalf("expected error for overly long pet name, got nil")
	}

	// 4. Invalid Status
	invalidStatusPet := &Pet{Name: "Buddy", Status: "malicious_status_injection"}
	if err := ValidatePet(invalidStatusPet); err == nil {
		t.Fatalf("expected error for invalid pet status, got nil")
	}

	// 5. Invalid Photo URL (javascript: URI scheme)
	maliciousURLPet := &Pet{
		Name:      "Max",
		PhotoUrls: []string{"javascript:alert('xss')"},
	}
	if err := ValidatePet(maliciousURLPet); err == nil {
		t.Fatalf("expected error for javascript: URL, got nil")
	}

	// 6. Overly Long Tag Name (> 50 characters)
	longTagPet := &Pet{
		Name: "Bella",
		Tags: []Tag{{Name: strings.Repeat("T", 51)}},
	}
	if err := ValidatePet(longTagPet); err == nil {
		t.Fatalf("expected error for overly long tag name, got nil")
	}
}

func TestInputValidation_Order(t *testing.T) {
	// 1. Valid Order
	validOrder := &Order{PetId: 10, Quantity: 2, Status: "placed"}
	if err := ValidateOrder(validOrder); err != nil {
		t.Fatalf("expected valid order to pass, got: %v", err)
	}

	// 2. PetId <= 0
	badPetIdOrder := &Order{PetId: 0, Quantity: 1}
	if err := ValidateOrder(badPetIdOrder); err == nil {
		t.Fatalf("expected error for petId <= 0, got nil")
	}

	// 3. Quantity <= 0
	zeroQtyOrder := &Order{PetId: 1, Quantity: 0}
	if err := ValidateOrder(zeroQtyOrder); err == nil {
		t.Fatalf("expected error for quantity <= 0, got nil")
	}

	// 4. Quantity exceeds max adoption limit (e.g. > 3)
	excessQtyOrder := &Order{PetId: 1, Quantity: 4}
	if err := ValidateOrder(excessQtyOrder); err == nil {
		t.Fatalf("expected error for quantity exceeding adoption limit, got nil")
	}

	// 5. Invalid Order Status
	invalidStatusOrder := &Order{PetId: 1, Quantity: 1, Status: "hacked"}
	if err := ValidateOrder(invalidStatusOrder); err == nil {
		t.Fatalf("expected error for invalid order status, got nil")
	}
}

func TestInputValidation_ValidateID(t *testing.T) {
	// Valid numeric ID
	if err := ValidateID("12345"); err != nil {
		t.Fatalf("expected valid numeric ID, got %v", err)
	}

	// Valid 24-char hex ObjectID
	if err := ValidateID("507f1f77bcf86cd799439011"); err != nil {
		t.Fatalf("expected valid hex ObjectID, got %v", err)
	}

	// Invalid empty ID
	if err := ValidateID("   "); err == nil {
		t.Fatalf("expected error for empty ID, got nil")
	}

	// Invalid string containing injection / traversal
	if err := ValidateID("../invalid-id"); err == nil {
		t.Fatalf("expected error for invalid ID '../invalid-id', got nil")
	}
}

func TestFineGrainedAuthorization_ResourceOwnership(t *testing.T) {
	userAlice := &AuthUser{ID: "alice", Roles: []string{"user"}}
	userBob := &AuthUser{ID: "bob", Roles: []string{"user"}}
	userAdmin := &AuthUser{ID: "admin_user", Roles: []string{"admin"}}

	// Alice owns resource "res-123"
	resourceOwner := "alice"

	// 1. Alice should be authorized to manage her own resource
	if !userAlice.CanManageResource(resourceOwner) {
		t.Errorf("Alice should be allowed to manage her own resource")
	}

	// 2. Bob should NOT be authorized to manage Alice's resource
	if userBob.CanManageResource(resourceOwner) {
		t.Errorf("Bob must NOT be allowed to manage Alice's resource")
	}

	// 3. Admin should be authorized to manage any resource
	if !userAdmin.CanManageResource(resourceOwner) {
		t.Errorf("Admin must be allowed to manage any resource")
	}
}

func TestBusinessLogicLimits_Adoption(t *testing.T) {
	app := &Application{infoLog: log.New(io.Discard, "", 0), errorLog: log.New(io.Discard, "", 0)}
	Orders = []Order{}

	// User Charlie attempts to place orders
	// Order 1: adopt 2 pets
	o1 := Order{Id: 101, PetId: 1, Quantity: 2}
	b1, _ := json.Marshal(o1)
	req1 := httptest.NewRequest("POST", "/petstore/v2/store/order", bytes.NewReader(b1))
	req1.Header.Set("X-User-Id", "charlie")
	rr1 := httptest.NewRecorder()
	app.PlaceOrder(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Fatalf("expected order 1 to succeed, got %d: %s", rr1.Code, rr1.Body.String())
	}

	// Order 2: adopt 2 more pets (2 + 2 = 4 > 3 -> should fail)
	o2 := Order{Id: 102, PetId: 2, Quantity: 2}
	b2, _ := json.Marshal(o2)
	req2 := httptest.NewRequest("POST", "/petstore/v2/store/order", bytes.NewReader(b2))
	req2.Header.Set("X-User-Id", "charlie")
	rr2 := httptest.NewRecorder()
	app.PlaceOrder(rr2, req2)
	if rr2.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 Unprocessable Entity for exceeding max 3 adoptions, got %d", rr2.Code)
	}

	// Order 3: adopt 1 pet (2 + 1 = 3 <= 3 -> should succeed)
	o3 := Order{Id: 103, PetId: 3, Quantity: 1}
	b3, _ := json.Marshal(o3)
	req3 := httptest.NewRequest("POST", "/petstore/v2/store/order", bytes.NewReader(b3))
	req3.Header.Set("X-User-Id", "charlie")
	rr3 := httptest.NewRecorder()
	app.PlaceOrder(rr3, req3)
	if rr3.Code != http.StatusOK {
		t.Fatalf("expected order 3 to succeed within limit, got %d: %s", rr3.Code, rr3.Body.String())
	}

	// Order 4: adopt 1 more pet (now has 3 -> any additional order fails)
	o4 := Order{Id: 104, PetId: 4, Quantity: 1}
	b4, _ := json.Marshal(o4)
	req4 := httptest.NewRequest("POST", "/petstore/v2/store/order", bytes.NewReader(b4))
	req4.Header.Set("X-User-Id", "charlie")
	rr4 := httptest.NewRecorder()
	app.PlaceOrder(rr4, req4)
	if rr4.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 when exceeding 3 adoptions, got %d", rr4.Code)
	}
}

func TestMalformedJSONPayloadReturnsBadRequest(t *testing.T) {
	app := &Application{infoLog: log.New(io.Discard, "", 0), errorLog: log.New(io.Discard, "", 0)}

	// Send malformed JSON to AddPet
	req := httptest.NewRequest("POST", "/petstore/v2/pet", bytes.NewReader([]byte("{invalid-json-payload")))
	ctx := context.WithValue(req.Context(), AuthUserKey, &AuthUser{ID: "test-user"})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	app.AddPet(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request on malformed JSON, got %d", rr.Code)
	}
}

func TestInvalidIDPathParameterReturnsBadRequest(t *testing.T) {
	app := &Application{infoLog: log.New(io.Discard, "", 0), errorLog: log.New(io.Discard, "", 0)}

	req := httptest.NewRequest("GET", "/petstore/v2/pet/not-an-id!", nil)
	req = mux.SetURLVars(req, map[string]string{"petId": "not-an-id!"})
	rr := httptest.NewRecorder()
	app.GetPetById(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request on invalid petId format, got %d", rr.Code)
	}
}
