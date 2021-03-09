FROM golang:latest AS builder
WORKDIR /build
ADD . .
RUN go mod download && go mod verify
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-w -s"
RUN strip peloton-go

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/peloton-go /opt/
EXPOSE 9000
CMD ["/opt/peloton-go"]
