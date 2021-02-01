ARG CGO_ENABLED=1
ARG REPOSITORY=../czarcoin.org/czarcoin
ARG PACKAGE=czarcoin.org/czarcoin/cmd/gateway
FROM czarcoin/golang as build-env

# final stage
FROM alpine
ENV CONF_PATH=/root/.local/share/czarcoin/uplink/config.yaml \
    API_KEY= \
    SATELLITE_ADDR=
EXPOSE 7776/udp \
       7777
WORKDIR /app
COPY --from=build-env /app /app/gateway
COPY cmd/gateway/entrypoint /entrypoint
ENTRYPOINT ["/entrypoint"]
