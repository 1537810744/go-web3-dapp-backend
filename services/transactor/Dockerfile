FROM golang:1.21-alpine

WORKDIR /app
COPY . .
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download

RUN go build -o transactor ./

EXPOSE 8080
CMD ["./transactor"]
