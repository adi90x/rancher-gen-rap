FROM alpine:edge
MAINTAINER <amaurel90@gmail.com>

RUN apk add --no-cache ca-certificates
RUN apk add --no-cache nano
RUN apk add --no-cache unzip
ENV RANCHER_GEN_RELEASE v0.2.0

RUN wget -O /tmp/rancher-gen.zip  https://gitlab.com/adi90x/go-rancher-gen/builds/artifacts/master/download?job=compile-go
RUN unzip /tmp/rancher-gen.zip -d /usr/local/bin \
	&& chmod +x /usr/local/bin/rancher-gen

ENTRYPOINT ["/usr/local/bin/rancher-gen"]
