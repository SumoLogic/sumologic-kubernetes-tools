FROM golang:1.14.1 as builder
RUN mkdir /build
ADD ./src/go /build/
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o k8s-api-test cmd/k8s-api-test/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o stress-tester cmd/stress-tester/main.go

FROM rust:alpine3.11 as rust-builder
COPY ./src/rust/receiver-mock /build
WORKDIR /build
RUN apk update && apk upgrade && apk add g++
RUN RUSTFLAGS="-C target-feature=-crt-static" cargo build --release
RUN find / -name libgcc_s.so.1

FROM alpine:3.9
RUN set -ex \
    && echo "http://dl-cdn.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories \
    && echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories \
    && apk update \
    && apk upgrade \
    && apk add --no-cache \
    bash \
    busybox-extras \
    curl \
    libc6-compat \
    openssl \
    net-tools \
    vim \
    wget

ENV HELM_VERSION="v3.2.0"
RUN wget -q https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz -O - | tar -xzO linux-amd64/helm > /usr/local/bin/helm \
    && chmod +x /usr/local/bin/helm

ADD ./src/ssh/motd /etc/motd
ADD ./src/ssh/profile  /etc/profile

ADD ./src/commands/check /usr/bin/check
ADD ./src/commands/tools-usage /usr/bin/tools-usage
ADD ./src/commands/template /usr/bin/template

COPY --from=builder /build/k8s-api-test /usr/bin/
COPY --from=builder /build/stress-tester /usr/bin/
COPY --from=rust-builder /build/target/release/receiver-mock /usr/bin
COPY --from=rust-builder /usr/lib/libgcc_s.so.1 /usr/lib

CMD ["/usr/bin/tools-usage"]