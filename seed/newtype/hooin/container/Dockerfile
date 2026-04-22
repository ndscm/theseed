FROM ghcr.io/ndscm/ubuntu:latest

# Install debug tools
RUN apt update
RUN apt install -y curl
RUN apt install -y gnupg2
RUN apt install -y lsb-release

COPY ./hooin-server /opt/hooin/hooin-server

COPY --chmod=0755 ./entrypoint.sh /opt/hooin/entrypoint.sh

RUN mkdir -p /etc/hooin
# The default team file is only used for demo purposes.
# In production, you should mount your own team file with proper random tokens.
COPY ./team.json /etc/hooin/team.json

RUN useradd --shell /bin/bash --create-home hooin
USER hooin
WORKDIR /home/hooin

ENTRYPOINT [ "/opt/hooin/entrypoint.sh" ]
CMD ["--static_team_file", "/etc/hooin/team.json", "--verbose"]
