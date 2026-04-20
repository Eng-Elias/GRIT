FROM golang:1.22-alpine AS builder

RUN apk add --no-cache build-base git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/grit ./cmd/grit/

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata git

WORKDIR /app
COPY --from=builder /app/grit .

EXPOSE 8080

CMD ["./grit"]
