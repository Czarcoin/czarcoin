ARG CGO_ENABLED=1
ARG REPOSITORY=../czarcoin.org/czarcoin
ARG PACKAGE=czarcoin.org/czarcoin/cmd/uplink
FROM czarcoin/golang as build-env

# final stage
FROM alpine
ENV CONF_PATH=/root/.czarcoin/uplink/config.yaml \
    API_KEY= \
    SATELLITE_ADDR=
EXPOSE 7776/udp \
       7777
WORKDIR /app
COPY --from=build-env /app /app/uplink
COPY cmd/uplink/entrypoint /entrypoint
ENTRYPOINT ["/entrypoint"]
