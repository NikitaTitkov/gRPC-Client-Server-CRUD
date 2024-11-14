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
		logrus.Fatal("failed to get env: ", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	opts := options.Client()

	opts.SetAuth(options.Credential{
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_password"),
	})

	opts.ApplyURI("localhost:27019")

	dbClient, err := mongo.Connect(ctx,opts)
	if err != nil{
		logrus.Fatal(err)
	}

	if err:=dbClient.Ping(context.Background(),nil); err != nil{
		
	}

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

func initEnv() error {
	if err := gotenv.Load(configPath); err != nil {
		return err
	}
	return nil
}
