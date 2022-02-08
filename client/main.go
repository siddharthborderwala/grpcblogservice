package main

import (
	"blogcrud/blogpb"
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	fmt.Println("blog client")

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	// dial up a connection
	cc, err := grpc.Dial("localhost:50051", opts...)
	if err != nil {
		log.Fatalf("error while connecting to gRPC service: %v", err)
	}
	defer cc.Close()

	// create a blog service client
	c := blogpb.NewBlogServiceClient(cc)

	in := &blogpb.CreateBlogRequest{
		Blog: &blogpb.Blog{
			AuthorId: "siddharth",
			Title:    "My First Blog",
			Content:  "Content of my first blog",
		},
	}

	// make a CreateBlog rpc request
	res, err := c.CreateBlog(context.Background(), in)
	if err != nil {
		log.Fatalf("failed to creat blog: %v", err)
	}
	fmt.Printf("created a blog %v", res.GetBlog())
}
