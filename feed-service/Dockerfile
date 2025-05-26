FROM golang:1.24.0 AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main .
EXPOSE 3000
CMD ["./main"]