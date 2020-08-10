package main

import (
	"context"
	"fmt"
	calculatorpb "george_custom/grpc_course/calculator/calpb"
	"io"
	"log"
	"math"
	"net"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

// using struct is more convenient since the Calculate method use pointer receiver
// can directly pass &server{}, if using other underlying type(string, int), will need to assign a variable first
type server string

// Calculate function name is the same as the defined one in rpc proto file
// check CalculateServiceServer interface to see what methods to implement
func (s *server) Calculate(ctx context.Context, req *calculatorpb.CalculateRequest) (*calculatorpb.CalculateResponse, error) {
	fmt.Printf("Calculate function was invoked with %v\n", req)
	numOne := req.GetCalculator().GetNumOne()
	numTwo := req.GetCalculator().GetNumTwo()

	result := numOne + numTwo

	response := &calculatorpb.CalculateResponse{
		Result: result,
	}
	return response, nil

}

func (s *server) PrimeNumberDecomposition(req *calculatorpb.PrimeNumberRequest, stream calculatorpb.CalculateService_PrimeNumberDecompositionServer) error {
	fmt.Printf("PrimeNumberDecomposition function was invoked with %v\n", req)
	num := req.GetNumber()
	divisor := int64(2)

	for num > 1 {

		if num%divisor == 0 {
			res := &calculatorpb.PrimeNumberResponse{
				Result: divisor,
			}
			stream.Send(res)
			num = num / divisor
		} else {
			divisor++
		}
	}
	return nil

}

func (s *server) ComputeAverage(stream calculatorpb.CalculateService_ComputeAverageServer) error {
	fmt.Printf("ComputeAverage function was invoked with a streaming request...")
	sum := 0.0
	count := 0

	// stream has two functions
	// SendAndClose(*ComputeAverageResponse) error
	// Recv() (*ComputeAverageRequest, error)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// we finished reading the client stream
			res := &calculatorpb.ComputeAverageResponse{
				Answer: sum / float64(count),
			}
			return stream.SendAndClose(res)
		}
		sum += float64(req.GetNumber())
		count += 1.0

	}

}

func (s *server) FindMaximum(stream calculatorpb.CalculateService_FindMaximumServer) error {
	fmt.Printf("FindMaximum function was invoked with a streaming request...\n")
	max := math.Inf(-1)
	current := int64(0)

	// Send(*FindMaximumResponse) error
	// Recv() (*FindMaximumRequest, error)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// reached the end of receiving
			return nil
		}
		if err != nil {
			log.Fatalf("Something wrong with receiving from client: %v", err)
			return err
		}
		current = req.GetNumber()
		if float64(current) > max {
			max = float64(current)
			res := &calculatorpb.FindMaximumResponse{
				Result: int64(max),
			}
			// only when there's a maximum that the server sends back the response, otherwise do nothing
			err = stream.Send(res)
			if err != nil {
				log.Fatalf("error while sending data to client: %v", err)
				return err
			}
		}
	}

}

func (s *server) SquareRoot(ctx context.Context, req *calculatorpb.SquareRequest) (*calculatorpb.SquareResponse, error) {
	fmt.Println("Received SquareRoot RPC")
	num := req.GetNumber()
	if num < 0 {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Received a negative number: %v", num),
		)
	}

	res := &calculatorpb.SquareResponse{
		NumberRoot: math.Sqrt(float64(num)),
	}

	return res, nil

}

func main() {
	fmt.Println("Hello World! I will start serving.....")

	lis, err := net.Listen("tcp", "0.0.0.0:50050")
	if err != nil {
		log.Fatalf("Failed to listen; %v", err)
	}

	// create a new server object
	s := grpc.NewServer()
	s1 := server("calculator service")

	calculatorpb.RegisterCalculateServiceServer(s, &s1)

	// Register reflection service on gRPC server.
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
