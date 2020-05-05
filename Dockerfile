
FROM alpine:3.11 as parity_install

ARG REPO=x86_64-unknown-linux-gnu
ARG VERSION=2.7.2

RUN wget https://releases.parity.io/ethereum/v${VERSION}/${REPO}/parity && chmod a+x parity

RUN mv parity /usr/local/bin/parity

################################################################################

FROM ubuntu:20.04 as go_builder

RUN apt-get update && apt-get install -y wget gcc

ARG GO_VERSION=1.14

RUN wget https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
    && rm go${GO_VERSION}.linux-amd64.tar.gz

ENV GOROOT=/usr/local/go
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

WORKDIR /go/src

################################################################################

FROM go_builder as go_build

COPY . /go/src

RUN go generate ./...
 
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go install cmd/app/main.go

################################################################################

FROM go_build as dev_install

ENV DEBIAN_FRONTEND=noninteractive 

RUN apt-get update && apt-get install -y lighttpd

COPY --from=parity_install /usr/local/bin/parity /usr/local/bin/

################################################################################

FROM ubuntu:20.04 as prod_install

ENV DEBIAN_FRONTEND=noninteractive 

RUN apt-get update && apt-get install -y lighttpd jq

RUN useradd -ms /bin/bash eth

RUN mkdir /chain && chown -R eth:eth /chain
RUN mkdir -p /var/www/localhost/htdocs \
    && chown -R eth:eth /var/www/localhost/htdocs \
    && chown -R eth:eth /etc/lighttpd/

USER eth

COPY --from=parity_install /usr/local/bin/parity /usr/local/bin/
COPY --from=go_build /go/bin/main /usr/local/bin/bnc-eth

WORKDIR /home/eth

EXPOSE 8545 8546 30303 5000

ENTRYPOINT ["bnc-eth"]

################################################################################

