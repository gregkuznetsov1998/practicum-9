FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod ./
COPY *.go ./

RUN go build -o bionicpro-auth

EXPOSE 8000

CMD ["./bionicpro-auth"]