package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/diegoclair/grpc-go-course/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

const (
	addressHost = "localhost:50051"
)

func main() {

	fmt.Println("Hello I'm a client")

	// https://grpc.io/docs/guides/auth/ -> here we can see the docs explaining how to do insecure connection and with TLS/SSL
	tls := true
	opts := grpc.WithInsecure()

	if tls {
		certFile := "ssl/ca.crt" //Certificate Authority Trust certificate
		creds, sslErr := credentials.NewClientTLSFromFile(certFile, "")
		if sslErr != nil {
			log.Fatalf("Error while loading CA trust certificate: %v", sslErr)
		}
		//if you got an error because of certificate, you can run:  export GODEBUG=x509ignoreCN=0

		opts = grpc.WithTransportCredentials(creds)
	}

	cc, err := grpc.Dial(addressHost, opts)
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer cc.Close()

	c := greetpb.NewGreetServiceClient(cc)
	// fmt.Printf("Created client: %f", c)

	doUnaryRequest(c)
	//doServerStreamingRequest(c)
	//doClientStreamingRequest(c)
	//doBiDiStreamingRequest(c)

	//doUnaryRequestWithDeadline(c, 5*time.Second) //should complete
	//doUnaryRequestWithDeadline(c, 1*time.Second) //should timeout, because the server will response after 3 seconds
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
		err := stream.Send(req[i])
		if err != nil {
			log.Fatalf("Error while sending streaming data to server: %v", err)
		}
		time.Sleep(1 * time.Second)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while receiving response for LongGreet: %v", err)
	}
	log.Printf("LongGreet response: %v\n", res)
}

func doBiDiStreamingRequest(c greetpb.GreetServiceClient) {

	fmt.Println("Starting to do a Bi Directional Streaming RPC...")

	req := []*greetpb.GreetEveryoneRequest{
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

	// we create a stream by invoking the client
	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Error while calling GreetEveryone RPC: %v", err)
	}

	//we don't need to use go routine, but in this case is good to see the request and receiving doing at the same time (parallel)

	// we send a bunch of messages to the client (go routine)
	waitChannel := make(chan struct{})
	go func() {
		for i := range req {
			fmt.Printf("Sending message: %v\n", req[i])
			stream.Send(req[i])
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
			fmt.Printf("Received: %v\n", res.GetResult())
		}
		close(waitChannel)
	}()

	// block until everything is done
	<-waitChannel
}

func doUnaryRequestWithDeadline(c greetpb.GreetServiceClient, timeout time.Duration) {

	fmt.Println("Starting to do a Unary with deadline RPC...")

	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Diego Clair",
			LastName:  "Rodrigues",
		},
	}

	//the documentation recommend to do the requests with a timeout defined
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := c.GreetWithDeadline(ctx, req)
	if err != nil {

		statusErr, ok := status.FromError(err)
		if ok { // its is a grpc error?
			if statusErr.Code() == codes.DeadlineExceeded {
				fmt.Println("The timeout was hit! Deadline")
			} else {
				fmt.Println("Unexpected error: ", statusErr)
			}
		} else {
			log.Fatalf("Error while calling GreetWithDeadline RPC: %v", err)
		}
		return
	}

	log.Printf("Response from GreetWithDeadline: %v", res.GetResult())
}
