FROM nginx:1.16.0-alpine

RUN apk -U --no-cache add curl ca-certificates\
    && mkdir -p /etc/confd \
    && curl -sLf https://github.com/kelseyhightower/confd/releases/download/v0.16.0/confd-0.16.0-linux-amd64 > /usr/bin/confd \
    && chmod +x /usr/bin/confd \
    && apk del curl

COPY templates /etc/confd/templates/
COPY conf.d /etc/confd/conf.d/
COPY nginx-proxy /usr/bin/

ENTRYPOINT nginx-proxy