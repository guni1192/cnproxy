# cnproxy

cnproxy is Cloud Native forward proxy.


## Features

* http/https supprot (http CONNECT method)
* http healthcheck
* structured logging
* metrics
* tracing (TODO)


## Usage

1. Launch the proxy server.

```shell
./run.sh
```

2. send a request to the proxy server.

```shell
curl -s --proxy http://localhost:1192 https://example.com
```
