
FROM golang:latest as builder


WORKDIR /app
ENV GOOS=linux
ENV CGO_ENABLED=1
COPY . .
RUN go mod tidy
RUN  go build -C .  -o ./bin


FROM debian:bookworm
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /app
COPY --from=builder app/bin .
EXPOSE 8080


CMD ["./bin"]

