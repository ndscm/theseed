FROM ghcr.io/ndscm/ubuntu:latest

RUN apt update
RUN apt install -y build-essential
RUN apt install -y clang
RUN apt install -y clang-format
RUN apt install -y curl
RUN apt install -y git
RUN apt install -y gnupg2
RUN apt install -y iputils-ping
RUN apt install -y jq
RUN apt install -y libcap2-bin
RUN apt install -y lsb-release
RUN apt install -y netcat-openbsd
RUN apt install -y rsync
RUN apt install -y ssh
RUN apt install -y vim
RUN apt install -y zsh

RUN curl --fail --silent https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor > /etc/apt/trusted.gpg.d/docker.gpg
RUN echo "deb https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" >/etc/apt/sources.list.d/docker.list
RUN curl --fail --silent https://cli.github.com/packages/githubcli-archive-keyring.gpg -o /etc/apt/trusted.gpg.d/githubcli-archive-keyring.gpg
RUN echo "deb https://cli.github.com/packages stable main" >/etc/apt/sources.list.d/github-cli.list
RUN curl --fail --silent --show-error --location https://deb.nodesource.com/setup_24.x | bash

RUN apt update
RUN apt install -y docker-ce-cli
RUN apt install -y gh
RUN apt install -y nodejs

RUN corepack enable
RUN npm install --global @bazel/bazelisk

COPY ./ndscm /usr/bin/ndscm
COPY ./amadeus-server /opt/amadeus/amadeus-server

COPY ./gain-home.c /opt/amadeus/gain-home.c
RUN gcc /opt/amadeus/gain-home.c -o /opt/amadeus/gain-home
RUN setcap cap_chown,cap_fowner+ep /opt/amadeus/gain-home

COPY --chmod=0755 ./entrypoint.sh /opt/amadeus/entrypoint.sh

RUN useradd --shell /usr/bin/zsh --create-home amadeus
USER amadeus
WORKDIR /home/amadeus

ENTRYPOINT [ "/opt/amadeus/entrypoint.sh" ]

CMD ["/opt/amadeus/amadeus-server", "--openid_discovery", "/etc/steins/openid_discovery.json", "--verbose"]
