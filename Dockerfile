FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY pkg ./pkg

RUN go build -o server ./cmd 

FROM alpine:latest
COPY --from=builder /app/server ./server

# Expose the port that the server listens on
EXPOSE 8080

# Set the entry point for the container
CMD ["./server"]
