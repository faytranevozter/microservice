FROM golang:1.22 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .

FROM alpine:3.19
RUN adduser -D app && apk add --no-cache ca-certificates
USER app
WORKDIR /home/app
COPY --from=builder /app/server ./server
EXPOSE 8082
ENV SERVICE_PORT=8082
CMD ["./server"]


