FROM golang:1.15-alpine AS build

WORKDIR /src

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /bin/vulture ./cmd/vulture

FROM scratch
COPY --from=build /bin/vulture /
CMD ["/vulture"]
