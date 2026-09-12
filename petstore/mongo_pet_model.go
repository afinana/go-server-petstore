package petstore

import (
	"context"
	"strconv"
	"strings"

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

// FindByID will be used to find a pet registry by numeric id
func (m *PetModel) FindByID(ctx context.Context, id string) (*Pet, error) {
	index, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	var pet = Pet{}
	err = m.C.FindOne(ctx, bson.M{"id": index}).Decode(&pet)
	if err != nil {
		return nil, err
	}

	return &pet, nil
}

// FindByIDOrHex resolves a pet either by 24-character hex ObjectID or numeric ID
func (m *PetModel) FindByIDOrHex(ctx context.Context, id string) (*Pet, error) {
	if p, err := primitive.ObjectIDFromHex(id); err == nil {
		var pet Pet
		err = m.C.FindOne(ctx, bson.M{"_id": p}).Decode(&pet)
		if err == nil {
			return &pet, nil
		}
		if err != mongo.ErrNoDocuments {
			return nil, err
		}
	}

	if index, err := strconv.ParseInt(id, 10, 64); err == nil {
		var pet Pet
		err = m.C.FindOne(ctx, bson.M{"id": index}).Decode(&pet)
		if err == nil {
			return &pet, nil
		}
		if err != mongo.ErrNoDocuments {
			return nil, err
		}
	}

	return nil, mongo.ErrNoDocuments
}

// CountByOwner counts how many pets are owned by a given user
func (m *PetModel) CountByOwner(ctx context.Context, ownerID string) (int64, error) {
	if ownerID == "" {
		return 0, nil
	}
	return m.C.CountDocuments(ctx, bson.M{"ownerId": ownerID})
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
		index, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil, err
		}
		return m.C.DeleteOne(ctx, bson.M{"id": index})
	}
	return m.C.DeleteOne(ctx, bson.M{"_id": p})
}

// FindByStatus safely queries pets matching a list of statuses
func (m *PetModel) FindByStatus(ctx context.Context, status []string) ([]Pet, error) {
	var validStatuses []string
	for _, item := range status {
		trimmed := strings.ToLower(strings.TrimSpace(item))
		if validPetStatuses[trimmed] {
			validStatuses = append(validStatuses, trimmed)
		}
	}

	if len(validStatuses) == 0 {
		return []Pet{}, nil
	}

	filter := bson.M{"status": bson.M{"$in": validStatuses}}
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

// FindByTags safely queries pets matching a list of tags
func (m *PetModel) FindByTags(ctx context.Context, tags []string) ([]Pet, error) {
	var validTags []string
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" && len(trimmed) <= MaxTagLength {
			validTags = append(validTags, trimmed)
		}
	}

	if len(validTags) == 0 {
		return []Pet{}, nil
	}

	filter := bson.M{"tags.name": bson.M{"$in": validTags}}
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
