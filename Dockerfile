FROM golang:1.20-alpine as builder

WORKDIR /app/
COPY . .
RUN go build -v .

FROM alpine:3.17

RUN apk --no-cache add ca-certificates
COPY --from=builder /app/gman /usr/local/bin/
ENTRYPOINT ["gman"]
