FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN go build -o blog ./cmd/blog

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
ADD https://github.com/benbjohnson/litestream/releases/download/v0.3.13/litestream-v0.3.13-linux-amd64.tar.gz /tmp/litestream.tar.gz
RUN tar -xzf /tmp/litestream.tar.gz -C /usr/local/bin && rm /tmp/litestream.tar.gz
RUN adduser -D -h /app blog
WORKDIR /app
COPY --from=build /app/blog .
COPY content/ ./content/
COPY templates/ ./templates/
COPY static/ ./static/
COPY litestream.yml /etc/litestream.yml
COPY entrypoint.sh .
RUN chown -R blog:blog /app
USER blog
EXPOSE 8080
ENTRYPOINT ["./entrypoint.sh"]
