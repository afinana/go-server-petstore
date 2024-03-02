package petstore

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/streadway/amqp"
)

// OrderModel represent a mgo database session with a order data model
type OrderModel struct {
	C *redis.Client
	Q *amqp.Connection
}

// REDIS METHODS
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

// Delete will be used to delete a order registry
func (m *OrderModel) Delete(id string) error {

	ctx := context.Background()
	// Delete pet by id
	err := m.C.Del(ctx, id).Err()

	return err
}

// RABBITMQ methods
// Insert will be used to insert a new order registry
func (m *OrderModel) Insert(order Order) (*Order, error) {
	// Add order
	body, err := json.Marshal(order)
	if err != nil {
		// Checks if the order was not found
		return nil, err
	}

	// create a new channel
	ch, err := m.Q.Channel()
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	// Publish a message to the queue
	err = ch.Publish(
		"petstore", // exchange
		"orders",   // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		})
	if err != nil {
		// Checks if the order was not found
		return nil, err
	}

	return &order, nil
}

// Update will be used to update a order registry
func (m *OrderModel) Update(order Order) (*Order, error) {
	// Clean order register
	m.Delete(strconv.FormatInt(order.Id, 10))
	// publish message to update order
	return m.Insert(order)
}
