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
	defer con.Close()

	client := users_v1.NewUsersV1Client(con)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := client.Create(ctx, &users_v1.CreateIn{
		User: &users_v1.User{
			ID:    1,
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
	log.Printf("%v", r)
}
