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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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
	if _, ok := users.elements[req.User.GetID()]; ok {
		return nil, status.Errorf(codes.AlreadyExists, "user with ID %d already exists", req.User.GetID())
	}
	users.elements[req.User.GetID()] = req.User
	return &users_v1.CreateOut{ID: req.User.GetID()}, nil
}

// Get - returns the user associated with the specified ID
func (s *Server) Get(_ context.Context, req *users_v1.GetIn) (*users_v1.GetOut, error) {
	users.mutex.RLock()
	defer users.mutex.RUnlock()
	user := users.elements[req.ID]

	if user == nil {
		return nil, status.Errorf(codes.NotFound, "user with ID = %d not found", req.GetID())
	}
	return &users_v1.GetOut{User: user}, nil
}

// GetAll - method to get all users with limit and offset.
func (s *Server) GetAll(_ context.Context, req *users_v1.GetAllIn) (*users_v1.GetAllOut, error) {
	userSlice := make([]*users_v1.User, 0, len(users.elements))
	position := int64(0)
	users.mutex.RLock()
	defer users.mutex.RUnlock()
	for _, user := range users.elements {
		if position < req.GetOffset() {
			position++
			continue
		}
		userSlice = append(userSlice, user)
		if len(userSlice) == int(req.GetLimit()) {
			break
		}
	}
	return &users_v1.GetAllOut{Users: userSlice}, nil
}

// Delete - method that deletes a user from the in-memory users database.
func (s *Server) Delete(_ context.Context, req *users_v1.DeleteIn) (*emptypb.Empty, error) {
	users.mutex.Lock()
	defer users.mutex.Unlock()
	if _, ok := users.elements[req.GetID()]; !ok {
		return nil, status.Errorf(codes.NotFound, "user with ID = %d not found", req.GetID())
	}
	delete(users.elements, req.GetID())
	return &emptypb.Empty{}, nil
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
