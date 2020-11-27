package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/ethanjan/grpcupload/pkg/grpcapi"
	"github.com/ethanjan/grpcupload/service"
	"google.golang.org/grpc"
)

func main() {
	// This specifies a new local listening port, which can be changed with flags.
	port := flag.Int("port", 8888, "Listening port.")
	flag.Parse()

	log.Printf("Starting server on port: %d", *port)

	// This creates a new uploadServer and grpcServer.
	uploadServer := service.NewUploadServer()
	grpcServer := grpc.NewServer()

	// This registers a new grpc service.
	pb.RegisterUploadServiceServer(grpcServer, uploadServer)

	// This starts the listener for a given port and address.
	address := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("Failed to start listener: ", err)
	}

	// This starts the grpc server.
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("Failed to start grpc server.: ", err)
	}
}
