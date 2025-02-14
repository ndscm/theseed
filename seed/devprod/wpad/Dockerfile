FROM node:latest

RUN useradd --create-home wpad
USER wpad
WORKDIR /home/wpad

ADD ./wpad.js ./
ADD ./proxy.pac ./

CMD [ "node", "./wpad.js" ]
