FROM golang:1.17-alpine

COPY go.mod /go.mod
COPY main.go /main.go
COPY entrypoint.sh /entrypoint.sh

RUN go build -o /main /main.go

ENTRYPOINT ["/entrypoint.sh"]
