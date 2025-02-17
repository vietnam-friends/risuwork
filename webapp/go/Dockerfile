FROM golang:1.22-bookworm AS base

WORKDIR /tmp
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get -y update && \
    apt-get -y upgrade && \
    apt-get install -y wget gcc g++ make locales locales-all && \
    wget -q https://dev.mysql.com/get/mysql-apt-config_0.8.29-1_all.deb && \
    apt-get -y install ./mysql-apt-config_0.8.29-1_all.deb && \
    apt-get -y install default-mysql-client

RUN locale-gen en_US.UTF-8
RUN useradd --uid=1001 --create-home isucon
USER isucon

RUN mkdir -p /home/isucon/webapp/go
RUN mkdir -p /home/isucon/tmp/go/.cache

FROM base AS download

WORKDIR /home/isucon/webapp/go
USER isucon

COPY --chown=isucon:isucon ./go/go.mod ./go/go.sum /home/isucon/webapp/go/
RUN go env -w GOMODCACHE=/home/isucon/tmp/go/.cache
RUN --mount=type=cache,uid=1001,target=/home/isucon/tmp/go/.cache,sharing=locked go mod download -x

FROM download AS dev

COPY --chown=isucon:isucon ./go/ /home/isucon/webapp/go/
COPY --chown=isucon:isucon ./sql/ /home/isucon/webapp/sql/

FROM dev AS build

WORKDIR /home/isucon/webapp/go
USER isucon

RUN go env -w GOMODCACHE=/home/isucon/tmp/go/.cache
RUN go env -w GOCACHE=/home/isucon/tmp/go/.cache
RUN --mount=type=cache,uid=1001,target=/home/isucon/tmp/go/.cache,sharing=locked CGO_ENABLED=0 go build -o risuwork .

ENV GOPATH=/home/isucon/tmp/go

ENV LANG en_US.UTF-8
ENV LANGUAGE en_US:en
ENV LC_ALL en_US.UTF-8

ENV TZ utc

EXPOSE 8080
CMD ["/home/isucon/webapp/go/risuwork"]
