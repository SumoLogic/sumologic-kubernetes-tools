FROM golang:1.15.2 as go-builder
RUN mkdir /build
ADD ./src/go /build/
WORKDIR /build
# Following flags are passed to ld (external linker):
# * -w to decrease binary size by not including debug info
# * -extldflags "-static" to build static binaries
RUN CGO_ENABLED=0 GOOS=linux \
    go build -a \
        -ldflags '-w -extldflags "-static"' \
        -o k8s-api-test cmd/k8s-api-test/main.go
RUN CGO_ENABLED=0 GOOS=linux \
    go build -a \
        -ldflags '-w -extldflags "-static"' \
        -o stress-tester cmd/stress-tester/main.go

FROM rust:alpine3.11 as rust-builder
COPY ./src/rust/receiver-mock /build
WORKDIR /build
RUN apk update && apk upgrade && apk add g++
RUN RUSTFLAGS="-C target-feature=-crt-static" cargo build --release
RUN find / -name libgcc_s.so.1

FROM alpine:3.9
RUN set -ex \
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

ENV HELM_VERSION="v3.2.4"
RUN wget -q https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz -O - | tar -xzO linux-amd64/helm > /usr/local/bin/helm \
    && chmod +x /usr/local/bin/helm

ADD ./src/ssh/motd /etc/motd
ADD ./src/ssh/profile  /etc/profile

ADD ./src/commands/check /usr/bin/check
ADD ./src/commands/fix-log-symlinks /usr/bin/fix-log-symlinks
ADD ./src/commands/tools-usage /usr/bin/tools-usage
ADD ./src/commands/template /usr/bin/template
ADD ./src/commands/template-dependency /usr/bin/template-dependency

RUN curl -LJ https://github.com/mikefarah/yq/releases/download/3.2.1/yq_linux_amd64 -o /usr/bin/yq
RUN chmod +x /usr/bin/yq

COPY --from=go-builder /build/k8s-api-test /usr/bin/
COPY --from=go-builder /build/stress-tester /usr/bin/
COPY --from=rust-builder /build/target/release/receiver-mock /usr/bin
COPY --from=rust-builder /usr/lib/libgcc_s.so.1 /usr/lib

CMD ["/usr/bin/tools-usage"]
