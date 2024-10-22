#build stage
FROM golang:alpine AS builder
ARG GIT_USERNAME
ARG GIT_PASSWORD 
RUN apk add --no-cache git

RUN echo -e "machine node71.otclick.ru\nlogin ${GIT_USERNAME}\npassword ${GIT_PASSWORD}" > ~/.netrc
RUN chmod 600 ~/.netrc

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o myapp ./cmd/app

#final stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/myapp /app

EXPOSE 8080
CMD ["/app/myapp"]
