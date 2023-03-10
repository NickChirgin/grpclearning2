package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/nickchirgin/grpclearning2/blog/blogpb"
	"google.golang.org/grpc"
)

type server struct {
	blogpb.UnimplementedBlogServiceServer
}
func main(){
	fmt.Println("Blog service server started")	
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
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
}