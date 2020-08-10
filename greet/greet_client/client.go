package main

import (
	"context"
	"fmt"
	"george_custom/grpc_course/greet/greetpb"
	"io"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Hello I'm a client")

	certFile := "../../ssl/ca.crt"
	creds, err := credentials.NewClientTLSFromFile(certFile, "")
	if err != nil {
		log.Fatalf("Error while loading CA trust certificate: %v", err)
	}

	// create connection to the server
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	defer conn.Close()

	// c is GreetServiceClient(interface), has a Greet() function
	c := greetpb.NewGreetServiceClient(conn)
	// fmt.Printf("Created client: %f", c)
	// doUnary(c)
	// doServerStreaming(c)
	// doClientStreaming(c)
	// doBiDiStreaming(c)

	// will implement 3s in server to response after receiving the call
	doUnaryWithDeadline(c, 5*time.Second) // will complete
	doUnaryWithDeadline(c, 1*time.Second) // will timeout

}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do a Unary RPC...")
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "George",
			LastName:  "Chen",
		},
	}

	// return (*GreetResponse, error)
	// call the function on server end Greet()
	response, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling Greet RPC: %v", err)
	}
	log.Printf("Response from Greet: %v", response.GetResult())

}

func doServerStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do a Server Streaming RPC...")
	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Kuan-Chou",
			LastName:  "Chen",
		},
	}

	// check GreetServiceClient interface
	manyResponse, err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling GreetManyTimes RPC: %v", err)
	}

	for {
		msg, err := manyResponse.Recv()
		if err == io.EOF {
			// reach the end of the stream
			break
		}
		if err != nil {
			log.Fatalf("error while reading stream: %v", err)
		}
		log.Printf("Response from GreetManyTimes: %v", msg.GetResult())
	}

}

func doClientStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do a Client Streaming RPC...")

	requests := []*greetpb.LongGreetRequest{
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Roger",
				LastName:  "Federer",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Rafael",
				LastName:  "Nadal",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Novak",
				LastName:  "Djokovic",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Belinda",
				LastName:  "Bancic",
			},
		},
	}

	// return (GreetService_LongGreetClient, error)
	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("error while calling LongGreet: %v", err)
	}

	// we iterate over our slice and send each message individually
	for _, req := range requests {
		fmt.Printf("Sending req: %v\n", req)
		stream.Send(req)
		time.Sleep(100 * time.Millisecond)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error while receiving response from LongGreet: %v", err)
	}
	fmt.Printf("LongGreet Response: %v", res.GetResult())

}

func doBiDiStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do a BiDi streaming RPC...")
	requests := []*greetpb.GreetEveryoneRequest{
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Roger",
				LastName:  "Federer",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Rafael",
				LastName:  "Nadal",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Novak",
				LastName:  "Djokovic",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Belinda",
				LastName:  "Bancic",
			},
		},
	}

	// Send(*GreetEveryoneRequest) error
	// Recv() (*GreetEveryoneResponse, error)

	// we create a stream by invoking the client
	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Error while creating stream: %v", err)
	}

	waitc := make(chan struct{})
	// we send a bunch of messages to the client
	go func() {
		// function to send a bunch of messages
		for _, req := range requests {
			// send each request
			fmt.Printf("Sending message: %v\n", req)
			stream.Send(req)
			time.Sleep(1000 * time.Millisecond)

		}
		stream.CloseSend()
	}()

	// we receive a bunch of messages from the
	go func() {
		// function to receive a bunch of messages
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				// reach the end of the stream

				break
			}
			if err != nil {
				log.Fatalf("error while reading stream: %v", err)

			}
			log.Printf("Response from GreetManyTimes: %v", msg.GetResult())
		}
		close(waitc)

	}()

	// block until everything is done
	<-waitc

}

func doUnaryWithDeadline(c greetpb.GreetServiceClient, timeout time.Duration) {
	fmt.Println("Starting to do a Unary with deadline RPC...")
	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "George",
			LastName:  "Chen",
		},
	}

	// return (*GreetResponse, error)
	// call the function on server end Greet()

	//GreetWithDeadline(ctx context.Context, in *GreetWithDeadlineRequest, opts ...grpc.CallOption) (*GreetWithDeadlineResponse, error)

	// will wait for "timeout" seconds to complete the call, if hasn't received any response within this range, will cancel the call(request)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	response, err := c.GreetWithDeadline(ctx, req)
	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				fmt.Println("Timeout was hit! Deadline was exceeded")
			} else {
				fmt.Printf("unexpected error: %v", statusErr)
			}
		} else {
			log.Fatalf("error while calling Greet RPC: %v", err)
		}
	} else {
		log.Printf("Response from Greet: %v", response.GetResult())

	}

}
