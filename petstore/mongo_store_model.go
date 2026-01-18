package petstore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// StoreModel represent a mgo database session with a order data model
type StoreModel struct {
	C *mongo.Collection
}

// All method will be used to get all records from orders table
func (m *StoreModel) All(ctx context.Context) ([]Order, error) {
	var b []Order

	// Find all orders
	orderCursor, err := m.C.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	err = orderCursor.All(ctx, &b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// FindByID will be used to find a order registry by id
func (m *StoreModel) FindByID(ctx context.Context, id string) (*Order, error) {
	p, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	// Find order by id
	var order = Order{}
	err = m.C.FindOne(ctx, bson.M{"_id": p}).Decode(&order)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

// Insert will be used to insert a new order registry
func (m *StoreModel) Insert(ctx context.Context, order Order) (*mongo.InsertOneResult, error) {
	return m.C.InsertOne(ctx, order)
}

// Delete will be used to delete a order registry
func (m *StoreModel) Delete(ctx context.Context, id string) (*mongo.DeleteResult, error) {
	p, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return m.C.DeleteOne(ctx, bson.M{"_id": p})
}
