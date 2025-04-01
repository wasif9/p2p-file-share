FROM golang:1.24 AS builder

WORKDIR /app
COPY bootstrap/main.go main.go
COPY go.mod go.sum ./

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

RUN go build -o bootstrap-node main.go

FROM scratch
COPY --from=builder /app/bootstrap-node /bootstrap-node

EXPOSE 4001
CMD ["/bootstrap-node"]
