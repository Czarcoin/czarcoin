ARG CGO_ENABLED=1
ARG REPOSITORY=../czarcoin.org/czarcoin
ARG PACKAGE=czarcoin.org/czarcoin/cmd/storagenode
FROM czarcoin/golang as build-env

# final stage
FROM alpine
ENV CONF_PATH=/root/.local/share/czarcoin/storagenode/config.yaml \
    SATELLITE_ADDR=
EXPOSE 7776/udp \
       7777
WORKDIR /app
COPY --from=build-env /app /app/storagenode
COPY cmd/storagenode/entrypoint /entrypoint
ENTRYPOINT ["/entrypoint"]
