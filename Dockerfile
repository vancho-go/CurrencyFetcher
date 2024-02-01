FROM golang:1.20.7-alpine3.18
LABEL authors="vancho"

RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go build -o main ./cmd/server/main.go

EXPOSE 8080
CMD ["/app/main"]