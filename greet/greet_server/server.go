package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/diegoclair/grpc-go-course/greet/greetpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

const (
	port = ":50051"
)

type server struct {
	greetpb.UnimplementedGreetServiceServer
}

func (s *server) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	fmt.Printf("Greet function was invoked with %v\n", req)

	firstName := req.GetGreeting().GetFirstName()

	result := "Hello " + firstName
	res := &greetpb.GreetResponse{
		Result: result,
	}
	return res, nil
}

func (s *server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {

	fmt.Printf("GreetManyTimes function was invoked with %v\n", req)

	firstName := req.GetGreeting().GetFirstName()

	res := &greetpb.GreetManyTimesResponse{}

	for i := 0; i < 10; i++ {
		result := "Hello " + firstName + " - number " + strconv.Itoa(i)
		res.Result = result
		err := stream.Send(res)
		if err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

func (s *server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {

	fmt.Printf("LongGreet function was invoked with a streaming request\n")

	var result string

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			//we've reached the end of the stream
			fmt.Println("Process finished!")
			return stream.SendAndClose(&greetpb.LongGreetResponse{
				Result: result,
			})

		}
		if err != nil {
			log.Fatalf("Error while reading client streaming: %v", err)
			return err
		}

		firstName := req.GetGreeting().GetFirstName()
		result += "Hello " + firstName + "! "
	}
}

func (s *server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {

	fmt.Printf("GreetEveryone function was invoked with a streaming request\n")

	for {
		firstName, err := s.getFirstNameFromRequest(stream)
		if err == io.EOF {
			return nil //we've reached the end of the stream
		}
		if err != nil {
			log.Fatalf("Error while reading client stream request: %v", err)
			return err
		}

		err = s.processResponse(stream, firstName)
		if err != nil {
			return err
		}
	}

}

func (s *server) getFirstNameFromRequest(stream greetpb.GreetService_GreetEveryoneServer) (result string, err error) {
	req, err := stream.Recv()
	if err != nil {
		return result, err
	}
	firstName := req.GetGreeting().GetFirstName()

	return firstName, nil
}

func (s *server) processResponse(stream greetpb.GreetService_GreetEveryoneServer, firstName string) (err error) {

	resBody := "Hello " + firstName + "! "

	err = stream.Send(&greetpb.GreetEveryoneResponse{
		Result: resBody,
	})
	if err != nil {
		log.Fatalf("Error while sending data to client stream: %v", err)
		return err
	}

	return nil
}

func (s *server) GreetWithDeadline(ctx context.Context, req *greetpb.GreetWithDeadlineRequest) (*greetpb.GreetWithDeadlineResponse, error) {

	fmt.Printf("Greet function was invoked with %v\n", req)

	firstName := req.GetGreeting().GetFirstName()

	result := "Hello " + firstName
	res := &greetpb.GreetWithDeadlineResponse{
		Result: result,
	}

	time.Sleep(3 * time.Second)
	if ctx.Err() == context.Canceled {
		fmt.Println("The client canceled the request!")
		return nil, status.Error(codes.Canceled, "The client canceled the request")
	}
	return res, nil
}

func main() {

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// https://grpc.io/docs/guides/auth/ -> here we can see the docs explaining how to do insecure connection and with TLS/SSL
	tls := true
	opts := []grpc.ServerOption{}

	if tls {
		certFile := "ssl/server.crt"
		keyFile := "ssl/server.pem"

		creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)
		if sslErr != nil {
			log.Fatalf("Failed loading certificates: %v", sslErr)
		}

		opts = append(opts, grpc.Creds(creds))
	}

	s := grpc.NewServer(opts...)
	greetpb.RegisterGreetServiceServer(s, &server{})

	fmt.Println("Greet server listening on port: ", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
