FROM alpine:3.15

MAINTAINER Lei Jianzhong

LABEL VERSION=v1.0.2 \
      ARCH=AMD64 \
      DESCRIPTION="A redis data clean tool"

COPY  bin/redis-agent /usr/local/bin/redis-agent

# 因为是在contos下编译的，动态链接库的位置是/lib64, 而alpine的链接库是在lib下，所以这里创建一个软连接到/lib64下
RUN  mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

EXPOSE 6389

# /bin/sh -c "/usr/local/bin/redis_agent"
ENTRYPOINT ["/usr/local/bin/redis-agent"]
