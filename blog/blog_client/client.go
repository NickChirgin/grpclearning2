package main

import (
	"context"
	"fmt"
	"log"

	"github.com/nickchirgin/grpclearning2/blog/blogpb"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("blog client")
	c, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	fmt.Println("Connected")
	defer c.Close()	
	bl := blogpb.NewBlogServiceClient(c)	
	/*
	blog :=	&blogpb.Blog{
		AuthorId: "Nick",
		Title: "First Blog",
		Content: "Content of the first blog",
	}
	res, err := bl.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{Blog: blog})
	if err != nil {
		log.Fatalf("Unexpected error: %v", err)
	}
	fmt.Printf("Blog has been created: %v", res.Blog.GetId())
	*/
	fmt.Println("Reading the blog")
	res, err2 := bl.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: "640b94dbcb7f2639bafba762"})
	if err2 != nil {
		fmt.Printf("Error happened while reading %v", err)
	}
	fmt.Printf("Blog was read: %v", res)
}