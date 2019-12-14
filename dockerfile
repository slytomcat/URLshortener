FROM alpine:latest
WORKDIR /opt/URLshortener
COPY URLshortener .
CMD ["./URLshortener"]
