package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/diegoclair/grpc-go-course/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	//doClientStreamingRequest(c)
	//doBiDiStreamingRequest(c)

	doErrorunary(c)
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

func doBiDiStreamingRequest(c calculatorpb.CalculatorServiceClient) {

	fmt.Println("Starting to do a Bi Directional Streaming RPC...")

	// we create a stream by invoking the client
	stream, err := c.FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("Error while opening stream and calling FindMaximum RPC: %v", err)
	}

	//we don't need to use go routine, but in this case is good to see the request and receiving doing at the same time (parallel)

	// we send a bunch of messages to the client (go routine)
	waitChannel := make(chan struct{})
	go func() {
		numbers := []int64{4, 7, 2, 19, 4, 6, 32}
		for i := range numbers {
			fmt.Printf("Sending message: %v\n", numbers[i])
			stream.Send(&calculatorpb.FindMaximumRequest{
				Number: numbers[i],
			})
			time.Sleep(1 * time.Second)
		}
		stream.CloseSend()
	}()

	// we receive a bunch of messages from the server (go routine)
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				//we've reached the end of the stream
				break
			}
			if err != nil {
				log.Fatalf("Error while receiving data from the server: %s", err)
				break
			}
			fmt.Printf("Received a new maximum of...: %v\n", res.GetMaximum())
		}
		close(waitChannel)
	}()

	// block until everything is done
	<-waitChannel
}

func doErrorunary(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting the doErrorunary function - correct call")
	//correct call
	doErrorCall(c, 10)

	fmt.Println("Starting the doErrorunary function - error call")
	//error call
	doErrorCall(c, -2)
}

func doErrorCall(c calculatorpb.CalculatorServiceClient, number int64) {
	res, err := c.SquareRoot(context.Background(), &calculatorpb.SquareRootRequest{Number: number})
	if err != nil {
		respErr, ok := status.FromError(err) //check if the statusError is a grpc statusError
		if ok {
			//actual error from gRPC (user error)
			fmt.Println("Error message from server: ", respErr.Message())
			fmt.Println("Status code from server: ", respErr.Code())
			if respErr.Code() == codes.InvalidArgument {
				fmt.Println("We probably sent a negative number!")
				return
			}
		} else {
			log.Fatalf("Error while calling SquareRoot: %v", err)
			return
		}
	}

	fmt.Printf("Result of square root of %v: %v\n", number, res.GetNumberRoot())
}
