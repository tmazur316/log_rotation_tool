# syntax=docker/dockerfile:1

FROM golang:1.19-alpine

WORKDIR /go/src/log_rotation_tool
COPY . /go/src/log_rotation_tool
RUN go mod download
RUN go build -o ./log_rotation_tool
CMD ["./log_rotation_tool"]