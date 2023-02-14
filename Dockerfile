FROM golang:1.18 AS builder

COPY . /src
WORKDIR /src

RUN GOPROXY=https://goproxy.cn go build -v -o bin/server

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/bin /app

WORKDIR /app

EXPOSE 7000
EXPOSE 7001
VOLUME /app/conf

CMD ["./server", "-C", "./conf/config.toml"]
