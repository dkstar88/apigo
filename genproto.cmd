protoc --go_out=. --go-grpc_out=. ./grpc/*.proto
protoc --go_out=plugins=grpc:. ./grpc/*.proto