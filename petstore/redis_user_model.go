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
func (m *UserModel) FindAll() ([]User, error) {
	// Define variables
	ctx := context.Background()
	users := []User{}

	// Get all keys starting with user:* from redis
	iter := m.C.Scan(ctx, 0, "user:*", 0).Iterator()

	for iter.Next(ctx) {
		// store iter.Val() in id
		id := iter.Val()
		// show log of id
		fmt.Println(id)

		// Find user by id
		user, err := m.FindByID(id)
		if err != nil {
			return nil, err
		}
		users = append(users, *user)
	} // for

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// FindByID will be used to find a user registry by id
func (m *UserModel) FindByID(id string) (*User, error) {
	ctx := context.Background()

	// Find user by id
	data, err := m.C.Get(ctx, id).Result()
	if err != nil {
		// Checks if the user was not found
		return nil, err
	}
	var user User
	err = json.Unmarshal([]byte(data), &user)
	if err != nil {
		return nil, err
	}
	return &user, nil

}

// FindByID will be used to find a user registry by id
func (m *UserModel) FindByName(name string) (*User, error) {
	ctx := context.Background()

	// Find user in hset by name
	id, err := m.C.HGet(ctx, "users_name", name).Result()
	if err != nil {
		// Checks if the user was not found
		return nil, err
	}

	// Find user by id
	data, err := m.C.Get(ctx, id).Result()
	if err != nil {
		// Checks if the user was not found
		return nil, err
	}

	var user User
	err = json.Unmarshal([]byte(data), &user)
	if err != nil {
		return nil, err
	}
	return &user, nil

}

// Insert will be used to insert a new user registry
func (m *UserModel) Insert(user User) (*User, error) {

	// Add pet
	json, err := json.Marshal(user)
	if err != nil {
		// Checks if the pet was not found
		return nil, err
	}

	ctx := context.Background()
	// Add user with id
	user_key := fmt.Sprintf("user:%v", user.Id)
	err = m.C.Set(ctx, user_key, json, 0).Err()
	if err != nil {
		// Checks if the pet was not found
		return nil, err
	}

	// Add users name to hset user_names with id
	_, err = m.C.HSet(ctx, "user_names", user.Username, user_key).Result()
	if err != nil {
		// Checks if the hset was not found
		return nil, err
	}

	return &user, nil
}

// Delete will be used to delete a user registry
func (m *UserModel) Delete(id string) error {

	ctx := context.Background()

	// Find user by id
	data, err := m.C.Get(ctx, id).Result()
	if err != nil {
		// Checks if the user was not found
		return err
	}
	var user User
	err = json.Unmarshal([]byte(data), &user)
	if err != nil {
		return err
	}

	// Delete user by id
	err = m.C.Del(ctx, fmt.Sprintf("user:%v", id)).Err()
	if err != nil {
		// Checks if the user was not found
		return err
	}

	// Delete username of index
	err = m.C.HDel(ctx, "user_names", user.Username).Err()
	if err != nil {
		// Checks if the user was not found
		return err
	}
	return nil

}
