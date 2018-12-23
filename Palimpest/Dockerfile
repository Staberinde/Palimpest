FROM golang:1.10.5-alpine

WORKDIR /Go/src/Palimpest
#ENV GOPATH=/Palimpest/vendor
ENV GOPATH=/Go

COPY main.go ./main.go
COPY main_test.go ./main_test.go

RUN apk add -u \
    git \
    curl \
    bash \
    postgresql

RUN mkdir -p /Go/bin
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
RUN [[ -f Gopkg.toml ]] || /Go/bin/dep init
RUN /Go/bin/dep ensure

RUN go get -d -v ./...

RUN go install -v ./...
#ENV GOPATH=/Go/src/Palimpest
