package petstore

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// OrderModel represent a mgo database session with a order data model
type OrderModel struct {
	C *redis.Client
}

// All method will be used to get all records from orders table
func (m *OrderModel) All() ([]string, error) {
	// Define variables
	ctx := context.Background()
	stores := []string{}
	var err error

	// Find all stores
	iter := m.C.Scan(ctx, 0, "stores:*", 0).Iterator()

	for iter.Next(ctx) {
		// Find store by id
		store, err := m.C.Get(ctx, iter.Val()).Result()
		if err != nil {
			// Checks if the store was not found
			return nil, err
		}
		stores = append(stores, store)
	} // for

	if err := iter.Err(); err != nil {
		return nil, err

	}

	return stores, err
}

// FindByID will be used to find a order registry by id
func (m *OrderModel) FindByID(id string) (string, error) {
	ctx := context.Background()

	// Find order by id
	order, err := m.C.Get(ctx, id).Result()
	if err != nil {
		// Checks if the order was not found
		return "", err
	}
	return order, nil
}

// Insert will be used to insert a new order registry
func (m *OrderModel) Insert(order Order) (*Order, error) {
	// Add order
	data, err := json.Marshal(order)
	ctx := context.Background()
	if err != nil {
		// Checks if the order was not found
		return nil, err
	}
	// Find order by id
	err = m.C.Set(ctx, fmt.Sprintf("order:%v", order.ID), data, 0).Err()
	if err != nil {
		// Checks if the order was not found
		return nil, err
	}
	return &order, nil
}

// Delete will be used to delete a order registry
func (m *OrderModel) Delete(id string) error {

	ctx := context.Background()
	// Delete pet by id
	err := m.C.Del(ctx, id).Err()

	return err
}
