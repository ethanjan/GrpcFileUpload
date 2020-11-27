module github.com/ethanjan/grpcupload

go 1.14

replace github.com/ethanjan/grpcupload/transfer => ./internal/pkg/transfer

require (
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.1.1
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.6.0 // indirect
	google.golang.org/grpc v1.32.0
	google.golang.org/protobuf v1.24.0
)
