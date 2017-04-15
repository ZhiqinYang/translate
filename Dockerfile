FROM alpine:latest
COPY tranlate /usr/bin/
ENTRYPOINT /usr/bin/auth
EXPOSE 10000 8080 6060
