# build stage
FROM golang:1.26-alpine AS builder
WORKDIR /src

ARG APP_CMD=./cmd/url-short

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/server ${APP_CMD}

# runtime stage
FROM alpine:3.21
WORKDIR /app
RUN adduser -D appuser
USER appuser

COPY --from=builder /app/server /app/server

EXPOSE 8080
CMD ["/app/server"]