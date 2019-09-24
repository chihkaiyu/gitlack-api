FROM golang:1.13.0-stretch

ENV GO111MODULE=on

RUN mkdir -p /home/gitlack/db

WORKDIR /home/gitlack

COPY . /home/gitlack

RUN GOOS=linux go build -a -v -o main ./cmd

ENTRYPOINT [ "./main" ]