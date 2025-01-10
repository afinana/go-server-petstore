/*
 * Swagger Petstore
 *
 * This is a sample server Petstore server.  You can find out more about Swagger at [http://swagger.io](http://swagger.io) or on [irc.freenode.net, #swagger](http://swagger.io/irc/).  For this sample, you can use the api key `special-key` to test the authorization filters.
 *
 * API version: 1.0.6
 * Contact: apiteam@swagger.io
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package petstore

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateUser adds a new user to the store
func (app *Application) CreateUser(w http.ResponseWriter, r *http.Request) {

	// Enable CORS
	if r.Method == "OPTIONS" {
		app.enableCors(&w, r)
		w.WriteHeader(http.StatusOK)
		return
	}
	app.enableCors(&w, r)

	// Define User model
	var m User
	// Get request information
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		app.serverError(w, err)
	}

	// Insert new Users
	insertResult, err := app.users.Insert(m)
	if err != nil {
		app.serverError(w, err)
	}
	m.ID = insertResult.InsertedID.(primitive.ObjectID)

	app.infoLog.Printf("New user have been created, id=%s", insertResult.InsertedID)
	w.Header().Set("Content-Type", "Application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func (app *Application) CreateUsersWithArrayInput(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json; charset=UTF-8")
	app.enableCors(&w, r)
	w.WriteHeader(http.StatusOK)
}

func (app *Application) CreateUsersWithListInput(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json; charset=UTF-8")
	app.enableCors(&w, r)
	w.WriteHeader(http.StatusOK)
}

func (app *Application) DeleteUser(w http.ResponseWriter, r *http.Request) {

	if r.Method == "OPTIONS" {
		app.enableCors(&w, r)
		w.WriteHeader(http.StatusOK)
		return
	}
	// Get id from incoming url
	vars := mux.Vars(r)
	id := vars["id"]

	// Delete Users by id
	deleteResult, err := app.users.Delete(id)
	if err != nil {
		app.serverError(w, err)
	}

	app.infoLog.Printf("Have been eliminated %d user(s)", deleteResult.DeletedCount)
	w.Header().Set("Content-Type", "Application/json; charset=UTF-8")
	app.enableCors(&w, r)
	w.WriteHeader(http.StatusOK)

}

func (app *Application) GetUserByName(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	name := vars["username"]
	fmt.Printf("GetUserByName name: %s\n", name)

	result, err := app.users.FindByUserName(name)
	if err != nil {
		if err.Error() == "ErrNoDocuments" {
			app.infoLog.Printf("User not found")
			w.WriteHeader(http.StatusNotFound)
		} else {
			app.serverError(w, err)
		}
	}

	w.Header().Set("Content-Type", "Application/json; charset=UTF-8")
	app.enableCors(&w, r)
	json.NewEncoder(w).Encode(result)
	w.WriteHeader(http.StatusOK)
}

func (app *Application) LoginUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json; charset=UTF-8")
	app.enableCors(&w, r)
	w.WriteHeader(http.StatusOK)
}

func (app *Application) LogoutUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json; charset=UTF-8")
	app.enableCors(&w, r)
	w.WriteHeader(http.StatusOK)
}

func (app *Application) UpdateUser(w http.ResponseWriter, r *http.Request) {

	if r.Method == "OPTIONS" {
		app.enableCors(&w, r)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Define User model
	var m User
	// Get request information
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		app.serverError(w, err)
	}

	// Update Users
	updateResult, err := app.users.Update(m.ID.String(), m)
	if err != nil {
		app.serverError(w, err)
	}

	app.infoLog.Printf("User have been updated, id=%s", updateResult.UpsertedID)
	w.Header().Set("Content-Type", "Application/json; charset=UTF-8")
	app.enableCors(&w, r)
	w.WriteHeader(http.StatusOK)
}

// create get all users
func (app *Application) GetAllUsers(w http.ResponseWriter, r *http.Request) {

	result, err := app.users.All()
	if err != nil {
		app.serverError(w, err)
	}

	w.Header().Set("Content-Type", "Application/json; charset=UTF-8")
	app.enableCors(&w, r)
	json.NewEncoder(w).Encode(result)
	w.WriteHeader(http.StatusOK)
}
