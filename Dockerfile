
FROM alpine:3.11 as parity_install

# OpenEthereum has downgrade version
#ARG DIST=linux
#ARG VERSION=3.1.0
#RUN wget https://github.com/openethereum/openethereum/releases/download/v${VERSION}/openethereum-${DIST}-v${VERSION}.zip \
#   && unzip openethereum-${DIST}-v${VERSION}.zip -d /usr/local/bin/ \
#   && chmod a+x /usr/local/bin/openethereum

ARG DIST=x86_64-unknown-linux-gnu
ARG VERSION=2.7.2
RUN wget https://releases.parity.io/ethereum/v${VERSION}/${DIST}/parity \
    && chmod a+x parity \
    && mv parity /usr/local/bin/parity

################################################################################

FROM ubuntu:20.04 as go_builder

RUN apt-get update && apt-get install -y wget gcc

ARG GO_VERSION=1.15

RUN wget https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
    && rm go${GO_VERSION}.linux-amd64.tar.gz

ENV GOROOT=/usr/local/go
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

WORKDIR /go/src

################################################################################

FROM alpine:3.11 as solc_install

ARG SOLC_VERSION=0.8.0

RUN wget https://github.com/ethereum/solidity/releases/download/v${SOLC_VERSION}/solc-static-linux \
    && mv solc-static-linux /usr/local/bin/solc \
    && chmod a+x /usr/local/bin/solc

################################################################################

FROM go_builder as go_build

COPY . /go/src

RUN go generate ./...
 
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go install cmd/app/main.go

################################################################################

FROM go_build as dev_install

ENV DEBIAN_FRONTEND=noninteractive 

RUN apt-get update && apt-get install -y lighttpd

COPY --from=solc_install /usr/local/bin/solc /usr/local/bin/
COPY --from=parity_install /usr/local/bin/parity /usr/local/bin/

################################################################################

FROM ubuntu:20.04 as prod_install_base

ENV DEBIAN_FRONTEND=noninteractive 

RUN apt-get update && apt-get install -y lighttpd jq

RUN useradd -ms /bin/bash eth

RUN mkdir /chain && chown -R eth:eth /chain
RUN mkdir -p /var/www/localhost/htdocs \
    && chown -R eth:eth /var/www/localhost/htdocs \
    && chown -R eth:eth /etc/lighttpd/

################################################################################

FROM prod_install_base as prod_install

USER eth

COPY --from=solc_install /usr/local/bin/solc /usr/local/bin/
COPY --from=parity_install /usr/local/bin/parity /usr/local/bin/

COPY --from=go_build /go/bin/main /usr/local/bin/bnc-eth

WORKDIR /home/eth

EXPOSE 8545 8546 30303 5000

ENTRYPOINT ["bnc-eth"]

################################################################################

