#the commands (generate) bellow was got in https://grpc.io/docs/languages/go/quickstart/
greet_generate:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative greet/greetpb/greet.proto

calculator_generate:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative calculator/calculatorpb/calculator.proto