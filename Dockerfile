FROM golang:latest as build
WORKDIR /go/src/github.com/kylegrantlucas/speedtest-to-influxdb
COPY . .
RUN go build -o app .

FROM gcr.io/distroless/base
COPY --from=build /go/src/github.com/kylegrantlucas/speedtest-to-influxdb /
CMD ["/app"]