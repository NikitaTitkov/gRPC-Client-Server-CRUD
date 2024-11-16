package handlers

import (
	"context"

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

// Server implements the UsersV1Server interface for handling gRPC requests.
type Server struct {
	users_v1.UnimplementedUsersV1Server
	db *mongo.Database
}

// NewServ creates a new server instance with a connection to the database.
func NewServ(db *mongo.Database) *Server {
	return &Server{
		db: db,
	}
}

// Create - creates a new user.
func (s *Server) Create(ctx context.Context, req *users_v1.CreateIn) (*users_v1.CreateOut, error) {
	user := req.GetUser()

	err := s.db.Collection("users").FindOne(ctx, bson.M{"email": user.GetEmail()}).Err()
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

// Get - returns the = associated with the specified ID
func (s *Server) Get(ctx context.Context, req *users_v1.GetIn) (*users_v1.GetOut, error) {
	var user users_v1.User
	id, err := primitive.ObjectIDFromHex(req.GetID())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format: %v", err)
	}

	err = s.db.Collection("users").FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	user.ID = id.Hex()
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
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			logrus.Printf("failed to close cursor: %v", err)
		}
	}()

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
func (s *Server) Update(ctx context.Context, req *users_v1.UpdateIn) (*emptypb.Empty, error) {
	collection := s.db.Collection("users")

	filter := bson.M{"id": req.GetID()}
	err := collection.FindOne(ctx, filter).Err()
	if err == mongo.ErrNoDocuments {
		return nil, status.Errorf(codes.NotFound, "user with ID = %s not found", req.GetID())
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check user existence: %v", err)
	}

	update := bson.M{"$set": bson.M{}}
	if req.User.GetName() != nil {
		update["$set"].(bson.M)["name"] = req.User.GetName().GetValue()
	}
	if req.User.GetAge() != nil {
		update["$set"].(bson.M)["age"] = req.User.GetAge().GetValue()
	}
	if req.User.GetEmail() != nil {
		update["$set"].(bson.M)["email"] = req.User.GetEmail().GetValue()
	}
	if req.User.Info.GetStreet() != nil {
		update["$set"].(bson.M)["info.street"] = req.User.Info.GetStreet().GetValue()
	}
	if req.User.Info.GetCity() != nil {
		update["$set"].(bson.M)["info.city"] = req.User.Info.GetCity().GetValue()
	}
	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// Delete - method that deletes a user from the in-memory users database.
func (s *Server) Delete(ctx context.Context, req *users_v1.DeleteIn) (*emptypb.Empty, error) {
	collection := s.db.Collection("users")
	id, err := primitive.ObjectIDFromHex(req.GetID())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format: %v", err)
	}

	filter := bson.M{"_id": id}

	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	if result.DeletedCount == 0 {
		return nil, status.Errorf(codes.NotFound, "user with ID = %s not found", req.GetID())
	}

	return &emptypb.Empty{}, nil
}
