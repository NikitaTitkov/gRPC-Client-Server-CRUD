FROM golang:1.23-alpine AS builder

COPY . /github.com/NikitaTitkov/gRPC-Server-CRUD/source/
WORKDIR /github.com/NikitaTitkov/gRPC-Server-CRUD/source/

RUN go mod download
RUN go build -o ./bin/server cmd/server/main.go

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /github.com/NikitaTitkov/gRPC-Server-CRUD/source/bin/server .

CMD ["./server"]