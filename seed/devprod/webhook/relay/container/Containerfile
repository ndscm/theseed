FROM ghcr.io/ndscm/ubuntu:latest

COPY ./webhook-relay /opt/webhook/webhook-relay

COPY --chmod=0755 ./entrypoint.sh /opt/webhook/webhook-relay-entrypoint.sh

RUN useradd --shell /bin/bash --create-home relay
USER relay
WORKDIR /home/relay

ENTRYPOINT [ "/opt/webhook/webhook-relay-entrypoint.sh" ]
