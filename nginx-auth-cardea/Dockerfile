FROM busybox:musl
MAINTAINER Maciej Pasternacki <maciej@3ofcoins.net>

ADD nginx-auth-cardea /nginx-auth-cardea
USER 999
EXPOSE 8080
ENTRYPOINT [ "/nginx-auth-cardea" ]
