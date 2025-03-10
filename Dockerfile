FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY main.go ./

RUN go build -o ingest

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/ingest /app/ingest
COPY templates/ /app/templates/

# Create uploads directory
RUN mkdir -p /app/uploads

EXPOSE 8080

CMD ["/app/ingest"]
