generate:
	protoc --proto_path=pkg/grpcapi pkg/grpcapi/*.proto --go_out=plugins=grpc:pkg/grpcapi

runserver:
	go run cmd/server/main.go -port 8888

runclient:
	go run cmd/client/main.go -address 0.0.0.0:8888

runclean:
	rm pkg/grpcapi/*.go 

