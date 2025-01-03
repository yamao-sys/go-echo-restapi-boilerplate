# syntax=docker/dockerfile:1

FROM golang:1.23.1-alpine3.20

WORKDIR /app

COPY . .

RUN go mod tidy && \
	go install honnef.co/go/tools/cmd/staticcheck@latest && \
	go install github.com/rubenv/sql-migrate/...@latest && \
	go install github.com/volatiletech/sqlboiler/v4@latest && \
	go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-mysql@latest && \
	go install github.com/joho/godotenv/cmd/godotenv@latest && \
	go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest

RUN apk add --update make

CMD sh -c "godotenv -f .env go run main.go"
