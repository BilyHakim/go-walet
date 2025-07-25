FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:3.17

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/app .

COPY .env .

EXPOSE 8080

CMD ["./app"]
