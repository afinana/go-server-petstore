module middleland.net/swaggerapi

go 1.18

require middleland.net/petstore v0.0.0-00010101000000-000000000000

require (
	github.com/gorilla/mux v1.8.0 // indirect
	go.mongodb.org/mongo-driver v1.11.0 // indirect
)

replace middleland.net/petstore => ./petstore
