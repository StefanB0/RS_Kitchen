# syntax=docker/dockerfile:1
FROM golang:1.19-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY pkg ./pkg

RUN go mod download

COPY *.go ./

RUN go build -o /docker-kitchen

EXPOSE 8881

CMD [ "/docker-kitchen" ]