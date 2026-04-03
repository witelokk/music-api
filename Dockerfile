FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o music-api ./cmd/api

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/music-api .
ENV HOST=0.0.0.0
ENV PORT=8080
EXPOSE 8080
CMD ["./music-api"]
