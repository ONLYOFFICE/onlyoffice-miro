FROM golang:1.24 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o /app/main ./cmd/server

FROM alpine:3
WORKDIR /app
COPY --from=builder /app/main ./main
ENTRYPOINT ["./main"]
