FROM golang:alpine AS builder
RUN apk add git
# necessary environmet variables
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container & build
COPY main.go .
RUN go build -o gman main.go 

WORKDIR /app

# Copy binary from build to main folder
RUN cp /build/gman .

# Build a small image
FROM scratch

COPY --from=builder /app/gman /

# Command to run
ENTRYPOINT ["/gman"]