FROM golang:1.17-alpine as builder

RUN apk --no-cache update &&\
    apk add gcc musl-dev git

WORKDIR /go/src/app/ethash-mining-pool

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o ethereum-pool main.go

FROM alpine
WORKDIR /app
COPY --from=builder /go/src/app/ethash-mining-pool/ethereum-pool .
ENTRYPOINT [ "./ethereum-pool" ]