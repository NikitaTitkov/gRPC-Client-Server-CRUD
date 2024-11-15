package handlers

import (
	"context"
	"sync"

	"github.com/NikitaTitkov/gRPC-Server-CRUD/pkg/users_v1"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	userID := result.InsertedID.(primitive.ObjectID).Hex()

	return &users_v1.CreateOut{ID: userID}, nil
}

// Get - returns the user associated with the specified ID
func (s *Server) Get(ctx context.Context, req *users_v1.GetIn) (*users_v1.GetOut, error) {
	var user users_v1.User

	id, err := primitive.ObjectIDFromHex(req.GetID())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format: %v", err)
	}

	err = s.db.Collection("users").FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user with ID = %s not found", req.GetID())
	}
	return &users_v1.GetOut{User: &user}, nil
}

// GetAll - method to get all users with limit and offset.
func (s *Server) GetAll(ctx context.Context, req *users_v1.GetAllIn) (*users_v1.GetAllOut, error) {

	collection := s.db.Collection("users")
	filter := bson.M{}

	findOptions := options.Find().
		SetSkip(req.GetOffset()).
		SetLimit(req.GetLimit())

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		logrus.Printf("MongoDB query error: %v", err)
		return nil, status.Errorf(codes.Internal, "database query failed: %v", err)
	}
	defer cursor.Close(ctx)

	var users []*users_v1.User
	for cursor.Next(ctx) {
		var user users_v1.User
		if err := cursor.Decode(&user); err != nil {
			logrus.Printf("Error decoding user: %v", err)
			return nil, status.Errorf(codes.Internal, "failed to decode user: %v", err)
		}
		users = append(users, &user)
	}

	if err := cursor.Err(); err != nil {
		logrus.Printf("Cursor error: %v", err)
		return nil, status.Errorf(codes.Internal, "cursor error: %v", err)
	}

	return &users_v1.GetAllOut{Users: users}, nil
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
