package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"

	"github.com/diegoclair/grpc-go-course/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	port = ":50051"
)

type server struct {
	calculatorpb.UnimplementedCalculatorServiceServer
}

func (s *server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	fmt.Printf("Sum function was invoked with %v\n", req)
	firstNumber := req.GetFirstNumber()
	secondNumber := req.GetSecondNumber()

	result := int64(firstNumber + secondNumber)
	res := &calculatorpb.SumResponse{
		Result: result,
	}

	return res, nil
}

func (s *server) PrimeNumberDecomposition(req *calculatorpb.PrimeNumberDecompositionRequest, stream calculatorpb.CalculatorService_PrimeNumberDecompositionServer) error {
	fmt.Printf("PrimeNumberDecomposition function was invoked with %v\n", req)
	number := req.GetNumber()

	divisor := 2
	n := int(number)

	res := &calculatorpb.PrimeNumberDecompositionResponse{}
	for n > 1 {
		if n%divisor == 0 {
			res.PrimeFactor = int64(divisor)
			err := stream.Send(res)
			if err != nil {
				return err
			}

			n = n / divisor
		} else {
			divisor++
		}
		fmt.Println("Divisor has increased to ", divisor)
	}

	return nil
}

func (s *server) ComputeAverage(stream calculatorpb.CalculatorService_ComputeAverageServer) error {

	fmt.Printf("Received ComputeAverage RPC\n")

	var sum int64
	var quantity int64
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			//we've reached the end of the stream
			result := float32(sum) / float32(quantity)
			return stream.SendAndClose(&calculatorpb.ComputeAverageResponse{
				Result: result,
			})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
			return err
		}

		sum += res.GetNumber()
		quantity++
	}
}

func (s *server) FindMaximum(stream calculatorpb.CalculatorService_FindMaximumServer) error {

	fmt.Printf("FindMaximum function was invoked\n")
	var maximumNumber int64

	for {
		number, err := s.getNumberFromRequest(stream)
		if err == io.EOF {
			return nil //we've reached the end of the stream
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
			return err
		}

		if number > maximumNumber {
			maximumNumber = number
			err = s.processResponse(stream, maximumNumber)
			if err != nil {
				return err
			}
		}
	}

}

func (s *server) getNumberFromRequest(stream calculatorpb.CalculatorService_FindMaximumServer) (number int64, err error) {

	req, err := stream.Recv()
	if err != nil {
		return number, err
	}

	number = req.GetNumber()

	return number, nil
}

func (s *server) processResponse(stream calculatorpb.CalculatorService_FindMaximumServer, maximum int64) (err error) {

	err = stream.Send(&calculatorpb.FindMaximumResponse{
		Maximum: maximum,
	})
	if err != nil {
		log.Fatalf("Error while sending data to client stream: %v", err)
		return err
	}

	return nil
}

func (s *server) SquareRoot(ctx context.Context, req *calculatorpb.SquareRootRequest) (numberRoot *calculatorpb.SquareRootResponse, err error) {

	//function that return a status error to client
	number := req.GetNumber()

	if number < 0 {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("The number cannot be negative: %v", number))
	}

	numberRoot = &calculatorpb.SquareRootResponse{
		NumberRoot: math.Sqrt(float64(number)),
	}

	return numberRoot, nil
}

func main() {

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Could not get the listener: %v", err)
	}

	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	fmt.Println("Server listening on port: ", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
