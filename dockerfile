FROM alpine
MAINTAINER slytomcat <slytomcat@mail.ru>
WORKDIR /opt/URLshortener
COPY URLshortener .
CMD ["./URLshortener"]

