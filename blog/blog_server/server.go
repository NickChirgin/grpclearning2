package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/nickchirgin/grpclearning2/blog/blogpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var collection *mongo.Collection

type server struct {
	blogpb.UnimplementedBlogServiceServer
}

type item struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string `bson:"author_id"`
	Content string `bson:"content"`
	Title string `bson:"title"`
}
func main(){
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	fmt.Println("Blog service server started")	
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://root:qwerty@172.17.0.2:27017"))
	if err != nil { log.Fatalf("Error while connecting to db: %v" ,err) }
	lis, err := net.Listen("tcp", "localhost:50051")
	collection = client.Database("grpc_test").Collection("blog")
	if err != nil {
		log.Fatalf("failed to listen %v", err)
	}
	s := grpc.NewServer()
	blogpb.RegisterBlogServiceServer(s, &server{})
	go func() {
		fmt.Println("Starting server")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch
	fmt.Println("Stopping the server")
	s.Stop()
	lis.Close()
	fmt.Println("End of program")
	client.Disconnect(context.TODO())
}

func (s *server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	fmt.Println("Creating blog")
	blog := req.GetBlog()
	data := item {
		AuthorID: blog.GetAuthorId(),
		Title: blog.GetTitle(),
		Content: blog.GetContent(),
	}
	res, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		) 
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		) 
	}
	return &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id: oid.Hex(),
			AuthorId: blog.GetAuthorId(),
			Title: blog.GetTitle(),
			Content: blog.GetContent(),
		},
	}, nil
}

func (*server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	fmt.Println("Read blog request")
	blogId := req.GetBlogId()
	oid, err := primitive.ObjectIDFromHex(blogId)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprint("Cannot parse ID"),
		)
	}
	data := &item{}
	filter := bson.D{{Key: "_id", Value: oid}}	
	result := collection.FindOne(context.Background(), filter)
	if err := result.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog with specified ID: %v", err),
		)
	}
	return &blogpb.ReadBlogResponse{
		Blog: &blogpb.Blog{
			AuthorId: data.AuthorID,
			Id: data.ID.Hex(),
			Content: data.Content,
			Title: data.Title,
		},
	}, nil	
}