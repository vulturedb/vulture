FROM golang:1.15

WORKDIR /go/src/app
COPY . .

RUN go install -v ./...

EXPOSE 6667
CMD ["vulture"]
