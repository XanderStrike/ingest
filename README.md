# Ingest

A hole you can drop files into. Saves them to a folder on your server, that's it.

![demo](https://github.com/user-attachments/assets/f71afd16-a49b-490e-8a7e-7fe583dcff7d)


### Run

```bash
docker run -p 8080:8080 -v $(pwd)/uploads:/app/uploads xanderstrike/ingest
```

### Configuration

You can configure the maximum file size using the `MAX_FILE_SIZE` environment variable (in bytes):

```bash
# Set maximum file size to 100MB
docker run -p 8080:8080 -e MAX_FILE_SIZE=104857600 -v $(pwd)/uploads:/app/uploads xanderstrike/ingest

# Set unlimited file size (default)
docker run -p 8080:8080 -e MAX_FILE_SIZE=0 -v $(pwd)/uploads:/app/uploads xanderstrike/ingest
```

### Access the service

Open your browser and navigate to:
```
http://localhost:8080
```
