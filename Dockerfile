# The build stage
FROM golang:1.21-alpine as builder
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o chat-app .

# The run stage
FROM alpine:latest
RUN apk add --no-cache ca-certificates sqlite-libs
WORKDIR /app
COPY --from=builder /app/chat-app .
# RUN chmod +x chat-app
EXPOSE 8080
CMD ["/app/chat-app"]
