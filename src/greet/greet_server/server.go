package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/chriswilliams1977/go-grpc-course/src/greet/greetpb"

	"google.golang.org/grpc"
)

//SERVICE OBJECT
//as we add services we add them as an implementation of this type
//you must add interface implementation for Service Server
//check server API interface for GreetService service
type server struct{}

//for our server type use grpc we must implement interface
//you must implement this interface for GreetServiceServer as defined in greet.pb.go
func (*server) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	fmt.Printf("Greet function was invoked with %v", req)
	//use req object to get information out of request
	firstname := req.GetGreeting().GetFirstName()
	result := "Hello " + firstname
	//GreetResponse takes a string as the result
	res := &greetpb.GreetResponse{
		Result: result,
	}

	return res, nil
}

//takes a request and a stream
func (*server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {

	firstname := req.GetGreeting().GetFirstName()
	for i := 0; i < 10; i++ {
		result := "Hello " + firstname + " number " + strconv.Itoa(i)
		res := &greetpb.GreetManyTimesResponse{
			Result: result,
		}
		//sends a stream of 10 messages
		stream.Send(res)
		time.Sleep(1000 * time.Millisecond)
	}
	//end server streaming
	return nil
}

//to get implementation check GreetServiceServer interface
//takes a stream
func (*server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {

	fmt.Printf("LongGreet func was invoked with a streaming request\n")

	result := ""

	for {
		//while the server recieves a stream do this
		req, err := stream.Recv()
		if err == io.EOF {
			//when we hit the end of the filw we have finished reading the stream
			//if this does not work an error will be returned
			return stream.SendAndClose(&greetpb.LongGreetResponse{
				Result: result,
			})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream %v", err)
		}
		//get firstname
		firstName := req.GetGreeting().GetFirstName()
		result += "Hello" + firstName + "!"
	}
}

func (*server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {
	fmt.Printf("GreetEveryone func was invoked with a streaming request\n")

	//NOTE: Server decides when to end stream using SendAndClose()
	//Here we just keep responding as an example

	//Iterate over stream
	//everytime we receive request from client, respond to it

	for {
		//RECIEVE STREAM
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading client stream : %v", err)
			return err
		}
		firstname := req.GetGreeting().GetFirstName()
		result := "Hello" + firstname + "! "
		//STREAM RESPONSE
		sendErr := stream.Send(&greetpb.GreetEveryoneResponse{
			Result: result,
		})
		if sendErr != nil {
			log.Fatalf("Error while ending data to client : %v", sendErr)
			return sendErr
		}
	}
}

func main() {
	fmt.Println("Hello World")

	//CREATE A CONNECTION/LISTENER (connection on port binding)
	//take args for network and address
	//we open a tcp for network
	//50001 is the default port (address) for grpc
	lis, err := net.Listen("tcp", "0.0.0.0:50001")
	if err != nil {
		//for example cannot bind to port
		log.Fatalf("Failed to listen: %v", err)
	}

	//CREATE A SERVER
	s := grpc.NewServer()

	//REGISTER SERVICE
	//takes a grpc server and a type
	greetpb.RegisterGreetServiceServer(s, &server{})

	//SERVE SERVICE FROM SERVER
	//if server cannot serve throw error otherwise serve
	//shorthand for if in go
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
