FROM golang:1.22.3-alpine AS build
WORKDIR /app
COPY cmd /app/cmd
COPY internal /app/internal
COPY vendor /app/vendor
COPY go.mod /app/go.mod
COPY go.sum /app/go.sum
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -x -o server /app/cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/server .
COPY .env /app/.env
EXPOSE 8080
CMD ["/app/server"]