FROM golang:1.26-alpine AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux

ARG TARGETARCH
ENV GOARCH=$TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o /go-api main.go

#---------------------------------------------
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S appgroup && \
    adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /go-api /go-api

USER appuser:appgroup


EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/status || exit 1

CMD ["/go-api"]