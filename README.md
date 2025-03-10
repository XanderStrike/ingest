# Ingest

A hole you can drop files into. Saves them to a folder on your server, that's it.

### Build and run

```bash
docker build -t ingest .
docker run -p 8080:8080 -v $(pwd)/uploads:/app/uploads ingest
```

Note: The templates are included in the Docker image, so you don't need to mount them.

### Configuration

You can configure the maximum file size using the `MAX_FILE_SIZE` environment variable (in bytes):

```bash
# Set maximum file size to 100MB
docker run -p 8080:8080 -e MAX_FILE_SIZE=104857600 -v $(pwd)/uploads:/app/uploads ingest

# Set unlimited file size (default)
docker run -p 8080:8080 -e MAX_FILE_SIZE=0 -v $(pwd)/uploads:/app/uploads ingest
```

### Access the service

Open your browser and navigate to:
```
http://localhost:8080
```
