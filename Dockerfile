FROM alpine:latest
COPY translate /usr/bin/
ENTRYPOINT /usr/bin/translate
EXPOSE 10000 6060
