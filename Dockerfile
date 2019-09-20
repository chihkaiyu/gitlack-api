FROM golang:1.11.3-stretch

ENV GO111MODULE=on

ENV GOPROXY=https://proxy.golang.org

RUN mkdir -p /home/gitlack/db

WORKDIR /home/gitlack

COPY . /home/gitlack

RUN GOOS=linux go build -a -v -o main ./cmd

ENTRYPOINT [ "./main" ]