FROM golang:1.26.3-alpine3.23 AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -o out/binary

FROM alpine:3.23.4

WORKDIR /app

RUN adduser -D appuser \
	&& chown appuser:appuser /app

COPY --from=builder /app/out/binary /usr/local/bin/

USER appuser

ENTRYPOINT [ "/usr/local/bin/binary" ]
