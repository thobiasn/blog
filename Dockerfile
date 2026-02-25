FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o blog .

FROM alpine:3.21
WORKDIR /app
COPY --from=build /app/blog .
COPY content/ ./content/
COPY templates/ ./templates/
COPY static/ ./static/
EXPOSE 8080
CMD ["./blog", "serve"]
