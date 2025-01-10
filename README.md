# Go API Server- REDIS/Mongo db version for swagger (1.0.1)


This is a sample server Petstore server.  You can find out more about Swagger at [http://swagger.io](http://swagger.io) or on [irc.freenode.net, #swagger](http://swagger.io/irc/).  For this sample, you can use the api key `special-key` to test the authorization filters.

## Overview
This server was generated by the [swagger-codegen]
(https://github.com/swagger-api/swagger-codegen) project.  
By using the [OpenAPI-Spec](https://github.com/OAI/OpenAPI-Specification) from a remote server, you can easily generate a server stub.  
-

To see how to make this your own, look here:

[README](https://github.com/swagger-api/swagger-codegen/blob/master/README.md)

- API version: 1.0.6
- Build date: 2022-11-27T23:27:07.696Z


### Running the server
To run the server, follow these simple steps:

```sh
go run main.go
```

Using credentials

```sh
go run main.go -enableCredentials=true -serverAddr=0.0.0.0 -serverPort=8080 -MONGODB_USERNAME=root MONGODB_PASSWORD=example -mongoDatabase=petstore -mongoURI=mongodb://mongo:27017
```

Using credentials in URL

```sh
go run main.go -serverAddr=0.0.0.0 -serverPort=8080 -mongoDatabase=petstore -mongoURI=mongodb://root:example@localhost:27017
```


## Running with Docker

``` sh
docker -t go-server-petstore build .
docker run --name go-petstore  -p 8090:8080  go-server-petstore
```

# Mongo DB queries examples

``` sh
db.getCollection('pets').find({"category.id": {$lt: 2 }})

db.getCollection('pets').find({"category.id": 1, status: "pending" })

db.getCollection('pets').find( {tags: { $elemMatch : { name : "tag01" }}})

```
