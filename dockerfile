FROM alpine
WORKDIR /opt/URLshortener
COPY URLshortener .
CMD ["./URLshortener"]
