FROM golang:1.21-alpine

WORKDIR /app
COPY . .
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download

RUN go build -o indexer ./

EXPOSE 8081
CMD ["./indexer"]
