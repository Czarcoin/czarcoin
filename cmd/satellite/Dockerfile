ARG CGO_ENABLED=1
ARG REPOSITORY=../czarcoin.org/czarcoin
ARG PACKAGE=czarcoin.org/czarcoin/cmd/satellite
FROM czarcoin/golang as build-env

# final stage
FROM alpine
ENV API_KEY= \
    CONF_PATH=/root/.local/share/czarcoin/satellite/config.yaml \
    OVERLAY_URL=redis://redis:6379/?db=0 \
    IDENTITY_ADDR=:7777 \
    BOOTSTRAP_ADDR=bootstrap.czarcoin.org:8080 \
    CZARCOIN_MOCK_OVERLAY_NODES=
EXPOSE 7776/udp \
       7777 \
       8080
WORKDIR /app
COPY --from=build-env /app /app/satellite
COPY cmd/satellite/entrypoint /entrypoint
ENTRYPOINT ["/entrypoint"]
