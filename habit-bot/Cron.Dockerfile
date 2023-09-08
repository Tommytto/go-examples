FROM golang:1.20.2-alpine3.17
WORKDIR /app
COPY . .
RUN go build internal/ask_habit_confirm/main.go

FROM alpine:latest
WORKDIR /app
# utc
RUN echo "0 5 * * * cd /app && ./main" >> /var/spool/cron/crontabs/root

COPY --from=0 /app ./
CMD crond -f -l 2 -L /dev/stdout
