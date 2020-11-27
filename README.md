# GRPC File Upload
Learning golang grpc-protobuff client-server implementation for uploading files.

## Implementation
This software was implemented by utilizing the programming language known as Golang.

## Prerequisites
Golang and grpc/protobuff utilities must be installed.

## Usage
The Source file is "sourcestore/file.txt," which can be overriden by flags.

The Destination file is located in the "destinationstore" directory.

The Maximum Upload Size is: Approximately 4.29 Gigabytes

How to run the server: ```make runserver```

How to run the client: ```make runclient```

How to generate protocol buffer file: ```make generate```

How to remove all the auto-generated grpc files: ```make runclean```

## Images
Starting Server:

![Screen Shot 2020-11-25 at 2 05 57 PM](https://user-images.githubusercontent.com/8474410/100286318-72221b80-2f27-11eb-939c-69f7aea08462.png)

Running Client:

![Screen Shot 2020-11-27 at 9 23 34 AM](https://user-images.githubusercontent.com/8474410/100473038-50917300-3092-11eb-87a4-84b29503e806.png)
