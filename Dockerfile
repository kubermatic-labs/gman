FROM golang:1.14-alpine as builder
ENV GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /app/
COPY . .
RUN go mod download
RUN go build -o gman main.go 

FROM alpine:3.12

RUN apk --no-cache add ca-certificates
COPY --from=builder /app/gman .
ENTRYPOINT ["/gman"]
