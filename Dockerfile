# syntax=docker/dockerfile:1

## Build
FROM golang:1.22 AS build

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY petstore ./petstore
COPY main.go .

RUN go mod download

RUN go build -o /swagger

## Deploy
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /swagger /swagger

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/swagger"]
