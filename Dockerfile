FROM golang:1.22-alpine AS builder
WORKDIR /app

# Cache deps first
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /srv

# Copy binary and public assets
COPY --from=builder /app/app ./app
COPY public ./public

ENV PORT=8080
EXPOSE 8080

CMD ["./app"]


