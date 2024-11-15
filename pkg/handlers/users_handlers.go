package handlers

import (
	"context"
	"sync"

	"github.com/NikitaTitkov/gRPC-Server-CRUD/pkg/users_v1"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Handler struct {
	DB *mongo.Database
}

// Server represents the unimplemented gRPC server that handles user-related requests.
type syncMap struct {
	elements map[int64]*users_v1.User
	mutex    sync.RWMutex
}

var (
	users = &syncMap{elements: make(map[int64]*users_v1.User)}
)

// Server implements the UsersV1Server interface for handling gRPC requests.
type Server struct {
	users_v1.UnimplementedUsersV1Server
	db *mongo.Database
}

func NewServ(db *mongo.Database) *Server {
	return &Server{
		db: db,
	}
}

// Create - creates a new user.
func (s *Server) Create(ctx context.Context, req *users_v1.CreateIn) (*users_v1.CreateOut, error) {
	var existingUser bson.M
	user := req.GetUser()

	err := s.db.Collection("users").FindOne(ctx, bson.M{"email": user.GetEmail()}).Decode(&existingUser)
	if err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "user with email %s already exists", user.Email)
	}

	result, insertErr := s.db.Collection("users").InsertOne(ctx, user)
	if insertErr != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", insertErr)
	}

	userID := result.InsertedID.(primitive.ObjectID).Hex() // Преобразуем ObjectID в строку

	return &users_v1.CreateOut{ID: userID}, nil
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

// Update - method that updates a user in the in-memory users database.
func (s *Server) Update(_ context.Context, req *users_v1.UpdateIn) (*emptypb.Empty, error) {
	users.mutex.Lock()
	defer users.mutex.Unlock()
	if _, ok := users.elements[req.GetID()]; !ok {
		return nil, status.Errorf(codes.NotFound, "user with ID = %d not found", req.GetID())
	}
	if req.User.GetName() != nil {
		users.elements[req.GetID()].Name = req.User.GetName().GetValue()
	}
	if req.User.GetAge() != nil {
		users.elements[req.GetID()].Age = req.User.GetAge().GetValue()
	}
	if req.User.GetEmail() != nil {
		users.elements[req.GetID()].Email = req.User.GetEmail().GetValue()
	}
	if req.User.Info.GetStreet() != nil {
		users.elements[req.GetID()].Info.Street = req.User.Info.GetStreet().GetValue()
	}
	if req.User.Info.GetCity() != nil {
		users.elements[req.GetID()].Info.City = req.User.Info.GetCity().GetValue()
	}
	return &emptypb.Empty{}, nil
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
