FROM golang:alpine3.13

RUN mkdir -p /app
WORKDIR /app

COPY database/ database
COPY docs/ docs
COPY go.mod go.sum main.go ./
COPY pages/ pages
COPY public/ public
COPY web/ web

RUN go build -o main .

EXPOSE 80

ENTRYPOINT ["./main"]
