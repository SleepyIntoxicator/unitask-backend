FROM golang:1.17-alpine3.15

RUN apk update \
    && apk add --no-cache \
    postgresql

RUN mkdir /app
WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod verify
RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux go build  -o bin/APIServer/apiserver ./cmd/apiserver

RUN apk --no-cache add ca-certificates

ENTRYPOINT ["./bin/APIServer/apiserver"]