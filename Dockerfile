FROM alpine:latest

RUN apk add --verbose --no-cache bash
RUN apk add --verbose --no-cache curl
RUN apk add --verbose --no-cache docker-cli
RUN apk add --verbose --no-cache git
RUN apk add --verbose --no-cache mercurial
RUN apk add --verbose --no-cache make
RUN apk add --verbose --no-cache build-base

ENTRYPOINT ["/entrypoint.sh"]
CMD [ "-h" ]

COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

COPY glab_*.apk /tmp/
RUN apk add --allow-untrusted /tmp/glab_*.apk