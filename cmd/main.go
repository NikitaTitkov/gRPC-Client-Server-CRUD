package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/NikitaTitkov/gRPC-Server-CRUD/pkg/users_v1"
	"github.com/fatih/color"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	baseurl = "localhost:50051"
)

type syncMap struct {
	elements map[int64]*users_v1.User
	mutex    sync.RWMutex
}

var (
	users = &syncMap{elements: make(map[int64]*users_v1.User)}
)

// Server represents the unimplemented gRPC server that handles user-related requests.
type Server struct {
	users_v1.UnimplementedUsersV1Server
}

// Create - creates a new user.
func (s *Server) Create(_ context.Context, req *users_v1.CreateIn) (*users_v1.CreateOut, error) {
	users.mutex.Lock()
	defer users.mutex.Unlock()
	users.elements[req.User.GetID()] = req.User
	return &users_v1.CreateOut{ID: req.User.GetID()}, nil
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
