FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY main.go ./

RUN go build -o fileserver

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/fileserver /app/fileserver

# Create uploads directory
RUN mkdir -p /app/uploads

EXPOSE 8080

CMD ["/app/fileserver"]
