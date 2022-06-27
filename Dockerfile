FROM gcr.io/distroless/static-debian11:nonroot
COPY alertmanager-bot /
USER nonroot
ENTRYPOINT ["/alertmanager-bot"]
