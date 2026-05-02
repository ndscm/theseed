FROM ghcr.io/ndscm/ubuntu:latest AS base

ENV LANG=C.UTF-8

RUN apt update
RUN apt upgrade -y
RUN apt install -y gzip
RUN apt install -y openjdk-25-jdk-headless
RUN apt install -y tar

ADD ./keycloak.tar.gz /opt/keycloak/
COPY ./junixsocket-common.jar /opt/keycloak/providers/junixsocket-common.jar
COPY ./junixsocket-native-common.jar /opt/keycloak/providers/junixsocket-native-common.jar

RUN mkdir -p /opt/keycloak/data

RUN useradd --home-dir /opt/keycloak keycloak
RUN chown -R keycloak:keycloak /opt/keycloak
USER keycloak

ENV KC_RUN_IN_CONTAINER="true"

ENV KC_CACHE="local"
ENV KC_HEALTH_ENABLED="true"
ENV KC_METRICS_ENABLED="true"

ENV KC_HTTP_ENABLED="true"
ENV KC_PROXY_HEADERS="xforwarded"

ENV KC_DB="postgres"

WORKDIR /opt/keycloak
RUN /opt/keycloak/bin/kc.sh build

ENTRYPOINT ["/opt/keycloak/bin/kc.sh"]

CMD ["start", "--optimized"]
