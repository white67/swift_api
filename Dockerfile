FROM golang:1.24 AS builder

WORKDIR /app
COPY . .

RUN go mod tidy
RUN GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o main ./swift_api

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

RUN chmod +x /root/main

EXPOSE 8080

CMD ["/root/main"]