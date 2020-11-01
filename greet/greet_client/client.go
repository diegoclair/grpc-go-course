package main

import (
	"context"
	"fmt"
	"log"

	"github.com/diegoclair/grpc-go-course/greet/greetpb"
	"google.golang.org/grpc"
)

const (
	addressHost = "localhost:50051"
)

func main() {

	fmt.Println("Hello I'm a client")
	cc, err := grpc.Dial(addressHost, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer cc.Close()

	c := greetpb.NewGreetServiceClient(cc)
	// fmt.Printf("Created client: %f", c)

	doUnaryRequest(c)
}

func doUnaryRequest(c greetpb.GreetServiceClient) {

	fmt.Println("Starting to do a Unary RPC...")

	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Diego Clair",
			LastName:  "Rodrigues",
		},
	}
	res, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling Greet RPC: %v", err)
	}
	log.Printf("Response from Greet: %v", res.Result)
}
