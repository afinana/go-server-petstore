## Build
FROM golang:1.25 AS build

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY petstore ./petstore
COPY main.go .

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-server-petstore

## Deploy
##FROM gcr.io/distroless/base-debian11
FROM scratch

# Set the Current Working Directory inside the container
WORKDIR /

COPY --from=build /go-server-petstore /go-server-petstore

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
ENTRYPOINT ["/go-server-petstore"]