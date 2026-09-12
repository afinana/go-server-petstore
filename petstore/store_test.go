package petstore

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestPlaceGetDeleteOrder(t *testing.T) {
	app := &Application{infoLog: log.New(io.Discard, "", 0), errorLog: log.New(io.Discard, "", 0)}

	// Ensure Orders is empty
	Orders = []Order{}

	// 1. Unauthenticated PlaceOrder should fail with 401
	o := Order{Id: 1, PetId: 1, Quantity: 2}
	b, _ := json.Marshal(o)
	unauthReq := httptest.NewRequest("POST", "/petstore/v2/store/order", bytes.NewReader(b))
	unauthRR := httptest.NewRecorder()
	app.PlaceOrder(unauthRR, unauthReq)
	if unauthRR.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for unauthenticated order, got %d", unauthRR.Code)
	}

	// 2. Authenticated PlaceOrder
	req := httptest.NewRequest("POST", "/petstore/v2/store/order", bytes.NewReader(b))
	req.Header.Set("X-User-Id", "alice")
	rr := httptest.NewRecorder()
	app.PlaceOrder(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("PlaceOrder expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	if len(Orders) != 1 {
		t.Fatalf("expected Orders length 1, got %d", len(Orders))
	}
	if Orders[0].UserId != "alice" {
		t.Fatalf("expected order UserId to be 'alice', got %q", Orders[0].UserId)
	}

	// 3. Business Logic Limit: Alice already has 2 pets in order.
	// Placing another order for 2 pets should exceed MaxAdoptionsPerUser (2 + 2 = 4 > 3)
	o2 := Order{Id: 2, PetId: 2, Quantity: 2}
	b2, _ := json.Marshal(o2)
	reqExcess := httptest.NewRequest("POST", "/petstore/v2/store/order", bytes.NewReader(b2))
	reqExcess.Header.Set("X-User-Id", "alice")
	rrExcess := httptest.NewRecorder()
	app.PlaceOrder(rrExcess, reqExcess)
	if rrExcess.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 Unprocessable Entity on exceeding adoption limit, got %d: %s",
			rrExcess.Code, rrExcess.Body.String())
	}

	// 4. Placing an order for 1 pet should succeed (2 + 1 = 3 <= 3)
	o3 := Order{Id: 3, PetId: 2, Quantity: 1}
	b3, _ := json.Marshal(o3)
	reqOk := httptest.NewRequest("POST", "/petstore/v2/store/order", bytes.NewReader(b3))
	reqOk.Header.Set("X-User-Id", "alice")
	rrOk := httptest.NewRecorder()
	app.PlaceOrder(rrOk, reqOk)
	if rrOk.Code != http.StatusOK {
		t.Fatalf("expected 200 when within adoption limit, got %d: %s", rrOk.Code, rrOk.Body.String())
	}

	// 5. GetInventory with Alice's context
	reqInv := httptest.NewRequest("GET", "/petstore/v2/store/inventory", nil)
	reqInv.Header.Set("X-User-Id", "alice")
	rrInv := httptest.NewRecorder()
	app.GetInventory(rrInv, reqInv)
	if rrInv.Code != http.StatusOK {
		t.Fatalf("GetInventory expected 200, got %d", rrInv.Code)
	}

	// 6. Fine-Grained Authorization: Bob tries to delete Alice's order (should fail with 403)
	reqBobDelete := httptest.NewRequest("DELETE", "/petstore/v2/store/order/1", nil)
	reqBobDelete.Header.Set("X-User-Id", "bob")
	reqBobDelete = mux.SetURLVars(reqBobDelete, map[string]string{"orderId": "1"})
	rrBobDelete := httptest.NewRecorder()
	app.DeleteOrder(rrBobDelete, reqBobDelete)
	if rrBobDelete.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when Bob deletes Alice's order, got %d", rrBobDelete.Code)
	}

	// 7. Alice deletes her own order (should succeed with 200)
	reqAliceDelete := httptest.NewRequest("DELETE", "/petstore/v2/store/order/1", nil)
	ctx := context.WithValue(reqAliceDelete.Context(), AuthUserKey, &AuthUser{ID: "alice"})
	reqAliceDelete = reqAliceDelete.WithContext(ctx)
	reqAliceDelete = mux.SetURLVars(reqAliceDelete, map[string]string{"orderId": "1"})
	rrAliceDelete := httptest.NewRecorder()
	app.DeleteOrder(rrAliceDelete, reqAliceDelete)
	if rrAliceDelete.Code != http.StatusOK {
		t.Fatalf("expected 200 when Alice deletes her own order, got %d", rrAliceDelete.Code)
	}
}
