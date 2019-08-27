//create client to connect to server
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/chriswilliams1977/go-grpc-course/src/greet/greetpb"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Hello I am a client")

	//CREATE CONNECTION TO SERVER
	//grpc.dial takes two args
	//first is conn address
	//options - by defaukt grpc has ssl but for now we use insecure conn
	conn, err := grpc.Dial("localhost:50001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect %v", err)
	}

	//Close connection at end of program
	//go defer runs this at end of program
	defer conn.Close()

	//CREATE CLIENT
	c := greetpb.NewGreetServiceClient(conn)
	//fmt.Printf("Created the client %f", c)

	//doUnary(c)

	//doServerStreaming(c)

	//doClientStreaming(c)

	doBiDiStreaming(c)

}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do Unary rpc")
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Chris",
			LastName:  "Williams",
		},
	}
	res, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling Greet rpc: %v", err)
	}
	log.Printf("Response from Greet: %v", res.Result)
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting server streaming RPC")

	//CREATE REQUEST
	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Chris",
			LastName:  "Williams",
		},
	}

	//PLACE RPC CALL
	//WE GET A STREAM BACK
	//the GreetManyTimes function returns GreetService_GreetManyTimesClient, error
	resStream, err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling GreetManyTimes RPC %v", err)
	}

	//LOOP STREAM UNTIL WE REACH EOF
	for {
		//can call Recv ass many times as you want until you reach end of file (EOF)
		msg, err := resStream.Recv()
		if err == io.EOF {
			//we have reached end of stream
			break
		}

		if err != nil {
			log.Fatalf("erro while reading stream: %v", err)
		}

		log.Printf("Response from GreetManytimes : %v", msg.GetResult())
	}

}

func doClientStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting client streaming RPC")

	//CREATE REQUESTS
	//our stream.send below takes a LongGreetRequest
	//we create a array of these
	//*greetpb.LongGreetRequest get current object
	//&greetpb.LongGreetRequest set value of current object
	requests := []*greetpb.LongGreetRequest{
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Chris",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Sieg",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Luna",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Liv",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Bob",
			},
		},
	}

	//CREATE STREAM USING LongGreet
	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("error while calling LongGreet: %v", err)
	}

	//SEND DATA TO STREAM
	//iterate iterate over slice and send each message individually
	for _, req := range requests {
		fmt.Printf("Sending req %v\n", req)
		//DATA SEND POINT
		stream.Send(req)
		time.Sleep(1000 * time.Millisecond)
	}

	//CLOSE STREAM
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error while recieving response from LongGreet: %v", err)
	}

	fmt.Printf("LongGreet response : %v", res)
}

func doBiDiStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting BiDi streaming RPC")

	//create a stream by invoking client

	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Error while creating stream: %v", err)
		return
	}

	//create requests for the stream
	requests := []*greetpb.GreetEveryoneRequest{
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Chris",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Sieg",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Luna",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Liv",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Bob",
			},
		},
	}

	//way to block is to use a wait channel
	//this makes a channel with empty structs we can use to block
	waitc := make(chan struct{})

	//Stream is good so
	//sSEND MESSAGES to client via go routine
	//() at start and end means define and invoke func
	//this function runs in its own go routine
	//Send 5 messages then close
	go func() {
		//function to send a bunch a messages
		//iterate over requests
		for _, req := range requests {
			//send request in stream
			fmt.Printf("sending message : %v\n", req)
			stream.Send(req)
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
			fmt.Printf("recieved : %v\n", res.GetResult())
		}
		close(waitc)
	}()

	//block until everything is done
	//waits for channel to be closed
	<-waitc
}
