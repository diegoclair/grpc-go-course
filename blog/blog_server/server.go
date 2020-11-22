package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/diegoclair/grpc-go-course/blog/blogpb"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

type server struct {
	blogpb.UnimplementedBlogServiceServer
}

func main() {

	//if we crash the go code, we get the file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	blogpb.RegisterBlogServiceServer(s, &server{})

	go func() {
		fmt.Println("Blog server listening on port: ", port)

		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for Control C to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// Block until a signal is received
	<-ch
	fmt.Println("\nStopping the server")
	s.Stop()
	fmt.Println("Closing the listener")
	lis.Close()
	fmt.Println("End of program")
}
