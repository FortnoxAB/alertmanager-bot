FROM gcr.io/distroless/static-debian12:nonroot
COPY alertmanager-bot /
USER nonroot
ENTRYPOINT ["/alertmanager-bot"]
