package main

import (
	"bufio"
	"context"
	"flag"
	"io"
	"os"
	"path/filepath"
	"time"

	pb "github.com/ethanjan/grpcupload/pkg/grpcapi"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// This sets the maxImageSize to ~4.29Gb.
const (
	maxImageSize = 1 << 32
)

func main() {
	// This defaults to the file "sourcestore/file.txt" and the port "8888" on the IP address "0.0.0.0."
	filePtr := flag.String("filename", "sourcestore/file.txt", "specify file name for transfer")
	serverAddress := flag.String("address", "0.0.0.0:8888", "the server address")
	flag.Parse()
	log.Printf("Connecting to server %s", *serverAddress)

	// This attempts to create an insecure connection between the client and the server.
	conn, err := grpc.Dial(*serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatal("Failed to connect to server: ", err)
	}

	// This actually begins the process of the image upload.
	uploadClient := pb.NewUploadServiceClient(conn)
	uploadImage(uploadClient, *filePtr)
}

func uploadImage(uploadClient pb.UploadServiceClient, imagePath string) {
	// This opens the file for reading.
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatal("Failed to open image file: ", err)
	}
	// This ensures that the file isn't closed until the very end of the function.
	defer file.Close()

	// This attempts to get the stats about the file and assigns the info.Size() value to imageSize.
	info, err := os.Stat(imagePath)
	if err != nil {
		log.Fatalf("Cannot access image file: %v", err)
	}
	imageSize := info.Size()

	// This exits if the image is not within the specified file size limits.
	if imageSize > maxImageSize {
		log.Fatalf("Upload image is too large. Want < %v, Got %v", int64(maxImageSize), imageSize)
	}

	// This sets the context timeout of 1 minute.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	// This calls the cancel function to release resources.
	defer cancel()

	// This actually begins the process of uploading the image.
	stream, err := uploadClient.UploadImage(ctx)
	if err != nil {
		log.Fatal("Failed to upload image: ", err)
	}

	// This builds the request for the image upload.
	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				ImageType: filepath.Ext(imagePath),
				Size:      imageSize,
			},
		},
	}

	err = stream.Send(req)
	if err != nil {
		log.Fatal("Failed to send image info to the server: ", err, stream.RecvMsg(nil))
	}

	reader := bufio.NewReader(file)
	// Ideally, the buffer size should be anywhere from 16K to 64K.
	// Larger buffer sizes would decrease the number of garbage collectors.
	buffer := make([]byte, 1024)

	// This defines new variables to represent how much of the file has been transferred
	// and the progress of this transfer process.
	var (
		sentSize         int64
		progressInterval int
	)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF { // If we have reached the end of file, we must break out of the loop.
			break
		}
		if err != nil {
			log.Fatal("Failed to read chunk to buffer: ", err)
		}

		// This builds the request for uploading a chunk of data.
		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil {
			log.Fatal("Failed to send chunk to server: ", err, stream.RecvMsg(nil))
		}

		// This finds the size of the uploaded chunk and adds it to the sentSize variable
		// to represent the amount of data that has already been sent.
		uploadedChunk := len(buffer[:n])
		sentSize += int64(uploadedChunk)

		// This prevents the percentage calculation from dividing by zero.
		if imageSize > 0 {
			// This calculates the percent of data that has been uploaded.
			uploadPercent := (sentSize * 100 / imageSize)
			// If the percent of data that has been uploaded is greater than the interval,
			// then it prints the progress.
			if int(uploadPercent) > progressInterval {
				log.Printf("Image push progress: %v%%. Uploaded size: %d", uploadPercent, sentSize)
				progressInterval += 20
			}
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("Failed to receive response: ", err)
	}

	// Prints if the entire file has been transferred successfully.
	log.Printf("Upload file with id: %s, size: %d", res.GetId(), res.GetSize())
}
