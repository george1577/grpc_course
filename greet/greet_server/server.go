package main

import (
	"context"
	"fmt"
	"george_custom/grpc_course/greet/greetpb"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

type server struct{}

var numRequest int

// GreetRequest and GreetResponse are both struct(refer to the generated code greet.pb.go)
// check GreetServiceServer to see what methods to implement
func (s *server) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	numRequest++
	fmt.Printf("Greet function was invoked with %v\n", req)
	fmt.Printf("Received %v request so far...\n", numRequest)
	firstName := req.GetGreeting().GetFirstName()
	result := "Hello " + firstName

	// return needs to be a pointer
	res := &greetpb.GreetResponse{Result: result}
	return res, nil

}

// stream object is an interface with Send() method
// GreetServiceServer interface
func (s *server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	fmt.Printf("GreetManyTimes function was invoked with %v\n", req)
	firstName := req.GetGreeting().GetFirstName()
	for i := 0; i < 10; i++ {
		res := &greetpb.GreetManyTimesResponse{
			Result: "Hello " + firstName + " for the " + strconv.Itoa(i+1) + " times",
		}
		stream.Send(res)
		time.Sleep(1000 * time.Millisecond)
	}
	return nil
}

// check GreetServiceServer to see what methods to implement
func (s *server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {
	fmt.Println("LongGreet function was invoked with a streaming request")
	result := "Hello "
	// stream has two functions
	// SendAndClose(*LongGreetResponse) error
	// Recv() (*LongGreetRequest, error)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			res := &greetpb.LongGreetResponse{
				Result: result,
			}
			// if finish reading the client stream
			// will send all the responses "at once"
			// since the function needs to return an error interface, which aligns with the SendAndClose() return type
			// thus can directly return it when reaching the end of client
			return stream.SendAndClose(res)

		}
		if err != nil {
			log.Fatalf("Error while reading client stream, %v", err)
		}

		firstName := req.GetGreeting().GetFirstName()
		result += firstName + "! "

	}

}

func (s *server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {
	fmt.Println("GreetEveryone function was invoked with a streaming request")

	// stream has both Recv() and Send() function, can do as many times on both side
	// Send(*GreetEveryoneResponse) error
	// Recv() (*GreetEveryoneRequest, error)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil

		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
			return err

		}
		firstName := req.GetGreeting().GetFirstName()
		result := "Hello " + firstName + "! "
		res := &greetpb.GreetEveryoneResponse{
			Result: result,
		}
		err = stream.Send(res)
		if err != nil {
			log.Fatalf("Error while sending data to client: %v", err)
			return err
		}

	}

}

func (s *server) GreetWithDeadline(ctx context.Context, req *greetpb.GreetWithDeadlineRequest) (*greetpb.GreetWithDeadlineResponse, error) {
	fmt.Printf("GreetWithDeadline function was invoked with %v\n", req)

	// let server sleep for three seconds
	for i := 0; i < 3; i++ {
		if ctx.Err() == context.Canceled {
			// the client canceled the request
			fmt.Println("The client cancelled the request!!")
			return nil, status.Error(codes.DeadlineExceeded, "the client cancelled the request")
		}
		time.Sleep(1 * time.Second)
	}

	firstName := req.GetGreeting().GetFirstName()
	res := &greetpb.GreetWithDeadlineResponse{
		Result: "Hello " + firstName,
	}

	return res, nil

}

func main() {
	fmt.Println("Hello World! I will start serving.....")
	// listen for request, on port 50051(defualt grpc port)
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen; %v", err)
	}

	// create a new server object
	certFile := "../../ssl/server.crt"
	keyFile := "../../ssl/server.pem"
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		log.Fatalf("Failed loading certificates: %v", err)
	}

	opts := grpc.Creds(creds)
	s := grpc.NewServer(opts)

	// this function takes two arguments, grpc.server object and GreetServiceServer interface(in this example)
	// GreetServiceServer interface needs to define a Greet method
	greetpb.RegisterGreetServiceServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
