FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o camagru-api ./cmd/api/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/camagru-api .
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /app/entrypoint.sh .
COPY --from=builder /app/sql/schema ./sql/schema
COPY --from=builder /app/api ./api
COPY --from=builder /app/assets ./assets

RUN chmod +x ./entrypoint.sh

ENTRYPOINT [ "sh", "./entrypoint.sh" ]

CMD ["./camagru-api"]
