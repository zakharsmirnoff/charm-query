FROM golang:1.21

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o charm-query .

EXPOSE 5000

CMD ["./charm-query"]
