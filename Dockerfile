FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux go build -o /go-digisign ./cmd/app/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /go-digisign /app/go-digisign
COPY --from=builder /app/static ./static

EXPOSE 8080

CMD ["/app/go-digisign"]