package blogservice

import (
	"blogcrud/blogpb"
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BlogServiceServer struct {
	l    *log.Logger
	coll *mongo.Collection
}

type BlogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Content  string             `bson:"content"`
	Title    string             `bson:"title"`
}

// Create and returns a pointer to a BlogServiceServer struct
func NewBlogServiceServer(l *log.Logger, coll *mongo.Collection) *BlogServiceServer {
	return &BlogServiceServer{l, coll}
}

func (b *BlogServiceServer) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	b.l.Println("received a CreateBlog rpc request")
	blog := req.GetBlog()
	data := BlogItem{
		AuthorID: blog.GetAuthorId(),
		Title:    blog.GetTitle(),
		Content:  blog.GetContent(),
	}

	// insert a document in mongodb
	res, err := b.coll.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error: %v", err)
	}

	// convert the ID (type interface{}) into ObjectID
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal, "cannot convert to objectID: %v", err)
	}

	// return the successful response
	return &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       oid.Hex(),
			AuthorId: blog.GetAuthorId(),
			Title:    blog.GetTitle(),
			Content:  blog.GetContent(),
		},
	}, nil
}
