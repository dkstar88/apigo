FROM golang:1.14 AS builder
WORKDIR /go/src/apigo
COPY . .
ENV GO111MODULE=on
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/apigo/app .
EXPOSE 7000
CMD ["./app"]