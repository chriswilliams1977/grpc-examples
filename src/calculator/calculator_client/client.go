package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/chriswilliams1977/go-grpc-course/src/calculator/calculatorpb"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Hello I am a client")

	//CREATE CONNECTION TO SERVER
	conn, err := grpc.Dial("localhost:50001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect %v", err)
	}

	//Close connection at end of program
	//go defer runs this at end of program
	defer conn.Close()

	//CREATE CLIENT
	//Client take a connection to server as param
	c := calculatorpb.NewCalculatorServiceClient(conn)

	//doUnary(c)

	//doServerStreaming(c)

	//doClientStreaming(c)

	doBiDiStreaming(c)
}

func doUnary(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do Unary rpc")
	req := &calculatorpb.SumRequest{
		IntA: 3,
		IntB: 10,
	}
	//sum service takes a context and a request in this case type CalculatorRequest
	//create request above with relevant params and pass it to sum service
	res, err := c.Sum(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling Calculator rpc: %v", err)
	}
	log.Printf("Response from Calculator: %v", res.Result)
}

func doServerStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do PrimeDecomposition Server Streaming RPC...")

	req := &calculatorpb.PrimeNumberDecompositionRequest{
		Number: 12,
	}
	//PrimeNumberDecomposition service takes a context and a request in this case type PrimeNumberDecompositionRequest
	//create request above with relevant params and pass it to PrimeNumberDecomposition service
	stream, err := c.PrimeNumberDecomposition(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling PrimeNumberDecomposition rpc: %v", err)
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			//reached end of stream
			break
		}
		if err != nil {
			log.Fatalf("Something happened: %v", err)
		}
		fmt.Println(res.GetPrimeFactor())
	}

}

func doClientStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting client streaming RPC")

	//CREATE STREAM USING LongGreet
	stream, err := c.ComputeAverage(context.Background())
	if err != nil {
		log.Fatalf("error while calling ComputeAverage: %v", err)
	}

	numbers := []int32{3, 5, 9, 54, 23}

	//SEND DATA TO STREAM
	//iterate iterate over slice and send each message individually
	for _, number := range numbers {
		fmt.Printf("Sending number : %v\n", number)
		//DATA SEND POINT
		stream.Send(&calculatorpb.ComputeAverageRequest{
			Number: number,
		})
		time.Sleep(1000 * time.Millisecond)
	}

	//CLOSE STREAM
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error while recieving response from ComputeAverage: %v", err)
	}

	fmt.Printf("The average is : %v", res.GetAverage())
}

func doBiDiStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting FindMaximum BiDi streaming RPC")

	//create a stream by invoking client
	stream, err := c.FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("Error while opening stream: and calling FindMaximum %v", err)
		return
	}

	//way to block is to use a wait channel
	waitc := make(chan struct{})

	//SEND MESSAGES to client via go routine
	go func() {
		//create requests for the stream
		numbers := []int32{1, 5, 3, 6, 2, 20}

		//function to send a bunch a messages
		//iterate over requests
		for _, number := range numbers {
			//send request in stream
			fmt.Printf("sending message : %v\n", number)
			stream.Send(&calculatorpb.FindMaximumRequest{
				Number: number,
			})
			time.Sleep(1000 * time.Millisecond)
		}
		//finished sending stream
		stream.CloseSend()
	}()

	//RECEIVE MESSAGES from client
	go func() {
		//loop over messages
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				//if we are at the end of the stream close wait channel
				//this unblocks everything
				break
			}
			if err != nil {
				log.Fatalf("error while recieving %v", err)
				break
			}
			//print the result
			fmt.Printf("recieved : %v\n", res.GetMaxNumber())
		}
		close(waitc)
	}()

	//block until everything is done
	//waits for channel to be closed
	<-waitc
}
