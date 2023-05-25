# Собираем бинарный файл
FROM golang:alpine as builder
RUN apk add git
RUN apk add --update gcc musl-dev
WORKDIR /app
ADD ./go.mod /app/go.mod
ADD ./go.sum /app/go.sum
RUN go mod download
ADD . /app
ARG BUILD_FLAGS
RUN go build -ldflags "${BUILD_FLAGS}" -o ./server ./cmd/server
RUN mkdir -p /src
WORKDIR /src
ADD . /src

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/server /bin/server

ENTRYPOINT ["server"]
CMD ["start"]