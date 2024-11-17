package main

import (
	"context"
	"log"
	"time"

	"github.com/NikitaTitkov/gRPC-Server-CRUD/pkg/users_v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	addrurl = "localhost:50051"
)

func main() {
	con, err := grpc.NewClient(addrurl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connetct to %v : %v", addrurl, err)
	}
	defer func() {
		if err := con.Close(); err != nil {
			log.Printf("Failed to close connection: %v", err)
		}
	}()

	client := users_v1.NewUsersV1Client(con)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	rCreate, err := client.Create(ctx, &users_v1.CreateIn{
		User: &users_v1.CreateUser{
			Name:  "Nikita",
			Age:   20,
			Email: "nikita@gmail.com",
			Info: &users_v1.UserInfo{
				Street: "some street",
				City:   "some city",
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%v", rCreate)
}
