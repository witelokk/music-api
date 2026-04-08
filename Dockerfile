FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Install golang-migrate CLI with Postgres driver enabled
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.0

RUN CGO_ENABLED=0 GOOS=linux go build -o music-api ./cmd/api

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/music-api .
COPY --from=builder /app/openapi.yml .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

ENV HOST=0.0.0.0
ENV PORT=8080

EXPOSE 8080

CMD ["sh", "-c", "migrate -path ./migrations -database \"$DATABASE_URL\" up && ./music-api"]
