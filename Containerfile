FROM golang

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/sse-broker ./...

EXPOSE 8088

CMD ["sse-broker"]
