# Gateway

Documentation for developing and building the gateway service

Usage:

First make an identity:
```
go install czarcoin.org/czarcoin/cmd/gateway
gateway setup
```

The gateway shares the uplink config file.
You can edit `~/.czarcoin/uplink/config.yaml` to your liking. Then run it!

```
gateway run
```
