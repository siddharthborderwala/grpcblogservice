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

	// create a new blog and return its id
	blogId := CreateBlog(c)
	blog := ReadBlog(c, blogId)
	fmt.Println(blog)
}

func CreateBlog(c blogpb.BlogServiceClient) string {
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
		log.Fatalf("failed to create blog: %v", err)
		return ""
	}
	fmt.Printf("created a blog %v", res.GetBlog())
	return res.GetBlog().GetId()
}

func ReadBlog(c blogpb.BlogServiceClient, id string) *blogpb.Blog {
	in := &blogpb.ReadBlogRequest{
		BlogId: id,
	}
	res, err := c.ReadBlog(context.Background(), in)
	if err != nil {
		log.Fatalf("failed to read blog: %v", err)
		return nil
	}
	fmt.Printf("found a blog %v", res.GetBlog())
	return res.GetBlog()
}
