package petstore

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// PetModel represent a mgo database session with a pet data model
type PetModel struct {
	C *redis.Client
}

// All method will be used to get all records from pets table
func (m *PetModel) All() ([]Pet, error) {

	// Define variables
	ctx := context.Background()
	pets := []Pet{}

	// Find all pets
	iter := m.C.Scan(ctx, 0, "pet:*", 0).Iterator()

	for iter.Next(ctx) {
		// Find pet by id
		pet, _ := m.FindByID(iter.Val())
		pets = append(pets, *pet)
	} // for

	if err := iter.Err(); err != nil {
		return nil, err

	}

	return pets, nil
}

// FindByID will be used to find a pet registry by id
func (m *PetModel) FindByID(id string) (*Pet, error) {

	return m.FindByRedisID(fmt.Sprintf("pet:%v", id))
}

// Insert will be used to insert a new pet registry
func (m *PetModel) Insert(pet Pet) (*Pet, error) {

	// Add pet
	json, err := json.Marshal(pet)
	if err != nil {
		// Checks if the pet was not found
		return nil, err
	}

	ctx := context.Background()
	// Add pet with id
	err = m.C.Set(ctx, fmt.Sprintf("pet:%v", pet.ID), json, 0).Err()
	if err != nil {
		// Checks if the pet was not found
		return nil, err
	}

	// Add status to hset with id
	status_tag := fmt.Sprintf("pet_status:%v", pet.Status)
	pet_key := fmt.Sprintf("pet:%v", pet.ID)

	_, err = m.C.HSet(ctx, status_tag, pet_key, pet.Status).Result()
	if err != nil {
		// Checks if the hset was not found
		return nil, err
	}

	// Add tags to hset with id
	tags := pet.Tags
	for _, tag := range tags {
		_, err = m.C.HSet(ctx, fmt.Sprintf("pet_tags:%v", tag.Name), pet_key, tag.Name).Result()
		if err != nil {
			// Checks if the hset was not found
			return nil, err
		}
	}

	return &pet, nil
}

// Insert will be used to insert a new pet registry
func (m *PetModel) Update(pet Pet) (*Pet, error) {

	log.Printf("Update::FindByID of id:%d \n", pet.ID)

	// Clean pet register
	m.DeleteByRedisID(fmt.Sprintf("%v", pet.ID))

	return m.Insert(pet)
}

// Delete will be used to delete a pet registry
func (m *PetModel) Delete(id string) error {

	ctx := context.Background()
	// Delete pet by id
	err := m.C.Del(ctx, fmt.Sprintf("pet:%v", id)).Err()

	return err

}

// FindByStatus will be used to find a pet registry by status
func (m *PetModel) FindByStatus(status []string) ([]Pet, error) {

	return m.FindByTagsRedis("pet_status:", status)
}

// FindByTags will be used to find a pet registry by tag
func (m *PetModel) FindBytags(tags []string) ([]Pet, error) {
	return m.FindByTagsRedis("pet_tags:", tags)
}

// FindByTagsRedis will be used to find a pet registry by a list of statuses or tags
func (m *PetModel) FindByTagsRedis(prefix string, tags []string) ([]Pet, error) {

	// begin find
	ctx := context.Background()
	// create a list of pets empty
	pets := []Pet{}

	for _, tag := range tags {

		key := fmt.Sprintf("%v%v", prefix, tag)

		log.Printf("FindByTagsRedis::HGet of keys=%s \n", key)
		// Get all ids of the given tag
		ids, err := m.C.HKeys(ctx, key).Result()
		if err != nil {
			// Checks if the pet was not found
			return nil, err
		}
		for _, id := range ids {
			log.Printf("FindByTagsRedis::FindByID of id=%s \n", id)
			pet, err := m.FindByRedisID(id)
			if err != nil {
				// Checks if the pet was not found
				break
			}
			pets = append(pets, *pet)

		}

	}
	return pets, nil
}

// FindByID will be used to find a pet registry by id
func (m *PetModel) FindByRedisID(id string) (*Pet, error) {

	ctx := context.Background()

	// Find pet by id
	data, err := m.C.Get(ctx, id).Result()
	if err != nil {
		// Checks if the pet was not found
		return nil, err
	}

	pet := Pet{}
	err = json.Unmarshal([]byte(data), &pet)
	if err != nil {
		panic(err)
	}
	return &pet, nil
}

// FindByID will be used to find a pet registry by id
func (m *PetModel) DeleteByRedisID(id string) error {

	ctx := context.Background()

	pet, err := m.FindByRedisID(id)
	if err == nil {

		// Clean old pet registry
		status_tag := fmt.Sprintf("pet_status:%v", pet.Status)
		pet_key := fmt.Sprintf("pet:%v", pet.ID)
		_ = m.C.Del(ctx, pet_key)
		_ = m.C.HDel(ctx, status_tag, pet_key)
		for _, tag := range pet.Tags {
			tag_key := fmt.Sprintf("%v%v", "pet_tags:", tag)
			_ = m.C.HDel(ctx, tag_key, pet_key)

		}

	}
	return err
}
