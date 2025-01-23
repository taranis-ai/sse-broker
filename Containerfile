FROM golang AS builder


WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$(go env GOARCH) go build -a -installsuffix cgo -o sse-broker ./...

FROM scratch

LABEL description="SSE Broker"

COPY --from=builder /usr/src/app/sse-broker /usr/local/bin/sse-broker

EXPOSE 8088

CMD ["/usr/local/bin/sse-broker"]