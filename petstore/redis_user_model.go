package petstore

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/streadway/amqp"
)

// UserModel represent a mgo database session with a user data model
type UserModel struct {
	C *redis.Client
	Q *amqp.Connection
}

// REDIS METHODS
// All method will be used to get all records from users table
func (m *UserModel) All() ([]string, error) {
	// Define variables
	ctx := context.Background()
	users := []string{}

	// Find all users
	iter := m.C.Scan(ctx, 0, "users:*", 0).Iterator()

	for iter.Next(ctx) {
		// Find user by id
		user, err := m.C.Get(ctx, iter.Val()).Result()
		if err != nil {
			// Checks if the user was not found
			return nil, err
		}
		users = append(users, user)
	} // for

	if err := iter.Err(); err != nil {
		return nil, err

	}

	return users, nil
}

// FindByID will be used to find a user registry by id
func (m *UserModel) FindByID(id string) (string, error) {
	ctx := context.Background()

	// Find user by id
	user, err := m.C.Get(ctx, id).Result()
	if err != nil {
		// Checks if the user was not found
		return "", err
	}
	return user, nil
}

// Delete will be used to delete a user registry
func (m *UserModel) Delete(id string) error {

	ctx := context.Background()
	// Delete user by id
	err := m.C.Del(ctx, id).Err()
	if err != nil {
		// Checks if the user was not found
		return err
	}
	return nil

}

// RABBIT METHODS

// Insert will be used to insert a new user registry
func (m *UserModel) Insert(user User) (*User, error) {
	// Add user
	body, err := json.Marshal(user)
	if err != nil {
		// Checks if the user was not found
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
		"petstore",     // exchange
		"users-insert", // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		})
	if err != nil {
		// Checks if the user was not found
		return nil, err
	}

	return &user, nil
}

// Update will be used to update a user registry
func (m *UserModel) Update(user User) (*User, error) {
	// Clean user register
	m.Delete(strconv.FormatInt(user.Id, 10))
	// publish message to update user
	return m.Insert(user)
}
