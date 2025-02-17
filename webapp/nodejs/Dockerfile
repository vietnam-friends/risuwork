# syntax=docker/dockerfile:1

FROM node:20.7.0-bookworm-slim AS base

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

RUN mkdir -p /home/isucon/webapp/nodejs

FROM base AS download

WORKDIR /home/isucon/webapp/nodejs
USER isucon

COPY --chown=isucon:isucon ./nodejs/ /home/isucon/webapp/nodejs/

RUN --mount=type=cache,uid=1001,target=/root/.npm,sharing=locked \
    npm ci

FROM download AS dev

COPY --chown=isucon:isucon ./sql/ /home/isucon/webapp/sql/

FROM dev AS build

WORKDIR /home/isucon/webapp/nodejs
USER isucon

RUN --mount=type=cache,uid=1001,target=/root/.npm,sharing=locked \
    npm run build

ENV LANG en_US.UTF-8
ENV LANGUAGE en_US:en
ENV LC_ALL en_US.UTF-8
ENV TZ utc

EXPOSE 8080

CMD ["node", "dist/index.js"]