package main

import (
	"fmt"
	"net"

	"github.com/NikitaTitkov/gRPC-Server-CRUD/pkg/handlers"
	"github.com/NikitaTitkov/gRPC-Server-CRUD/pkg/users_v1"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	baseurl = "localhost:50051"
)

func main() {
	l, err := net.Listen("tcp", baseurl)
	if err != nil {
		logrus.Fatal(color.RedString("Listen error: "), err)

	}
	s := grpc.NewServer()
	reflection.Register(s)
	users_v1.RegisterUsersV1Server(s, &handlers.Server{})
	fmt.Println(color.GreenString("Server is running!"))

	if err := s.Serve(l); err != nil {
		logrus.Fatal(color.RedString("Serve error: "), err)
	}
}
