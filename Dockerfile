FROM alpine:3.8
RUN apk add --no-cache ca-certificates
COPY alertmanager-bot /
ENTRYPOINT ["/alertmanager-bot"]
