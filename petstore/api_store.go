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
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
)

var Orders []Order

func DeleteOrder(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	fmt.Printf("DeleteOrder::id is %s\n", vars["orderId"])
	id, err := strconv.ParseInt(vars["orderId"], 10, 32)
	if err != nil {
		panic(err)
	}

	for index, order := range Orders {
		if order.Id == id {
			Orders = append(Orders[:index], Orders[index+1:]...)
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

}

func GetInventory(w http.ResponseWriter, r *http.Request) {

	fmt.Println("GetInventory:: return all orders")
	json.NewEncoder(w).Encode(Orders)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

}

func GetOrderById(w http.ResponseWriter, r *http.Request) {
	var result Order
	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["orderId"], 10, 32)
	if err != nil {

		panic(err)
	}
	fmt.Printf("GetOrderById id: %d\n", id)

	for _, order := range Orders {
		if order.Id == id {
			result = order

		}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if reflect.ValueOf(result).IsZero() {
		w.WriteHeader(http.StatusNotFound)
	} else {
		json.NewEncoder(w).Encode(result)

	}
}

func PlaceOrder(w http.ResponseWriter, r *http.Request) {

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

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}
