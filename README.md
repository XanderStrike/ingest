# Simple File Upload Service

A lightweight web service written in Go that allows users to upload files to the server.

## Features

- Simple web interface for uploading files
- View and download previously uploaded files
- Docker support for easy deployment
- Configurable port via environment variable

## Quick Start with Docker

### Build the Docker image

```bash
docker build -t fileupload-service .
```

### Run the container

```bash
docker run -p 8080:8080 -v $(pwd)/uploads:/app/uploads fileupload-service
```

This command:
- Maps port 8080 from the container to port 8080 on your host
- Creates a volume that persists uploaded files to the `uploads` directory in your current folder

### Access the service

Open your browser and navigate to:
```
http://localhost:8080
```

## Configuration

You can configure the service using the following environment variables:

- `PORT`: The port the server will listen on (default: 8080)

Example with custom port:
```bash
docker run -p 9000:9000 -e PORT=9000 -v $(pwd)/uploads:/app/uploads fileupload-service
```

## Running without Docker

If you prefer to run the service without Docker:

1. Make sure you have Go installed
2. Clone this repository
3. Run the following commands:

```bash
go build -o fileserver
./fileserver
```

The service will be available at http://localhost:8080
