package petstore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserModel represent a mgo database session with a user data model
type UserModel struct {
	C *mongo.Collection
}

// All method will be used to get all records from users table
func (m *UserModel) All(ctx context.Context) ([]User, error) {
	var b []User

	// Find all users
	userCursor, err := m.C.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	err = userCursor.All(ctx, &b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// FindByID will be used to find a user registry by id
func (m *UserModel) FindByID(ctx context.Context, id string) (*User, error) {
	p, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	// Find user by id
	var user = User{}
	err = m.C.FindOne(ctx, bson.M{"_id": p}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Insert will be used to insert a new user registry
func (m *UserModel) Insert(ctx context.Context, user User) (*mongo.InsertOneResult, error) {
	return m.C.InsertOne(ctx, user)
}

// Delete will be used to delete a user registry
func (m *UserModel) Delete(ctx context.Context, id string) (*mongo.DeleteResult, error) {
	p, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return m.C.DeleteOne(ctx, bson.M{"_id": p})
}

// Update will be used to update a user registry
func (m *UserModel) Update(ctx context.Context, id string, user User) (*mongo.UpdateResult, error) {
	p, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return m.C.UpdateOne(ctx, bson.M{"_id": p}, bson.M{"$set": user})
}

// FindByUserName will be used to find a user registry by username
func (m *UserModel) FindByUserName(ctx context.Context, username string) (*User, error) {
	// Find user by username
	var user = User{}
	err := m.C.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
