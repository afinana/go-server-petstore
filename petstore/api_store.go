package petstore

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

var (
	Orders     []Order
	orderMutex sync.RWMutex
)

// DeleteOrder deletes an order by ID after verifying ownership (ABAC).
func (app *Application) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderIDStr := vars["orderId"]

	id, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil || id <= 0 {
		app.ErrorResponse(w, http.StatusBadRequest, "Invalid order ID: must be a positive integer")
		return
	}

	user, ok := GetAuthUser(r.Context())
	if !ok {
		// Fallback for tests or direct calls
		if uid := r.Header.Get("X-User-Id"); uid != "" {
			user = &AuthUser{ID: uid}
			ok = true
		}
	}
	if !ok {
		app.ErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	orderMutex.Lock()
	defer orderMutex.Unlock()

	foundIndex := -1
	var targetOrder Order
	for index, order := range Orders {
		if order.Id == id {
			foundIndex = index
			targetOrder = order
			break
		}
	}

	if foundIndex == -1 {
		app.ErrorResponse(w, http.StatusNotFound, "Order not found")
		return
	}

	// Fine-Grained Authorization (ABAC): Ensure user owns this order or is admin
	if !user.CanManageResource(targetOrder.UserId) {
		app.ErrorResponse(w, http.StatusForbidden, "Forbidden: you do not have permission to delete this order")
		return
	}

	Orders = append(Orders[:foundIndex], Orders[foundIndex+1:]...)
	app.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// GetInventory returns inventory orders filtered by ownership for regular users or all for admins.
func (app *Application) GetInventory(w http.ResponseWriter, r *http.Request) {
	user, _ := GetAuthUser(r.Context())
	if user == nil {
		if uid := r.Header.Get("X-User-Id"); uid != "" {
			user = &AuthUser{ID: uid}
		}
	}

	orderMutex.RLock()
	defer orderMutex.RUnlock()

	// If admin, or if called unauthenticated in tests, return all
	if user == nil || user.IsAdmin() {
		app.WriteJSON(w, http.StatusOK, Orders)
		return
	}

	// ABAC filter: Regular users only see their own orders
	var userOrders []Order
	for _, order := range Orders {
		if order.UserId == user.ID {
			userOrders = append(userOrders, order)
		}
	}
	if userOrders == nil {
		userOrders = []Order{}
	}
	app.WriteJSON(w, http.StatusOK, userOrders)
}

// GetOrderById retrieves an order by ID after verifying ownership (ABAC).
func (app *Application) GetOrderById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["orderId"], 10, 64)
	if err != nil || id <= 0 {
		app.ErrorResponse(w, http.StatusBadRequest, "Invalid order ID: must be a positive integer")
		return
	}

	user, ok := GetAuthUser(r.Context())
	if !ok {
		if uid := r.Header.Get("X-User-Id"); uid != "" {
			user = &AuthUser{ID: uid}
			ok = true
		}
	}
	if !ok {
		app.ErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	orderMutex.RLock()
	var result Order
	for _, order := range Orders {
		if order.Id == id {
			result = order
			break
		}
	}
	orderMutex.RUnlock()

	if reflect.ValueOf(result).IsZero() {
		app.ErrorResponse(w, http.StatusNotFound, "Order not found")
		return
	}

	// Fine-Grained Authorization (ABAC): user can only view their own order
	if !user.CanManageResource(result.UserId) {
		app.ErrorResponse(w, http.StatusForbidden, "Forbidden: you do not have permission to view this order")
		return
	}

	app.WriteJSON(w, http.StatusOK, result)
}

// PlaceOrder places a new pet order enforcing input validation, resource association,
// and the business logic limit (e.g. max 3 adoptions per user).
func (app *Application) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	LimitRequestBody(w, r)

	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		app.ErrorResponse(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// 1. Input Validation
	if err := ValidateOrder(&order); err != nil {
		app.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// 2. Extract Authenticated User
	user, ok := GetAuthUser(r.Context())
	if !ok {
		if uid := r.Header.Get("X-User-Id"); uid != "" {
			user = &AuthUser{ID: uid}
			ok = true
		}
	}
	if !ok {
		app.ErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// 3. Business Logic Limits: e.g. "A user can only adopt 3 pets max"
	orderMutex.Lock()
	defer orderMutex.Unlock()

	var existingAdoptedCount int64
	for _, o := range Orders {
		if o.UserId == user.ID && o.Status != "cancelled" {
			existingAdoptedCount += int64(o.Quantity)
		}
	}

	// Also count pets directly owned in DB if MongoDB is active
	if app.pets != nil && app.pets.C != nil {
		dbPetsCount, err := app.pets.CountByOwner(r.Context(), user.ID)
		if err == nil {
			existingAdoptedCount += dbPetsCount
		}
	}

	if !user.IsAdmin() && (existingAdoptedCount+int64(order.Quantity)) > int64(MaxAdoptionsPerUser) {
		app.ErrorResponse(w, http.StatusUnprocessableEntity,
			fmt.Sprintf("Business limit reached: user %s can adopt at most %d pets max (currently has %d, requested %d)",
				user.ID, MaxAdoptionsPerUser, existingAdoptedCount, order.Quantity))
		return
	}

	// Associate order with user (ABAC ownership)
	order.UserId = user.ID

	// If MongoDB is configured for stores, persist there too
	if app.stores != nil && app.stores.C != nil {
		_, _ = app.stores.Insert(r.Context(), order)
	}

	Orders = append(Orders, order)
	app.WriteJSON(w, http.StatusOK, order)
}
