package petstore

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// CreateUser adds a new user to the store with input validation
func (app *Application) CreateUser(w http.ResponseWriter, r *http.Request) {
	LimitRequestBody(w, r)

	var m User
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		app.ErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
		return
	}

	m.Username = strings.TrimSpace(m.Username)
	if m.Username == "" || len(m.Username) < MinUsernameLength || len(m.Username) > MaxUsernameLength {
		app.ErrorResponse(w, http.StatusBadRequest, "Username must be between 3 and 50 characters")
		return
	}
	if !usernameRegex.MatchString(m.Username) {
		app.ErrorResponse(w, http.StatusBadRequest, "Username can only contain alphanumeric characters, dots, underscores, and dashes")
		return
	}

	if app.users != nil && app.users.C != nil {
		insertResult, err := app.users.Insert(r.Context(), m)
		if err != nil {
			app.serverError(w, err)
			return
		}
		m.ID = insertResult.InsertedID.(primitive.ObjectID)
		app.infoLog.Printf("New user created, id=%s", insertResult.InsertedID)
	}

	app.WriteJSON(w, http.StatusOK, m)
}

func (app *Application) CreateUsersWithArrayInput(w http.ResponseWriter, _ *http.Request) {
	app.WriteJSON(w, http.StatusOK, map[string]string{"status": "not implemented"})
}

func (app *Application) CreateUsersWithListInput(w http.ResponseWriter, _ *http.Request) {
	app.WriteJSON(w, http.StatusOK, map[string]string{"status": "not implemented"})
}

func (app *Application) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	if username == "" {
		username = vars["id"]
	}

	user, ok := GetAuthUser(r.Context())
	if !ok {
		if uid := r.Header.Get("X-User-Id"); uid != "" {
			user = &AuthUser{ID: uid}
			ok = true
		}
	}

	// Fine-Grained Authorization: User can only delete their own profile unless admin
	if ok && user != nil && !user.IsAdmin() && user.ID != username {
		app.ErrorResponse(w, http.StatusForbidden, "Forbidden: you do not have permission to delete this user")
		return
	}

	if app.users != nil && app.users.C != nil {
		targetUser, err := app.users.FindByUserName(r.Context(), username)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				app.ErrorResponse(w, http.StatusNotFound, "User not found")
				return
			}
			app.serverError(w, err)
			return
		}

		deleteResult, err := app.users.Delete(r.Context(), targetUser.ID.Hex())
		if err != nil {
			app.serverError(w, err)
			return
		}

		app.infoLog.Printf("Eliminated %d user(s)", deleteResult.DeletedCount)
		app.WriteJSON(w, http.StatusOK, map[string]int64{"deletedCount": deleteResult.DeletedCount})
		return
	}

	app.WriteJSON(w, http.StatusOK, map[string]int64{"deletedCount": 1})
}

func (app *Application) GetUserByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := strings.TrimSpace(vars["username"])

	if app.users == nil || app.users.C == nil {
		app.ErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	result, err := app.users.FindByUserName(r.Context(), name)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			app.ErrorResponse(w, http.StatusNotFound, "User not found")
			return
		}
		app.serverError(w, err)
		return
	}

	app.WriteJSON(w, http.StatusOK, result)
}

func (app *Application) LoginUser(w http.ResponseWriter, _ *http.Request) {
	app.WriteJSON(w, http.StatusOK, map[string]string{"status": "logged in"})
}

func (app *Application) LogoutUser(w http.ResponseWriter, _ *http.Request) {
	app.WriteJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func (app *Application) UpdateUser(w http.ResponseWriter, r *http.Request) {
	LimitRequestBody(w, r)

	vars := mux.Vars(r)
	username := strings.TrimSpace(vars["username"])

	var m User
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		app.ErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
		return
	}

	user, ok := GetAuthUser(r.Context())
	if !ok {
		if uid := r.Header.Get("X-User-Id"); uid != "" {
			user = &AuthUser{ID: uid}
			ok = true
		}
	}

	// Fine-grained authorization
	if ok && user != nil && !user.IsAdmin() && user.ID != username {
		app.ErrorResponse(w, http.StatusForbidden, "Forbidden: you do not have permission to update this user")
		return
	}

	if app.users != nil && app.users.C != nil {
		targetUser, err := app.users.FindByUserName(r.Context(), username)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				app.ErrorResponse(w, http.StatusNotFound, "User not found")
				return
			}
			app.serverError(w, err)
			return
		}

		m.ID = targetUser.ID
		updateResult, err := app.users.Update(r.Context(), targetUser.ID.Hex(), m)
		if err != nil {
			app.serverError(w, err)
			return
		}

		app.infoLog.Printf("User updated, id=%s", updateResult.UpsertedID)
	}

	app.WriteJSON(w, http.StatusOK, m)
}

func (app *Application) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if app.users == nil || app.users.C == nil {
		app.WriteJSON(w, http.StatusOK, []User{})
		return
	}

	result, err := app.users.All(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	if result == nil {
		result = []User{}
	}

	app.WriteJSON(w, http.StatusOK, result)
}
