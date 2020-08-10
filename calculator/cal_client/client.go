package main

import (
	"context"
	"fmt"
	calculatorpb "george_custom/grpc_course/calculator/calpb"
	"io"
	"log"
	"time"

	"google.golang.org/grpc/codes"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func main() {
	// create connection to the server
	conn, err := grpc.Dial("localhost:50050", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	defer conn.Close()

	// return calculateServiceClient pointer
	c := calculatorpb.NewCalculateServiceClient(conn)
	// doUnary(c)

	// doServerStream(c)
	// doClientStream(c)

	// doBiDiStream(c)
	doSquare(c)

}

func doUnary(c calculatorpb.CalculateServiceClient) {
	req := &calculatorpb.CalculateRequest{
		Calculator: &calculatorpb.Calculate{
			NumOne: 15,
			NumTwo: 25,
		},
	}

	// return CalculateResponse and error
	res, err := c.Calculate(context.Background(), req)
	if err != nil {
		log.Fatalf("error calling grpc %v", err)
	}
	log.Printf("Response from Calculator: %v", res.GetResult())

}

func doServerStream(c calculatorpb.CalculateServiceClient) {
	req := &calculatorpb.PrimeNumberRequest{
		Number: 32078938212,
	}
	fmt.Printf("Decomposing %d .....\n", req.GetNumber())

	stream, err := c.PrimeNumberDecomposition(context.Background(), req)
	if err != nil {
		log.Fatalf("Error when calling server, %v", err)
	}

	n := 1

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("Reach the end of the stream")
			break
		}
		if err != nil {
			log.Fatalf("Something wrong with the streaming, %v", err)
		}
		log.Printf("Number %d prime number is: %v", n, msg.GetResult())
		n++

	}

}

func doClientStream(c calculatorpb.CalculateServiceClient) {
	requests := []*calculatorpb.ComputeAverageRequest{
		&calculatorpb.ComputeAverageRequest{
			Number: 10,
		},
		&calculatorpb.ComputeAverageRequest{
			Number: 20,
		},
		&calculatorpb.ComputeAverageRequest{
			Number: 33,
		},
		&calculatorpb.ComputeAverageRequest{
			Number: 27,
		},
		&calculatorpb.ComputeAverageRequest{
			Number: 46,
		},
		&calculatorpb.ComputeAverageRequest{
			Number: 93,
		},
	}

	stream, err := c.ComputeAverage(context.Background())
	if err != nil {
		log.Fatalf("errro while calling ComputeAverage: %v", err)
	}

	for _, req := range requests {
		fmt.Printf("Sending req: %v\n", req)
		stream.Send(req)
		time.Sleep(100 * time.Millisecond)

	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error while receiving response from ComputeAverage: %v", err)
	}
	fmt.Printf("The average is %v\n", res.GetAnswer())

}

func doBiDiStream(c calculatorpb.CalculateServiceClient) {
	numbers := []int64{1, 5, 6, 3, 4, 8, 10, 9}

	// Send(*FindMaximumRequest) error
	// Recv() (*FindMaximumResponse, error)

	stream, err := c.FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("Error while creating stream: %v", err)
	}

	waitc := make(chan struct{})

	// goroutine for sending request
	go func() {
		for _, num := range numbers {
			req := &calculatorpb.FindMaximumRequest{
				Number: num,
			}
			fmt.Printf("Sending message: %v\n", req)
			stream.Send(req)
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	// goroutine for receiving response
	go func() {
		for {
			// when receiving something from server, always check if it reaches the end
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("error when receiving response: %v", err)
				break
			}
			fmt.Printf("The current Maximum is: %v\n", res.GetResult())
		}
		close(waitc)
	}()

	// block until everything is done
	<-waitc

}

func doSquare(c calculatorpb.CalculateServiceClient) {
	// SquareRoot(ctx context.Context, in *SquareRequest, opts ...grpc.CallOption) (*SquareResponse, error)
	req := &calculatorpb.SquareRequest{
		Number: -10,
	}
	res, err := c.SquareRoot(context.Background(), req)
	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			// actual error from gRPC (user error)
			fmt.Printf("Error message from server: %v\n", respErr.Message())
			fmt.Println(respErr.Code())
			if respErr.Code() == codes.InvalidArgument {
				fmt.Println("We probably sent a negative number!!")
				return
			}
		} else {
			log.Fatalf("error when getting response from server: %v", err)
			return

		}

	}

	fmt.Printf("The square root of %d is %v\n", req.GetNumber(), res.GetNumberRoot())

}
