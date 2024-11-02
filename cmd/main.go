package main

import (
	"fmt"
	"log"
	"net"

	"github.com/NikitaTitkov/gRPC-Server-CRUD/pkg/users_v1"
	"github.com/fatih/color"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	baseurl = "localhost:50051"
)

// Server represents the unimplemented gRPC server that handles user-related requests.
type Server struct {
	users_v1.UnimplementedUsersV1Server
}

func main() {
	l, err := net.Listen("tcp", baseurl)
	if err != nil {
		log.Fatal(color.RedString("Listen	 error: "), err)

	}
	s := grpc.NewServer()
	reflection.Register(s)
	users_v1.RegisterUsersV1Server(s, &Server{})
	fmt.Println(color.GreenString("Server is running!"))

	if err := s.Serve(l); err != nil {
		log.Fatal(color.RedString("Serve error: "), err)
	}
}
