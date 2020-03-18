FROM scratch
WORKDIR /opt/URLshortener
COPY URLshortener .
CMD ["./URLshortener"]
