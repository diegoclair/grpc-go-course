package main

import (
	"context"
	"fmt"
	"log"

	"github.com/diegoclair/grpc-go-course/calculator/calculatorpb"
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

	c := calculatorpb.NewCalculatorServiceClient(cc)
	// fmt.Printf("Created client: %f", c)

	doUnaryRequest(c)
}

func doUnaryRequest(c calculatorpb.CalculatorServiceClient) {

	fmt.Println("Starting to do a Unary RPC...")

	req := &calculatorpb.SumRequest{
		FirstNumber:  7,
		SecondNumber: 8,
	}
	res, err := c.Sum(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling Sum RPC: %v", err)
	}
	log.Printf("Response from Sum: %v", res.GetResult())
}
