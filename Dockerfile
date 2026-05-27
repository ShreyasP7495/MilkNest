FROM golang:1.22-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/milknest ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

ENV TZ=Asia/Kolkata

WORKDIR /app
COPY --from=builder /out/milknest /app/milknest
COPY migrations /app/migrations

EXPOSE 8080

ENTRYPOINT ["/app/milknest"]
