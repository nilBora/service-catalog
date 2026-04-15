FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o manager-services ./app

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/manager-services .
VOLUME ["/data"]
EXPOSE 8090
CMD ["./manager-services", "--db=/data/services.db", "--address=:8090"]
