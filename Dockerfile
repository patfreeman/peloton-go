FROM golang:latest AS builder
WORKDIR /build
ADD . .
RUN go mod download && go mod verify
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-w -s" -o go-peloton && strip go-peloton

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/go-peloton /opt/
EXPOSE 9000
CMD ["/opt/go-peloton"]
