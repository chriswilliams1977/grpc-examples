package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/chriswilliams1977/go-grpc-course/src/calculator/calculatorpb"

	"google.golang.org/grpc"
)

//SERVICE OBJECT
//as we add services we add them as an implementation of this type
type server struct{}

//adding implementation of Sum service to this type
func (*server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	fmt.Printf("Sum function was invoked with %v", req)
	//use req object to get information out of request
	intA := req.GetIntA()
	intB := req.GetIntB()
	result := intA + intB
	//CalcukatorResponse takes a int as the result
	res := &calculatorpb.SumResponse{
		Result: result,
	}

	return res, nil
}

//Prime number decompostion means dividing a number by prime numbers
//for example 12
//2 is a prime number
//so we do 2*6=12
//6 is not prime so we divide it again
//2*3 = 6
//so the prime decomposition of 12 is 2*2*3
func (*server) PrimeNumberDecomposition(req *calculatorpb.PrimeNumberDecompositionRequest, stream calculatorpb.CalculatorService_PrimeNumberDecompositionServer) error {
	fmt.Printf("PrimeNumberDecomposition was invoked with %v", req)
	//number passed by client
	number := req.GetNumber()

	divisor := int64(2)

	//check how many times divisor can go into number
	//10%1=0, 10%2=0, 10%3=1 (3 goes 3 times with remainder 1)
	//so in this case 12%2=0 as 2 can go 6 times
	for number > 1 {

		if number%divisor == 0 {
			//stream response
			stream.Send(&calculatorpb.PrimeNumberDecompositionResponse{
				PrimeFactor: divisor,
			})
			number = number / divisor
		} else {
			divisor++
			fmt.Printf("Divisor has increased to %v", divisor)
		}
	}

	//end server streaming
	return nil
}

func (*server) ComputeAverage(stream calculatorpb.CalculatorService_ComputeAverageServer) error {
	fmt.Printf("ComputeAverage was invoked with streaming request\n")

	sum := int32(0)
	count := 0

	for {
		//while the server recieves a stream do this
		req, err := stream.Recv()
		if err == io.EOF {
			average := float64(sum) / float64(count)
			//when we hit the end of the filw we have finished reading the stream
			//if this does not work an error will be returned
			return stream.SendAndClose(&calculatorpb.ComputeAverageResponse{
				Average: average,
			})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream %v", err)
		}
		//Get req number from stream
		sum += req.GetNumber()
		count++
	}
}

func (*server) FindMaximum(stream calculatorpb.CalculatorService_FindMaximumServer) error {
	fmt.Printf("FindMaximum was invoked with streaming request\n")

	maxNumber := int32(0)

	for {
		//while the server recieves a stream do this
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading client stream %v", err)
			return err
		}
		number := req.GetNumber()
		//Get req number from stream
		if number > maxNumber {
			maxNumber = number
			sendErr := stream.Send(&calculatorpb.FindMaximumResponse{
				MaxNumber: maxNumber,
			})
			if sendErr != nil {
				log.Fatalf("Error while ending data to client : %v", sendErr)
				return sendErr
			}
		}

	}
}

func main() {
	//CREATE A CONNECTION/LISTENER (connection on port binding)
	lis, err := net.Listen("tcp", "0.0.0.0:50001")
	if err != nil {
		//for example cannot bind to port
		log.Fatalf("Failed to listen: %v", err)
	}

	//CREATE A SERVER
	s := grpc.NewServer()

	//REGISTER SERVICE
	//RegisterCalculatorServiceServer take two params
	//grpc server and a service in this case it implements CalculatorServiceServer interface
	//our server must therefore implement this interface as per Sum func above
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	//SERVE SERVICE FROM SERVER
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
