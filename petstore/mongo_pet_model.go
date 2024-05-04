package petstore

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// PetModel represent a mgo database session with a pet data model
type PetModel struct {
	C *mongo.Collection
}

// All method will be used to get all records from pets table
func (m *PetModel) All() ([]Pet, error) {
	// Define variables
	ctx := context.TODO()
	b := []Pet{}

	// Find all pets
	petCursor, err := m.C.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	err = petCursor.All(ctx, &b)
	if err != nil {
		return nil, err
	}

	return b, err
}

// FindByID will be used to find a pet registry by id
func (m *PetModel) FindByID(id string) (*Pet, error) {
	p, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	// Find pet by id
	var pet = Pet{}
	err = m.C.FindOne(context.TODO(), bson.M{"_id": p}).Decode(&pet)
	if err != nil {
		// Checks if the pet was not found
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("ErrNoDocuments")
		}
		return nil, err
	}

	return &pet, nil
}

// Insert will be used to insert a new pet registry
func (m *PetModel) Insert(pet Pet) (*mongo.InsertOneResult, error) {
	return m.C.InsertOne(context.TODO(), pet)
}

// Insert will be used to insert a new pet registry
func (m *PetModel) Update(pet Pet) (*mongo.InsertOneResult, error) {
	return m.C.InsertOne(context.TODO(), pet)
}

// Delete will be used to delete a pet registry
func (m *PetModel) Delete(id string) (*mongo.DeleteResult, error) {
	p, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return m.C.DeleteOne(context.TODO(), bson.M{"_id": p})
}

// FindByStatus will be used to find a pet registry by status
func (m *PetModel) FindByStatus(status []string) ([]Pet, error) {

	// db.getCollection('pets').find({
	//     $or: [
	//          {'status': 'tag01'} ,
	//          {'status': 'tag02'}
	//     ]
	// });
	var filters []bson.M
	for _, item := range status {
		filters = append(filters,
			bson.M{"status": item})
	}
	// filter is a single filter document that merges all filters
	filter := bson.M{"$or": filters}

	cursor, err := m.C.Find(context.TODO(), filter)
	if err != nil {
		panic(err)
	}
	// end find

	var pets []Pet
	if err = cursor.All(context.TODO(), &pets); err != nil {
		panic(err)
	}

	return pets, nil
}

// FindByStatus will be used to find a pet registry by status

func (m *PetModel) FindBytags(tags []string) ([]Pet, error) {

	// begin find
	// db.getCollection('pets').find({
	//     $or: [
	//          {'tags': {'$elemMatch': {'name': 'tag01'} } },
	//          {'tags': {'$elemMatch': {'name': 'tag02'} } }
	//     ]
	// });

	var filters []bson.M
	for _, tag := range tags {
		filters = append(filters,
			bson.M{"tags": bson.M{"$elemMatch": bson.M{"name": tag}}})
	}
	// filter is a single filter document that merges all filters
	filter := bson.M{"$or": filters}

	cursor, err := m.C.Find(context.TODO(), filter)
	if err != nil {
		panic(err)
	}
	// end find

	var pets []Pet
	if err = cursor.All(context.TODO(), &pets); err != nil {
		panic(err)
	}

	return pets, nil
}
