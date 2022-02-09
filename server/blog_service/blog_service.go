package blogservice

import (
	"blogcrud/blogpb"
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
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

func (b *BlogItem) toBlogPB() *blogpb.Blog {
	return &blogpb.Blog{
		Id:       b.ID.Hex(),
		AuthorId: b.AuthorID,
		Content:  b.Content,
		Title:    b.Title,
	}
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

func (b *BlogServiceServer) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	b.l.Println("received a ReadBlog rpc request")
	blogId := req.GetBlogId()

	// convert the id to an objectId
	oid, err := primitive.ObjectIDFromHex(blogId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot convert to object id: %v", err)
	}
	// create an empty BlogItem struct
	data := &BlogItem{}

	// filter for mongodb
	filter := bson.M{
		"_id": oid,
	}

	// find one result matching the filter
	res := b.coll.FindOne(context.Background(), filter)
	// store the result into data
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(codes.NotFound, "cannot find blog with specified id: %v", err)
	}

	// return the successful response
	return &blogpb.ReadBlogResponse{Blog: data.toBlogPB()}, nil
}

func (b *BlogServiceServer) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	b.l.Println("received an UpdateBlog rpc request")

	blog := req.GetBlog()
	// convert the id to an objectId
	oid, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot convert to object id: %v", err)
	}
	// create an empty BlogItem struct
	data := &BlogItem{}

	// filter for mongodb
	filter := bson.M{
		"_id": oid,
	}

	// find one result matching the filter
	res := b.coll.FindOne(context.Background(), filter)
	// store the result into data
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(codes.NotFound, "cannot find blog with specified id: %v", err)
	}

	// update the data struct
	data.AuthorID = blog.GetAuthorId()
	data.Content = blog.GetContent()
	data.Title = blog.GetTitle()

	// update one item
	_, err = b.coll.ReplaceOne(context.Background(), filter, data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot update object in mongodb: %v", err)
	}

	// return successful response
	return &blogpb.UpdateBlogResponse{Blog: data.toBlogPB()}, nil
}

func (b *BlogServiceServer) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	b.l.Println("received an UpdateBlog rpc request")

	// convert the id to an objectId
	oid, err := primitive.ObjectIDFromHex(req.GetBlogId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot convert to object id: %v", err)
	}

	// filter for mongodb
	filter := bson.M{
		"_id": oid,
	}

	// delete the blog
	res, err := b.coll.DeleteOne(context.Background(), filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot delete object in mongodb: %v", err)
	}

	// if nothing was deleted
	if res.DeletedCount == 0 {
		return nil, status.Errorf(codes.NotFound, "cannot find blog with specified id: %v", err)
	}

	// return successful response
	return &blogpb.DeleteBlogResponse{BlogId: req.GetBlogId()}, nil
}

func (b *BlogServiceServer) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {
	b.l.Println("received a ListBlog rpc request")

	// find all the blogs
	cursor, err := b.coll.Find(context.Background(), nil)
	if err != nil {
		return status.Errorf(codes.Internal, "unkown internal error: %v", err)
	}
	defer func() {
		err := cursor.Close(context.Background())
		if err != nil {
			b.l.Printf("error => failed to close cursor: %v", err)
		}
	}()

	for cursor.Next(context.Background()) {
		data := &BlogItem{}
		err := cursor.Decode(data)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to decode data from mongodb: %v", err)
		}
		err = stream.Send(&blogpb.ListBlogResponse{
			Blog: data.toBlogPB(),
		})
		if err != nil {
			return status.Errorf(codes.Internal, "failed to send data in response stream: %v", err)
		}
	}

	if err := cursor.Err(); err != nil {
		return status.Errorf(codes.Internal, "unkown internale error: %v", err)
	}

	return nil
}
