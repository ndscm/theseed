FROM ghcr.io/ndscm/ubuntu:latest

ADD ./stuff-server.tar.gz /opt/stuff/

RUN useradd --shell /bin/bash --create-home stuff
USER stuff
WORKDIR /home/stuff

ENTRYPOINT [ "/opt/stuff/stuff-server", "--webapp", "/opt/stuff/webapp" ]

CMD ["--verbose"]
