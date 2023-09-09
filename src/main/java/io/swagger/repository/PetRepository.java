package io.swagger.repository;

import io.swagger.model.Pet;
import org.springframework.data.mongodb.repository.MongoRepository;

import java.util.List;


public  interface PetRepository extends MongoRepository<Pet, String> {

    public Pet findByFirstName(String firstName);
    public List<Pet> findByLastName(String lastName);

}