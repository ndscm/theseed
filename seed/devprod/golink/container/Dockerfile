FROM ghcr.io/ndscm/ubuntu:latest

ADD ./golink-server.tar.gz /opt/golink/

RUN useradd --shell /bin/bash --create-home golink
USER golink
WORKDIR /home/golink

ENTRYPOINT [ "/opt/golink/golink-server", "--webapp", "/opt/golink/webapp" ]

CMD ["--verbose"]
