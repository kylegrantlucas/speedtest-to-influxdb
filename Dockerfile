FROM golang:1.10.1-alpine as builder
WORKDIR /go/src/github.com/kylegrantlucas/speedtest-to-influxdb
COPY . .
RUN go build -o /bin/application .

FROM alpine:latest
WORKDIR /root
COPY --from=builder /bin/application .
ENTRYPOINT ["/root/application"]