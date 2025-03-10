# Ingest

A hole you can drop files into. Saves them to a folder on your server, that's it.

### Build and run

```bash
docker build -t ingest .
docker run -p 8080:8080 -v $(pwd)/uploads:/app/uploads ingest
```

### Access the service

Open your browser and navigate to:
```
http://localhost:8080
```