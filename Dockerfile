FROM golang:alpine3.13

RUN mkdir -p /app
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY database/ database
COPY docs/ docs
COPY pages/ pages
COPY public/ public
COPY web/ web

RUN go build -o main .

EXPOSE 8080

ENTRYPOINT ["./main"]
