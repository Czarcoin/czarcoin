FROM golang:1.10

RUN apt-get update \
 && apt-get install -y build-essential

RUN cd / \
 && rm -rf /go \
 && git clone --recursive https://github.com/czarcoin/czarcoin-vendor.git /go \
 && cd /go \
 && ./setup.sh

COPY . /go/src/czarcoin.org/czarcoin

WORKDIR /go/src/czarcoin.org/czarcoin

RUN make build-dev-deps
