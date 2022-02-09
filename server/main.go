package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"

	"blogcrud/blogpb"
	blogservice "blogcrud/server/blog_service"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// if we crash the go code, we get the file name and the line number
	logger := log.Default()
	logger.SetFlags(log.LstdFlags | log.Lshortfile)

	logger.Println("process started...")

	// connect to mongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		logger.Fatalf("error => failed to connect to mongodb: %v", err)
	}
	defer client.Disconnect(context.Background())
	if err := client.Ping(context.Background(), nil); err != nil {
		logger.Fatalf("error => mongodb ping failed: %v", err)
	}
	logger.Println("connected to mongodb...")

	// access the collection
	blogCollection := client.Database("mydb").Collection("blog")

	// bind to port 50051
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		logger.Fatalf("error => failed to bind to port 50051: %v", err)
	}
	defer lis.Close()
	logger.Println("listener bound to port 50051...")

	opts := []grpc.ServerOption{}

	// create a new gRPC server
	s := grpc.NewServer(opts...)
	defer s.Stop()
	logger.Println("gRPC server instantiated...")

	// register blog service
	blogServiceServer := blogservice.NewBlogServiceServer(logger, blogCollection)
	blogpb.RegisterBlogServiceServer(s, blogServiceServer)
	logger.Println("blog service registered...")

	// enable service reflection
	reflection.Register(s)

	// server in a goroutine
	go func() {
		logger.Println("gRPC server listening for connections...")
		if err := s.Serve(lis); err != nil {
			logger.Fatalf("error => failed to serve: %v", err)
		}
	}()

	// create a buffered channel to listen for interrupt/kill
	sigChan := make(chan os.Signal, 128)

	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, os.Kill)

	sig := <-sigChan

	logger.Printf("received %s signal, shutting down gracefully", sig.String())
}
