package petstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// AddPet adds a new pet to the store with input validation, gateway identity association,
// and business logic limit checks.
func (app *Application) AddPet(w http.ResponseWriter, r *http.Request) {
	LimitRequestBody(w, r)

	var m Pet
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		app.ErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
		return
	}

	// 1. Input Validation: strings, lengths, allowed enums, URL formats
	if err := ValidatePet(&m); err != nil {
		app.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// 2. Extract Authenticated User from Gateway context
	user, ok := GetAuthUser(r.Context())
	if !ok {
		app.ErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// 3. Business Logic Limits: e.g. "A user can only adopt/own 3 pets max"
	if app.pets != nil && app.pets.C != nil {
		ownedCount, err := app.pets.CountByOwner(r.Context(), user.ID)
		if err != nil {
			app.serverError(w, err)
			return
		}
		if ownedCount >= MaxAdoptionsPerUser && !user.IsAdmin() {
			app.ErrorResponse(w, http.StatusUnprocessableEntity,
				fmt.Sprintf("Business limit reached: user %s already owns %d pets (maximum allowed is %d)",
					user.ID, ownedCount, MaxAdoptionsPerUser))
			return
		}
	}

	// Set resource owner
	m.OwnerID = user.ID

	// Insert new Pet into MongoDB
	if app.pets != nil && app.pets.C != nil {
		insertResult, err := app.pets.Insert(r.Context(), m)
		if err != nil {
			app.serverError(w, err)
			return
		}
		m.ID = insertResult.InsertedID.(primitive.ObjectID)
		app.infoLog.Printf("New pet created with ID=%s, Owner=%s", m.ID.Hex(), user.ID)
	}

	app.WriteJSON(w, http.StatusCreated, m)
}

// DeletePet deletes a pet by ID after verifying resource ownership (ABAC).
func (app *Application) DeletePet(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["petId"]

	// 1. Input Validation on Path Parameter
	if err := ValidateID(id); err != nil {
		app.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// 2. Extract Authenticated User
	user, ok := GetAuthUser(r.Context())
	if !ok {
		app.ErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	if app.pets != nil && app.pets.C != nil {
		// 3. Fine-Grained Authorization (ABAC): Check object ownership
		existingPet, err := app.pets.FindByIDOrHex(r.Context(), id)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				app.ErrorResponse(w, http.StatusNotFound, "Pet not found")
				return
			}
			app.serverError(w, err)
			return
		}

		// Deny if user is not the owner and not an admin
		if !user.CanManageResource(existingPet.OwnerID) {
			app.ErrorResponse(w, http.StatusForbidden, "Forbidden: you do not have permission to delete this pet")
			return
		}

		// Perform deletion
		deleteResult, err := app.pets.Delete(r.Context(), id)
		if err != nil {
			app.serverError(w, err)
			return
		}
		app.infoLog.Printf("Pet %s eliminated (%d document)", id, deleteResult.DeletedCount)
	}

	w.WriteHeader(http.StatusNoContent)
}

// FindPetsByStatus searches pets by status with sanitized inputs.
func (app *Application) FindPetsByStatus(w http.ResponseWriter, r *http.Request) {
	statusQuery := r.URL.Query().Get("status")
	if strings.TrimSpace(statusQuery) == "" {
		app.WriteJSON(w, http.StatusOK, []Pet{})
		return
	}

	status := strings.Split(statusQuery, ",")
	if app.pets == nil || app.pets.C == nil {
		app.WriteJSON(w, http.StatusOK, []Pet{})
		return
	}

	model, err := app.pets.FindByStatus(r.Context(), status)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			app.WriteJSON(w, http.StatusOK, []Pet{})
			return
		}
		app.serverError(w, err)
		return
	}

	if model == nil {
		model = []Pet{}
	}
	app.WriteJSON(w, http.StatusOK, model)
}

// FindPetsByTags searches pets by tags with sanitized inputs.
func (app *Application) FindPetsByTags(w http.ResponseWriter, r *http.Request) {
	tagQuery := r.URL.Query().Get("tags")
	if strings.TrimSpace(tagQuery) == "" {
		app.WriteJSON(w, http.StatusOK, []Pet{})
		return
	}

	tags := strings.Split(tagQuery, ",")
	if app.pets == nil || app.pets.C == nil {
		app.WriteJSON(w, http.StatusOK, []Pet{})
		return
	}

	model, err := app.pets.FindByTags(r.Context(), tags)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			app.WriteJSON(w, http.StatusOK, []Pet{})
			return
		}
		app.serverError(w, err)
		return
	}

	if model == nil {
		model = []Pet{}
	}
	app.WriteJSON(w, http.StatusOK, model)
}

// GetPetById retrieves a pet by ID with validated format.
func (app *Application) GetPetById(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["petId"]

	// Validate path parameter
	if err := ValidateID(id); err != nil {
		app.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if app.pets == nil || app.pets.C == nil {
		app.ErrorResponse(w, http.StatusNotFound, "Pet not found")
		return
	}

	model, err := app.pets.FindByIDOrHex(r.Context(), id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			app.ErrorResponse(w, http.StatusNotFound, "Pet not found")
			return
		}
		app.serverError(w, err)
		return
	}

	app.WriteJSON(w, http.StatusOK, model)
}

// UpdatePet updates an existing pet after validating input and verifying object ownership (ABAC).
func (app *Application) UpdatePet(w http.ResponseWriter, r *http.Request) {
	LimitRequestBody(w, r)

	var m Pet
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		app.ErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
		return
	}

	// 1. Input validation
	if err := ValidatePet(&m); err != nil {
		app.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// 2. Extract Authenticated User
	user, ok := GetAuthUser(r.Context())
	if !ok {
		app.ErrorResponse(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	if app.pets != nil && app.pets.C != nil {
		// Identify pet to update
		targetID := m.ID.Hex()
		if targetID == "" || targetID == primitive.NilObjectID.Hex() {
			if m.Id > 0 {
				targetID = strconv.FormatInt(m.Id, 10)
			}
		}

		if targetID != "" && targetID != primitive.NilObjectID.Hex() {
			existingPet, err := app.pets.FindByIDOrHex(r.Context(), targetID)
			if err != nil {
				if errors.Is(err, mongo.ErrNoDocuments) {
					app.ErrorResponse(w, http.StatusNotFound, "Pet not found")
					return
				}
				app.serverError(w, err)
				return
			}

			// 3. Fine-Grained Authorization (ABAC): Check object ownership
			if !user.CanManageResource(existingPet.OwnerID) {
				app.ErrorResponse(w, http.StatusForbidden, "Forbidden: you do not have permission to edit this pet")
				return
			}

			// Preserve existing owner unless admin
			if !user.IsAdmin() {
				m.OwnerID = existingPet.OwnerID
			}
			m.ID = existingPet.ID
		}

		// Perform update
		updateResult, err := app.pets.Update(r.Context(), m)
		if err != nil {
			app.serverError(w, err)
			return
		}
		app.infoLog.Printf("Pet updated: matched=%d, modified=%d", updateResult.MatchedCount, updateResult.ModifiedCount)
	}

	app.WriteJSON(w, http.StatusOK, m)
}

func (app *Application) UpdatePetWithForm(w http.ResponseWriter, _ *http.Request) {
	app.WriteJSON(w, http.StatusOK, map[string]string{"status": "not implemented"})
}

func (app *Application) UploadFile(w http.ResponseWriter, _ *http.Request) {
	app.WriteJSON(w, http.StatusOK, map[string]string{"status": "not implemented"})
}
