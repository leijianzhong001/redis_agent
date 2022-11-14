FROM alpine:3.15 as builder

## https://github.com/leijianzhong001/redis_agent/releases/download/v1.0.0/redis_agent-v1.0.0.linux-amd64.tar.gz
ARG EXPORTER_URL="https://github.com/leijianzhong001/redis_agent/releases/download"

ARG REDIS_EXPORTER_VERSION="1.0.0"

RUN  apk add --no-cache curl ca-certificates && \
      curl -fL -Lo /tmp/redis_agent-v${REDIS_EXPORTER_VERSION}.linux-amd64.tar.gz \
      ${EXPORTER_URL}/v${REDIS_EXPORTER_VERSION}/redis_agent-v${REDIS_EXPORTER_VERSION}.linux-amd64.tar.gz && \
      cd /tmp && tar -xvzf redis_agent-v${REDIS_EXPORTER_VERSION}.linux-amd64.tar.gz

FROM scratch

MAINTAINER Lei Jianzhong

LABEL VERSION=1.0.0 \
      ARCH=AMD64 \
      DESCRIPTION="A redis data clean tool"

COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /tmp/redis_agent /usr/local/bin/redis_agent

EXPOSE 6389

ENTRYPOINT ["/usr/local/bin/redis_agent"]
