FROM frolvlad/alpine-glibc

RUN apk add --no-cache tzdata
RUN cp /usr/share/zoneinfo/Asia/Seoul /etc/localtime
RUN echo "Asia/Seoul" > /etc/timezone
ADD nbgrepd /home/deploy/nbgrepd
WORKDIR /home/deploy
ENTRYPOINT [ "/home/deploy/nbgrepd" ]

