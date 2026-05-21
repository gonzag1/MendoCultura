FROM golang:1.26-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /mendocultura ./cmd

FROM alpine:3.21
WORKDIR /app
COPY --from=build /mendocultura /app/mendocultura
EXPOSE 8080
CMD ["/app/mendocultura"]
