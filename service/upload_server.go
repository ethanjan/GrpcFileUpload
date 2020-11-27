package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	pb "github.com/ethanjan/grpcupload/pkg/grpcapi"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// This sets maxImageSize to ~4.29Gb.
const (
	maxImageSize = 1 << 32
)

// This UploadServer type contains the imagePath as a string.
type UploadServer struct {
	imagePath string
}

// This UploadSaveImage is an interface that implements the UploadServer type.
type UploadSaveImage interface {
	UploadImage(stream pb.UploadService_UploadImageServer) (err error)
}

// This NewUploadServer returns a new UploadServer.
func NewUploadServer() *UploadServer {
	return &UploadServer{}
}

// This UploadImage is a client-streaming RPC to upload images.
func (server *UploadServer) UploadImage(stream pb.UploadService_UploadImageServer) (err error) {

	// This creates a new randomized universally unique identifier.
	imageID, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("Failed to generate image id: %w", err)
	}

	req, err := stream.Recv()
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot receive image info"))
	}

	// This requests information about the image type and sets the image path.
	imageType := req.GetInfo().GetImageType()
	imagePath := fmt.Sprintf("%s/%s%s", "destinationstore", imageID, imageType)

	// This actually creates the file for a given image path.
	file, err := os.Create(imagePath)
	if err != nil {
		return fmt.Errorf("Failed to create image file: %w", err)
	}

	// This creates the buffers to hold the chunked data, the size of the image,
	// the receivizng size of the file, and the amount of progress in transferring the file.
	imageData := bytes.Buffer{}
	var imageSize int64
	receivingSize := req.GetInfo().GetSize()
	progressInterval := 0

	for {
		// This captures any errors that may arise from context.
		err := contextError(stream.Context())
		if err != nil {
			return err
		}

		req, err := stream.Recv()
		if err == io.EOF { // If the end of the filed has been reached, break out of the loop.
			break
		}

		if err != nil {
			return logError(status.Errorf(codes.Unknown, "Failed to receive chunk data: %v", err))
		}

		// This receives the next chunk and assigns the size.
		chunk := req.GetChunkData()
		size := len(chunk)

		// This adds the size of the chunk to the total image size.
		imageSize += int64(size)

		// This will return an error if the total image size is greater
		// than the maximum allowed image size (approximately 4.29 Gigabytes).
		if imageSize > maxImageSize {
			return logError(status.Errorf(codes.InvalidArgument, "image is too large: %d > %d", imageSize, maxImageSize))
		}

		// This makes sure that the percentage calculated does not divide by 0.
		if receivingSize > 0 {
			// This finds the upload percentage.
			uploadPercent := (imageSize * 100 / receivingSize)
			// If the truncated upload percentage is greater than the progress interval,
			// then show the amount of progress.
			if int(uploadPercent) > progressInterval {
				log.WithFields(log.Fields{"CurrentSize": imageSize}).Infof("Upload progress: %v%% completed", uploadPercent)
				// This increments the progress interval so that progress is only shown in increments of 20%.
				progressInterval += 20
			}
		}

		// This writes the data to the buffer.
		_, err = imageData.Write(chunk)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "Failed to write data: %v", err))
		}

		// This writes the data to the file itself and returns any errors that may arise.
		_, err = imageData.WriteTo(file)
		if err != nil {
			return fmt.Errorf("Failed to write image to file: %w", err)
		}

	}

	// This formats the response using grpc.
	res := &pb.UploadImageResponse{
		Id:   imageID.String(),
		Size: int64(imageSize),
	}

	// This sends the stream back to the client and closes the connection.
	err = stream.SendAndClose(res)
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "Failed to send response: %v", err))
	}

	log.Printf("Saved image with id: %s, size: %d", imageID, imageSize)
	return nil
}

// This handles context errors.
func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return logError(status.Error(codes.Canceled, "Request is canceled."))
	case context.DeadlineExceeded:
		return logError(status.Error(codes.DeadlineExceeded, "Deadline is exceeded."))
	default:
		return nil
	}
}

// This defines the logError function that is frequently used in the above functions.
func logError(err error) error {
	if err != nil {
		log.Print(err)
	}
	return err
}
