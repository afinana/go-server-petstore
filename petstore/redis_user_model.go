package petstore

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
)

// UserModel represent a mgo database session with a user data model
type UserModel struct {
	C *redis.Client
}

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

// Insert will be used to insert a new user registry
func (m *UserModel) Insert(user User) (*User, error) {
	// Add user
	data, err := json.Marshal(user)
	ctx := context.Background()

	// Find user by id
	err = m.C.Set(ctx, fmt.Sprintf("order:%v", user.Id), data, 0).Err()
	if err != nil {
		// Checks if the user was not found
		return nil, err
	}
	return &user, nil
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
