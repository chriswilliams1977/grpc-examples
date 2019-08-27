#specify full proto filename
#specify plugins type 
protoc src/greet/greetpb/greet.proto --go_out=plugins=grpc:.

#generates proto buffer for calculator service
protoc src/calculator/calculatorpb/calculator.proto --go_out=plugins=grpc:.