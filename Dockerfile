FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./cmd/server
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /go/bin/goose .
COPY --from=builder /app/sql/schema ./sql/schema
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh
CMD ["./entrypoint.sh"]