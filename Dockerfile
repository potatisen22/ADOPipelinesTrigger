FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY ./ ./

RUN go build -o /bin/app ./src/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /bin/app /bin/app

ENTRYPOINT ["app"]