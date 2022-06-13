FROM starport/cli:0.19.1 AS starport-cli

#FROM node:16-alpine3.15
#
#WORKDIR /node/vue
#
#COPY vue/package.json .

# TODO: reimplement vue ui
#RUN echo "prepping realio-network vue builder" ; \
#    apk update && \
#    apk add --no-cache --virtual .gyp python2 make g++ && \
#    npm install && \
#    apk del .gyp && \
#    echo "done"

#COPY vue/ .

#RUN npm run build

FROM golang:1.17-buster AS secret-service-builder

ARG APTOPT
WORKDIR /go/src

RUN echo "building secret-service" ; \
    #mkdir -p github.com/yousefvand && \
    #git clone https://github.com/kainz/secret-service --depth 1 -b fix_collections_pathtypes github.com/yousefvand/secret-service && \
    #cd github.com/yousefvand/secret-service && \
    #go build -race --mod=readonly -o /go/bin/secretserviced ./cmd/app/secretserviced && \
    #go build -race --mod=readonly -o /go/bin/secretservice ./cmd/app/secretservice && \
    go install -race github.com/yousefvand/secret-service/cmd/app/secretserviced@v0.2.1 && \
    go install -race github.com/yousefvand/secret-service/cmd/app/secretservice@v0.2.1 && \
    echo done

FROM golang:1.17-buster AS realio-network-builder

# use something like --build-arg APTOPT='-o Acquire::HTTP::Proxy=http://proxy:3142' to use an apt-cacher for example
ARG APTOPT

WORKDIR /go/src/github.com/realiotech/realio-network

RUN echo "prepping realio-network builder" ; \
  apt-get $APTOPT update && \
  apt-get $APTOPT install -y --no-install-recommends --no-install-suggests \
  time && \
  rm -rf /var/lib/apt/lists/* && \
  echo "done"

## only copy go.mod and go.sum
COPY go.mod go.mod
COPY go.sum go.sum

RUN time go mod download

COPY . .

RUN echo "doing build" ; \
    go build -v --mod=readonly -o /go/bin/realio-networkd ./cmd/realio-networkd

FROM debian:buster

# use something like --build-arg APTOPT='-o Acquire::HTTP::Proxy=http://proxy:3142' to use an apt-cacher for example
ARG APTOPT
ENV DBUS_SESSION_BUS_ADDRESS=unix:path=/tmp/dbus-docker-session

RUN useradd -ms /bin/bash -d /data realio-cosmos && \
   apt-get $APTOPT update && \
   apt-get $APTOPT -y install --no-install-recommends --no-install-suggests \
     dbus \
     libsecret-tools \
     procps \
     s6 \
     && \
   apt-get $APTOPT clean && \
   rm -rf /var/lib/apt/lists && \
   echo done

COPY --from=realio-network-builder /go/bin/realio-networkd /usr/local/bin/
COPY --from=secret-service-builder /go/bin/secretservice* /usr/local/bin/
COPY genesis-testnet.json /usr/local/share/doc/realio-networkd/genesis-testnet.json
COPY s6 /usr/local/etc/s6/

VOLUME /data

run chown realio-cosmos:realio-cosmos /data

EXPOSE 9090 9091 26656

