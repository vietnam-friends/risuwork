FROM openjdk:21-jdk-slim

WORKDIR /tmp
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && \
    apt-get -y upgrade && \
    apt-get install -y maven && \
    apt-get install -y curl wget gcc g++ make sqlite3 locales locales-all && \
    wget -q https://dev.mysql.com/get/mysql-apt-config_0.8.29-1_all.deb && \
    apt-get -y install ./mysql-apt-config_0.8.29-1_all.deb && \
    apt-get -y update && \
    apt-get -y install default-mysql-client pdns-server pdns-backend-mysql

RUN rm -f /etc/powerdns/pdns.d/bind.conf

RUN locale-gen en_US.UTF-8
RUN useradd --uid=1001 --create-home isucon
USER isucon

RUN mkdir -p /home/isucon/webapp/java
WORKDIR /home/isucon/webapp/java
COPY --chown=isucon:isucon ./java /home/isucon/webapp/java/
COPY --chown=isucon:isucon ./sql /home/isucon/webapp/sql/
RUN mvn clean package -f /home/isucon/webapp/java
ENV LANG=en_US.UTF-8
ENV LANGUAGE=en_US:en
ENV LC_ALL=en_US.UTF-8

ENV TZ=utc

EXPOSE 8080
USER isucon
CMD ["java", "-jar", "/home/isucon/webapp/java/target/isucon2024.jar"]
