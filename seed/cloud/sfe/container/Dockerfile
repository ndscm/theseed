FROM ghcr.io/ndscm/ubuntu:latest

ADD ./sfe-server /opt/sfe/

RUN useradd --shell /bin/bash --create-home sfe
USER sfe
WORKDIR /home/sfe

ENTRYPOINT ["/opt/sfe/sfe-server"]

CMD ["--verbose"]
