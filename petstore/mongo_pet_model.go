package petstore

import (
	"context"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// PetModel represent a mongo database session with a pet data model
type PetModel struct {
	C *mongo.Collection
}

// FindAll method will be used to get all records from pets table
func (m *PetModel) FindAll(ctx context.Context) ([]Pet, error) {
	var pets = []Pet{}

	// Find all pets
	petCursor, err := m.C.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	err = petCursor.All(ctx, &pets)
	if err != nil {
		return nil, err
	}

	return pets, nil
}

// FindByHexID will be used to find a pet registry by id
func (m *PetModel) FindByHexID(ctx context.Context, id string) (*Pet, error) {
	p, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	// Find pet by id
	var pet = Pet{}
	err = m.C.FindOne(ctx, bson.M{"_id": p}).Decode(&pet)
	if err != nil {
		return nil, err
	}

	return &pet, nil
}

// FindByID will be used to find a pet registry by id
func (m *PetModel) FindByID(ctx context.Context, id string) (*Pet, error) {

	// convert id to number
	index, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	// Find pet by id
	var pet = Pet{}
	err = m.C.FindOne(ctx, bson.M{"id": index}).Decode(&pet)
	if err != nil {
		return nil, err
	}

	return &pet, nil

}

// Insert will be used to insert a new pet registry
func (m *PetModel) Insert(ctx context.Context, pet Pet) (*mongo.InsertOneResult, error) {
	return m.C.InsertOne(ctx, pet)
}

// Update will be used to update an existing pet registry
func (m *PetModel) Update(ctx context.Context, pet Pet) (*mongo.UpdateResult, error) {
	return m.C.UpdateOne(ctx, bson.M{"_id": pet.ID}, bson.M{"$set": pet})
}

// Delete will be used to delete a pet registry
func (m *PetModel) Delete(ctx context.Context, id string) (*mongo.DeleteResult, error) {
	p, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		// If not hex, try to delete by numeric ID
		index, err := strconv.Atoi(id)
		if err != nil {
			return nil, err
		}
		return m.C.DeleteOne(ctx, bson.M{"id": index})
	}
	return m.C.DeleteOne(ctx, bson.M{"_id": p})
}

// FindByStatus will be used to find a pet registry by status
func (m *PetModel) FindByStatus(ctx context.Context, status []string) ([]Pet, error) {
	var filters []bson.M
	for _, item := range status {
		filters = append(filters,
			bson.M{"status": item})
	}
	// filter is a single filter document that merges all filters
	filter := bson.M{"$or": filters}

	cursor, err := m.C.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var pets = []Pet{}
	if err = cursor.All(ctx, &pets); err != nil {
		return nil, err
	}

	return pets, nil
}

// FindByTags will be used to find a pet registry by tags
func (m *PetModel) FindByTags(ctx context.Context, tags []string) ([]Pet, error) {

	var filters []bson.M
	for _, tag := range tags {
		filters = append(filters,
			bson.M{"tags": bson.M{"$elemMatch": bson.M{"name": tag}}})
	}
	// filter is a single filter document that merges all filters
	filter := bson.M{"$or": filters}

	cursor, err := m.C.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var pets = []Pet{}
	if err = cursor.All(ctx, &pets); err != nil {
		return nil, err
	}

	return pets, nil
}
