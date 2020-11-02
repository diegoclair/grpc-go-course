package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

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

	//doUnaryRequest(c)
	//doServerStreamingRequest(c)
	doClientStreamingRequest(c)
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

func doServerStreamingRequest(c calculatorpb.CalculatorServiceClient) {

	fmt.Println("Starting to do a Server Streaming RPC...")

	req := &calculatorpb.PrimeNumberDecompositionRequest{
		Number: 21231654321654,
	}
	resStream, err := c.PrimeNumberDecomposition(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling PrimeNumberDecomposition RPC: %v", err)
	}

	for {
		res, err := resStream.Recv()
		if err == io.EOF {
			//we've reached the end of the stream
			fmt.Println("Process finished!")
			break
		}
		if err != nil {
			log.Fatalf("Error while reading the stream: %v", err)
		}
		log.Printf("Response from PrimeNumberDecomposition: %v", res.GetPrimeFactor())
	}
}

func doClientStreamingRequest(c calculatorpb.CalculatorServiceClient) {

	fmt.Println("Starting to do a ComputeAverage Client Streaming RPC...")

	stream, err := c.ComputeAverage(context.Background())
	if err != nil {
		log.Fatalf("Error while calling ComputeAverage RPC: %v", err)
	}

	numbers := []int64{3, 5, 9, 54, 23}
	for i := range numbers {
		fmt.Printf("Sending request number: %v\n", numbers[i])
		stream.Send(&calculatorpb.ComputeAverageRequest{
			Number: numbers[i],
		})
		time.Sleep(100 * time.Millisecond)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while reading ComputeAverage RPC: %v", err)
	}

	log.Printf("ComputeAverage response: %v", res.GetResult())
}
