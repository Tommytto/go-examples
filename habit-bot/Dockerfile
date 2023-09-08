FROM golang:1.20.2-alpine3.17
WORKDIR /app
COPY . .
RUN go build .

FROM alpine:latest
WORKDIR /app
COPY --from=0 /app ./
CMD ["./habit-bot", "--port", "80"]
