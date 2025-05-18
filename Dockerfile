FROM golang:1.24.3 AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 go build -o accessbot ./cmd/server

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/accessbot .

EXPOSE 8080

CMD ["./accessbot"]
