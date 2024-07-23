FROM golang:1.22-bookworm

RUN apt-get update && apt-get install -y debhelper
# Dev dependencies
RUN apt-get install -y systemd vim tree

WORKDIR /src

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=arm GOARM=6 go build ./cmd/nmealogger
RUN GOOS=linux GOARCH=arm GOARM=6 go build ./cmd/logupload
RUN GOOS=linux GOARCH=arm GOARM=6 go build ./cmd/gpsdfilter

RUN dpkg-buildpackage -us -uc
RUN ls -l ../*.deb
