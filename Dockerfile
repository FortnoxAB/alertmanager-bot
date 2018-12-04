FROM scratch
COPY alertmanager-bot /
ENTRYPOINT ["/alertmanager-bot"]
