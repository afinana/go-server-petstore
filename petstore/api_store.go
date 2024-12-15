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
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gorilla/mux"
)

var Orders []Order

func (app *Application) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	if r.Method == "OPTIONS" {
		app.enableCors(&w, r)
		w.WriteHeader(http.StatusOK)
		return
	}
	app.enableCors(&w, r)

	vars := mux.Vars(r)

	fmt.Printf("DeleteOrder::id is %s\n", vars["orderId"])
	id, err := strconv.ParseInt(vars["orderId"], 10, 64) // Changed 32 to 64
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	for index, order := range Orders {
		if order.ID == id { // Changed int(order.ID) == int(id) to order.ID == id
			Orders = append(Orders[:index], Orders[index+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	http.Error(w, "Order not found", http.StatusNotFound)
}

func (app *Application) GetInventory(w http.ResponseWriter, r *http.Request) {

	// Enable CORS
	if r.Method == "OPTIONS" {
		app.enableCors(&w, r)
		w.WriteHeader(http.StatusOK)
		return
	}
	app.enableCors(&w, r)

	fmt.Println("GetInventory:: return all orders")
	json.NewEncoder(w).Encode(Orders)

	w.Header().Set("Content-Type", "Application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

}

func (app *Application) GetOrderById(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	if r.Method == "OPTIONS" {
		app.enableCors(&w, r)
		w.WriteHeader(http.StatusOK)
		return
	}
	app.enableCors(&w, r)

	var result Order
	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["orderId"], 10, 32)
	if err != nil {

		panic(err)
	}
	fmt.Printf("GetOrderById id: %d\n", id)

	for _, order := range Orders {
		if order.ID == id {
			result = order

		}
	}

	w.Header().Set("Content-Type", "Application/json; charset=UTF-8")
	if reflect.ValueOf(result).IsZero() {
		w.WriteHeader(http.StatusNotFound)
	} else {
		json.NewEncoder(w).Encode(result)

	}
}

func (app *Application) PlaceOrder(w http.ResponseWriter, r *http.Request) {

	// Enable CORS
	if r.Method == "OPTIONS" {
		app.enableCors(&w, r)
		w.WriteHeader(http.StatusOK)
		return
	}
	app.enableCors(&w, r)

	// get the body of our POST request
	// unmarshal this into a new Article struct
	// append this to our Articles array.
	reqBody, _ := ioutil.ReadAll(r.Body)
	var order Order
	json.Unmarshal(reqBody, &order)

	// update our global Pets array to include
	// our new Pet
	Orders = append(Orders, order)
	json.NewEncoder(w).Encode(order)

	w.Header().Set("Content-Type", "Application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}
