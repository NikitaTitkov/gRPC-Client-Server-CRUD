package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/NikitaTitkov/gRPC-Server-CRUD/pkg/handlers"
	"github.com/NikitaTitkov/gRPC-Server-CRUD/pkg/users_v1"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/subosito/gotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	baseurl = "localhost:50051"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config-path", "local.env", "path to config file")
}

func main() {
	if err := initEnv(); err != nil {
		logrus.Fatalf("Failed to load environment variables from config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	opts := options.Client().ApplyURI(os.Getenv("DB_URI"))
	opts.SetAuth(options.Credential{
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
	})

	dbClient, err := mongo.Connect(ctx, opts)
	if err != nil {
		logrus.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := dbClient.Disconnect(ctx); err != nil {
			logrus.Errorf("Error disconnecting MongoDB client: %v", err)
		}
	}()

	if err := dbClient.Ping(ctx, nil); err != nil {
		logrus.Fatalf("MongoDB connection test failed: %v", err)
	}

	DB := dbClient.Database(os.Getenv("DB_NAME"))

	l, err := net.Listen("tcp", os.Getenv("BASEURL"))
	if err != nil {
		logrus.Fatalf("Listen error: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	users_v1.RegisterUsersV1Server(s, handlers.NewServ(DB))

	fmt.Println(color.GreenString("Server is running!"))

	if err := s.Serve(l); err != nil {
		logrus.Fatalf("Serve error: %v", err)
	}
}

func initEnv() error {
	if err := gotenv.Load(configPath); err != nil {
		return fmt.Errorf("error loading config file: %w", err)
	}
	return nil
}
