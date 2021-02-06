FROM golang:1.14-alpine AS build
WORKDIR /go/src/github.com/g3force/proxy-tcp-udp-mc
COPY go.mod go.mod
COPY cmd cmd
COPY pkg pkg
RUN go install ./...

# Start fresh from a smaller image
FROM alpine:3.9
COPY --from=build /go/bin/proxy-tcp-udp-mc /app/proxy-tcp-udp-mc
ENTRYPOINT ["/app/proxy-tcp-udp-mc"]
CMD []
