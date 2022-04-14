FROM golang:1.18.1 as go-builder
RUN mkdir /build
ADD ./src/go /build/
WORKDIR /build
# Following flags are passed to ld (external linker):
# * -w to decrease binary size by not including debug info
# * -extldflags "-static" to build static binaries
RUN CGO_ENABLED=0 GOOS=linux \
    go build \
        -ldflags '-w -extldflags "-static"' \
        -o k8s-api-test cmd/k8s-api-test/main.go
RUN CGO_ENABLED=0 GOOS=linux \
    go build \
        -ldflags '-w -extldflags "-static"' \
        -o stress-tester cmd/stress-tester/main.go
RUN CGO_ENABLED=0 GOOS=linux \
    go build \
        -ldflags '-w -extldflags "-static"' \
        -o customer-trace-tester cmd/customer-trace-tester/main.go

FROM rust:1.57.0-alpine3.13 as rust-builder
RUN apk update && apk upgrade && apk add g++

WORKDIR /receiver-mock
COPY ./src/rust/receiver-mock .
RUN cargo build --release

WORKDIR /logs-generator
COPY ./src/rust/logs-generator .
RUN cargo build --release

FROM alpine:3.15.4
ARG TARGETARCH
ARG TARGETOS
ENV HELM_VERSION="3.7.2"
ENV YQ_VERSION="3.4.1"
ENV KUBECTL_VERSION="v1.22.4"
ENV UPGRADE_2_0_SCRIPT_URL="https://raw.githubusercontent.com/SumoLogic/sumologic-kubernetes-collection/release-v2.0/deploy/helm/sumologic/upgrade-2.0.0.sh"
RUN set -ex \
    && apk update \
    && apk upgrade \
    && apk add --no-cache \
        bash \
        busybox-extras \
        coreutils \
        curl \
        libc6-compat \
        openssl \
        net-tools \
        vim \
        jq \
    && curl https://get.helm.sh/helm-v${HELM_VERSION}-${TARGETOS}-${TARGETARCH}.tar.gz | tar -xzO ${TARGETOS}-${TARGETARCH}/helm > /usr/local/bin/helm \
    && chmod +x /usr/local/bin/helm \
    && curl -LJ https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_${TARGETOS}_${TARGETARCH} -o /usr/bin/yq \
    && chmod +x /usr/bin/yq \
    && curl -LJ https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/${TARGETOS}/${TARGETARCH}/kubectl -o /usr/bin/kubectl \
    && chmod +x /usr/bin/kubectl \
    && curl -LJ "${UPGRADE_2_0_SCRIPT_URL}" -o /usr/local/bin/upgrade-2.0.0.sh \
    && chmod +x /usr/local/bin/upgrade-2.0.0.sh \
    && curl -LJ https://raw.githubusercontent.com/dwyl/english-words/master/words.txt -o /usr/local/wordlist.txt

COPY \
    ./src/ssh/motd \
    ./src/ssh/profile \
    /etc/

COPY \
    ./src/commands/check \
    ./src/commands/pvc-cleaner \
    ./src/commands/fix-log-symlinks \
    ./src/commands/tools-usage \
    ./src/commands/template \
    ./src/commands/template-dependency \
    ./src/commands/upgrade-2.0 \
    /usr/bin/

COPY ./src/commands/template-prometheus-mixin \
    /usr/local/template-prometheus-mixin
RUN ln -s /usr/local/template-prometheus-mixin/template-prometheus-mixin /usr/bin

COPY --from=go-builder \
    /build/k8s-api-test \
    /build/stress-tester \
    /build/customer-trace-tester \
    /usr/bin/

COPY --from=rust-builder \
    /receiver-mock/target/release/receiver-mock \
    /logs-generator/target/release/logs-generator \
    /usr/bin/

CMD ["/usr/bin/tools-usage"]
