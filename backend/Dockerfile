FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o camagru-api ./main.go 

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/camagru-api .

CMD ["./camagru-api"]