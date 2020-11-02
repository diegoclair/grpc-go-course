package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

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

	//doUnaryRequest(c)
	//doServerStreamingRequest(c)
	doClientStreamingRequest(c)
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
	log.Printf("Response from Greet: %v", res.GetResult())
}

func doServerStreamingRequest(c greetpb.GreetServiceClient) {

	fmt.Println("Starting to do a Server Streaming RPC...")

	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Diego Clair",
			LastName:  "Rodrigues",
		},
	}
	resStream, err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling GreetManyTimes RPC: %v", err)
	}

	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			//we've reached the end of the stream
			fmt.Println("Process finished!")
			break
		}
		if err != nil {
			log.Fatalf("Error while reading the stream: %v", err)
		}
		log.Printf("Response from GreetManyTimes: %v", msg.GetResult())
	}
}

func doClientStreamingRequest(c greetpb.GreetServiceClient) {

	fmt.Println("Starting to do a Client Streaming RPC...")

	req := []*greetpb.LongGreetRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Diego Clair",
				LastName:  "Rodrigues",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Marcos",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Maria",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Pedro",
			},
		},
	}

	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("Error while calling LongGreet RPC: %v", err)
	}

	//we interate over our slice and send each message individually
	for i := range req {
		fmt.Printf("Send request: %v\n", req[i])
		stream.Send(req[i])
		time.Sleep(1 * time.Second)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while receiving response for LongGreet: %v", err)
	}
	log.Printf("LongGreet response: %v\n", res)
}
